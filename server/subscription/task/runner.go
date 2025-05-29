package task

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	// "path/filepath" // No longer directly used for archive.txt
	"regexp"
	"time"
	"encoding/json" // For parsing yt-dlp JSON output

	// "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive" // No longer directly used for DownloadExists
	archiveDomain "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
	"github.com/robfig/cron/v3"
	// commonTypes "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/common" // For YtdlpVideoInfo if used directly
)

// The placeholder SubscriptionUpdateService interface is no longer needed,
// as domain.Service now includes these methods.
// type SubscriptionUpdateService interface {
// 	CreateSubscriptionUpdate(ctx context.Context, update *domain.SubscriptionVideoUpdate) error
// }

type TaskRunner interface {
	Submit(subcription *domain.Subscription) error
	Spawner(ctx context.Context)
	StopTask(id string) error
	Recoverer()
}

type monitorTask struct {
	Done         chan struct{}
	Schedule     cron.Schedule
	Subscription *domain.Subscription
}

type CronTaskRunner struct {
	mq             *internal.MessageQueue
	db             *internal.MemoryDB // This was for queuing downloads, may not be needed by fetcher directly anymore for that.
	archiveRepo    archiveDomain.Repository    // New field
	subUpdateService domain.Service // Changed from placeholder SubscriptionUpdateService to domain.Service

	tasks  chan monitorTask
	errors chan error
	running map[string]*monitorTask
}

func NewCronTaskRunner(
	mq *internal.MessageQueue,
	db *internal.MemoryDB,
	archiveRepo archiveDomain.Repository, // New param
	subUpdateService domain.Service, // Changed from placeholder SubscriptionUpdateService to domain.Service
) TaskRunner {
	return &CronTaskRunner{
		mq:              mq,
		db:              db,
		archiveRepo:     archiveRepo,    // Store injected repo
		subUpdateService: subUpdateService, // Store injected service
		tasks:           make(chan monitorTask),
		errors:          make(chan error),
		running:         make(map[string]*monitorTask),
	}
}

var argsSplitterRe = regexp.MustCompile(`(?mi)[^\s"']+|"([^"]*)"|'([^']*)'`)

func (t *CronTaskRunner) Submit(subcription *domain.Subscription) error {
	schedule, err := cron.ParseStandard(subcription.CronExpr)
	if err != nil {
		return err
	}

	job := monitorTask{
		Done:         make(chan struct{}),
		Schedule:     schedule,
		Subscription: subcription,
	}

	t.tasks <- job

	return nil
}

// Handles the entire lifecylce of a monitor job.
func (t *CronTaskRunner) Spawner(ctx context.Context) {
	for req := range t.tasks {
		t.running[req.Subscription.Id] = &req // keep track of the current job

		go func() {
			ctx, cancel := context.WithCancel(ctx) // inject into the job's context a cancellation singal
			fetcherEvents := t.doFetch(ctx, &req)  // retrieve the channel of events of the job

			for {
				select {
				case <-req.Done:
					slog.Info("stopping cron job and removing schedule", slog.String("url", req.Subscription.URL))
					cancel()
					return
				case <-fetcherEvents:
					slog.Info("finished monitoring channel", slog.String("url", req.Subscription.URL))
				}
			}
		}()
	}
}

// Stop a currently scheduled job
func (t *CronTaskRunner) StopTask(id string) error {
	task := t.running[id]
	if task != nil {
		t.running[id].Done <- struct{}{}
		delete(t.running, id)
	}
	return nil
}

// Start a fetcher and notify on a channel when a fetcher has completed
func (t *CronTaskRunner) doFetch(ctx context.Context, req *monitorTask) <-chan struct{} {
	completed := make(chan struct{})

	// generator func
	go func() {
		for {
			sleepFor := t.fetcher(ctx, req)
			completed <- struct{}{}

			time.Sleep(sleepFor)
		}
	}()

	return completed
}

// Perform the retrieval of the latest video of the channel.
// Returns a time.Duration containing the amount of time to the next schedule.
func (t *CronTaskRunner) fetcher(ctx context.Context, req *monitorTask) time.Duration {
	slog.Info("fetching latest video for channel", slog.String("channel", req.Subscription.URL))

	nextSchedule := time.Until(req.Schedule.Next(time.Now()))

	cmd := exec.CommandContext(
		ctx,
		config.Instance().DownloaderPath,
		req.Subscription.URL,
		"--dump-single-json",
		"--flat-playlist",
		"-I0:10", // Get info for the first 10 videos
		"--no-warnings",
	)

	output, err := cmd.Output()
	if err != nil {
		slog.Error("yt-dlp command failed for subscription", "url", req.Subscription.URL, "error", err)
		if ctxErr := ctx.Err(); ctxErr != nil { // Check if context was cancelled
			slog.Info("Context cancelled during yt-dlp command for subscription", "url", req.Subscription.URL, "error", ctxErr)
			return time.Duration(0) // Stop processing if context is done
		}
		t.errors <- err // Send other errors to the error channel
		return nextSchedule // Return nextSchedule to attempt again later
	}

	var channelDump domain.YtdlpChannelDump
	if err := json.Unmarshal(output, &channelDump); err != nil {
		slog.Error("Failed to unmarshal yt-dlp JSON output for subscription", "url", req.Subscription.URL, "error", err)
		t.errors <- err
		return nextSchedule // Return nextSchedule
	}

	for _, videoEntry := range channelDump.Entries {
		if videoEntry.WebpageURL == "" {
			slog.Warn("Skipping video entry with empty WebpageURL", "subscriptionId", req.Subscription.Id, "title", videoEntry.Title)
			continue
		}

		exists, checkErr := t.archiveRepo.IsSourceDownloaded(ctx, videoEntry.WebpageURL)
		if checkErr != nil {
			slog.Error("Failed to check if video source is downloaded", "videoURL", videoEntry.WebpageURL, "error", checkErr)
			// Potentially send to t.errors or just log and continue
			continue
		}

		if !exists {
			newVideoUpdate := &domain.SubscriptionVideoUpdate{
				SubscriptionID: req.Subscription.Id,
				VideoURL:       videoEntry.WebpageURL,
				VideoTitle:     videoEntry.Title,
				ThumbnailURL:   videoEntry.Thumbnail,
				// PublishedAt will be set below
			}

			if videoEntry.UploadDate != "" {
				parsedTime, parseErr := time.Parse("20060102", videoEntry.UploadDate)
				if parseErr == nil {
					newVideoUpdate.PublishedAt = parsedTime
				} else {
					slog.Warn("Failed to parse upload_date for video update", "date", videoEntry.UploadDate, "videoTitle", videoEntry.Title, "error", parseErr)
					// Optionally, set PublishedAt to time.Now() or a zero value if parsing fails,
					// or leave it as the zero value of time.Time
				}
			}

			if createErr := t.subUpdateService.CreateSubscriptionUpdate(ctx, newVideoUpdate); createErr != nil {
				slog.Error("Failed to create subscription video update", "videoURL", newVideoUpdate.VideoURL, "subscriptionId", req.Subscription.Id, "error", createErr)
				// Potentially send to t.errors
			} else {
				slog.Info("Detected new video for subscription, update created", "channel", req.Subscription.URL, "videoTitle", newVideoUpdate.VideoTitle)
			}
		}
	}

	slog.Info(
		"Subscription fetcher finished, next schedule",
		slog.String("url", req.Subscription.URL),
		slog.Any("duration", nextSchedule),
	)

	return nextSchedule
}

func (t *CronTaskRunner) Recoverer() {
	panic("unimplemented")
}
