// server/subscription/service/service.go
package service

import (
	"bytes"
	"context"
	"database/sql" // Required for sql.ErrNoRows
	"encoding/json"
	"errors" // Keep if used by other stubs or for errors.Is
	"fmt"
	"log/slog" // For logging
	"os/exec"

	"github.com/google/uuid"                                                      // For temporary ID generation in Submit (used in stub)
	archiveDomain "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain" 
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/data" // For data.Subscription
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/task" // Added task import
)

type service struct {
	repo          domain.Repository // This is subscriptionDomain.Repository
	archiveRepo   archiveDomain.Repository
	runner        task.TaskRunner // Added
	// Add other dependencies if needed, like a logger
}

// NewService creates a new subscription service.
func NewService(repo domain.Repository, archiveRepo archiveDomain.Repository, runner task.TaskRunner) domain.Service { // Added runner
	return &service{
		repo:        repo,
		archiveRepo: archiveRepo,
		runner:      runner, // Added
	}
}

// GetChannelVideos implements domain.Service.
func (s *service) GetChannelVideos(ctx context.Context, subscriptionID string) (*domain.YtdlpChannelDump, error) {
	slog.Info("Fetching subscription details", "subscriptionID", subscriptionID)
	subData, err := s.repo.Get(ctx, subscriptionID) // Use the new Get method

	if err != nil {
		// Check if it's a 'not found' error specifically if your repo.Get returns sql.ErrNoRows
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("Subscription not found (sql.ErrNoRows)", "subscriptionID", subscriptionID, "error", err)
			return nil, fmt.Errorf("subscription with ID %s not found: %w", subscriptionID, err)
		}
		slog.Error("Failed to get subscription from repository", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("failed to get subscription %s: %w", subscriptionID, err)
	}
	if subData == nil { // If repo.Get returns (nil, nil) for not found (as implemented)
		slog.Warn("Subscription not found (repo returned nil data and nil error)", "subscriptionID", subscriptionID)
		// This specific error message helps distinguish from the sql.ErrNoRows case if both are possible.
		return nil, fmt.Errorf("subscription with ID %s not found (data was nil)", subscriptionID)
	}

	channelURL := subData.URL
	slog.Info("Subscription found, proceeding to fetch channel videos", "subscriptionID", subscriptionID, "channelURL", channelURL)

	cmd := exec.CommandContext(ctx, config.Instance().DownloaderPath,
		channelURL,
		"--dump-single-json",
		"--flat-playlist",
		"--no-warnings",
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	slog.Info("Executing yt-dlp", "command", cmd.String())

	runErr := cmd.Run() // Changed variable name to avoid conflict with 'err' from repo.Get
	if runErr != nil {
		slog.Error("yt-dlp command failed", "error", runErr, "stderr", stderr.String())
		return nil, fmt.Errorf("yt-dlp command failed: %w; Stderr: %s", runErr, stderr.String())
	}

	var channelDump domain.YtdlpChannelDump
	if unmarshalErr := json.Unmarshal(out.Bytes(), &channelDump); unmarshalErr != nil { // Changed variable name
		slog.Error("Failed to unmarshal yt-dlp JSON output", "error", unmarshalErr, "outputSize", len(out.Bytes()))
		return nil, fmt.Errorf("failed to unmarshal yt-dlp JSON output: %w", unmarshalErr)
	}

	if channelDump.OriginalURL == "" {
		channelDump.OriginalURL = channelURL
	}
	if channelDump.ID == "" { 
		channelDump.ID = subData.Id 
	}

	// Check download status for each video
	for i := range channelDump.Entries {
		video := &channelDump.Entries[i] // Use pointer to modify in place
		if video.WebpageURL != "" { // Ensure there's a URL to check
			isDownloaded, checkErr := s.archiveRepo.IsSourceDownloaded(ctx, video.WebpageURL)
			if checkErr != nil {
				slog.Error("Failed to check if video is downloaded from archive", "videoURL", video.WebpageURL, "error", checkErr)
				video.IsDownloaded = false // Default to false on error
			} else {
				video.IsDownloaded = isDownloaded
			}
		} else {
			video.IsDownloaded = false // No URL to check
		}
	}

	slog.Info("Successfully fetched and parsed channel videos, and checked download status", "channelTitle", channelDump.Title, "entryCount", len(channelDump.Entries))
	return &channelDump, nil
}

// --- Stub implementations for other domain.Service methods (as provided in prompt) ---

func (s *service) Submit(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	slog.Info("Service.Submit called (stub)", "subscriptionURL", sub.URL)
	dataSub := &data.Subscription{
		URL:      sub.URL,
		Params:   sub.Params,
		CronExpr: sub.CronExpr,
	}
	if sub.Id == "" {
		dataSub.Id = uuid.NewString() 
	} else {
		dataSub.Id = sub.Id
	}

	savedDataSub, err := s.repo.Submit(ctx, dataSub) 
	if err != nil {
		slog.Error("repo.Submit failed in service stub", "error", err)
		return nil, fmt.Errorf("repo.Submit failed: %w", err)
	}

	return &domain.Subscription{
		Id:       savedDataSub.Id,
		URL:      savedDataSub.URL,
		Params:   savedDataSub.Params,
		CronExpr: savedDataSub.CronExpr,
	}, nil
}

func (s *service) List(ctx context.Context, start int64, limit int) (*domain.PaginatedResponse[[]domain.Subscription], error) {
	slog.Info("Service.List called (stub)", "start", start, "limit", limit)
	dataSubs, err := s.repo.List(ctx, start, limit)
	if err != nil {
		slog.Error("repo.List failed in service stub", "error", err)
		return nil, fmt.Errorf("repo.List failed: %w", err)
	}
	if dataSubs == nil || len(*dataSubs) == 0 {
		return &domain.PaginatedResponse[[]domain.Subscription]{Data: []domain.Subscription{}}, nil
	}

	domainSubs := make([]domain.Subscription, len(*dataSubs))
	for i, ds := range *dataSubs {
		domainSubs[i] = domain.Subscription{
			Id:       ds.Id,
			URL:      ds.URL,
			Params:   ds.Params,
			CronExpr: ds.CronExpr,
		}
	}

	var firstCursor int64 = 0 
	var nextCursor int64 = 0  
	
	return &domain.PaginatedResponse[[]domain.Subscription]{
		First: firstCursor,
		Next:  nextCursor,
		Data:  domainSubs,
	}, nil
}

func (s *service) UpdateByExample(ctx context.Context, example *domain.Subscription) error {
	slog.Info("Service.UpdateByExample called (stub)", "subscriptionID", example.Id)
	dataSub := &data.Subscription{
		Id:       example.Id,
		URL:      example.URL,
		Params:   example.Params,
		CronExpr: example.CronExpr,
	}
	return s.repo.UpdateByExample(ctx, dataSub)
}

func (s *service) Delete(ctx context.Context, id string) error {
	slog.Info("Service.Delete called (stub)", "subscriptionID", id)
	return s.repo.Delete(ctx, id)
}

func (s *service) GetCursor(ctx context.Context, id string) (int64, error) {
	slog.Info("Service.GetCursor called (stub)", "subscriptionID", id)
	return s.repo.GetCursor(ctx, id)
}
