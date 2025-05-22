// server/subscription/service/service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	// "errors" // Not explicitly used in the provided GetChannelVideos, but good for general Go.
	"fmt"
	"log/slog" // For logging
	"os/exec"

	"github.com/google/uuid"                                                      // For temporary ID generation in Submit
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/data" // For data.Subscription
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
)

type service struct {
	repo domain.Repository
	// Add other dependencies if needed, like a logger
}

// NewService creates a new subscription service.
func NewService(repo domain.Repository) domain.Service {
	return &service{
		repo: repo,
	}
}

// GetChannelVideos implements domain.Service.
func (s *service) GetChannelVideos(ctx context.Context, subscriptionID string) (*domain.YtdlpChannelDump, error) {
	// ** IMPORTANT LATER STEP: Fix this to fetch URL from repo using subscriptionID **
	// For the purpose of this subtask, we will use a dummy URL.
	// A real implementation MUST fetch the URL from the subscriptionID.
	// The `domain.Repository` interface currently lacks a GetById method. This will be added later.
	
	var channelURL string = "https://www.youtube.com/@youtube/videos" // Default to a known channel for testing
	slog.Warn(
		"Placeholder: Subscription fetching logic is not fully implemented. Using hardcoded default URL for yt-dlp.", 
		"subscriptionID", subscriptionID, 
		"hardcodedURL", channelURL,
	)
	// Example of how it might look with a Get method on the repo:
	/*
	actualSub, err := s.repo.Get(ctx, subscriptionID) // Assuming s.repo has a Get method
	if err != nil {
		slog.Error("Failed to fetch subscription by ID", "subscriptionID", subscriptionID, "error", err)
		return nil, fmt.Errorf("failed to retrieve subscription %s: %w", subscriptionID, err)
	}
	if actualSub == nil { // Should ideally be handled by error in Get, but good practice
		slog.Error("Subscription not found", "subscriptionID", subscriptionID)
		return nil, fmt.Errorf("subscription with ID %s not found", subscriptionID)
	}
	channelURL = actualSub.URL // Assuming data.Subscription has a URL field
	*/

	slog.Info("GetChannelVideos called", "subscriptionID", subscriptionID, "effectiveChannelURL", channelURL)


	// 2. Construct and execute yt-dlp command
	cmd := exec.CommandContext(ctx, config.Instance().DownloaderPath,
		channelURL, // Use the (currently hardcoded) URL here
		"--dump-single-json",
		"--flat-playlist",
		"--no-warnings",
		// Consider adding "--cookies-from-browser", "chrome" or similar if needed for private content
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	slog.Info("Executing yt-dlp", "command", cmd.String())

	err := cmd.Run()
	if err != nil {
		slog.Error("yt-dlp command failed", "error", err, "stderr", stderr.String())
		return nil, fmt.Errorf("yt-dlp command failed: %w; Stderr: %s", err, stderr.String())
	}

	// 3. Unmarshal JSON output
	var channelDump domain.YtdlpChannelDump
	if err := json.Unmarshal(out.Bytes(), &channelDump); err != nil {
		slog.Error("Failed to unmarshal yt-dlp JSON output", "error", err, "outputSize", len(out.Bytes()))
		// For debugging, log a snippet of the output:
		// outputForLogging := out.String()
		// if len(outputForLogging) > 500 { outputForLogging = outputForLogging[:500] + "..." }
		// slog.Debug("yt-dlp raw output on unmarshal error", "output", outputForLogging)
		return nil, fmt.Errorf("failed to unmarshal yt-dlp JSON output: %w", err)
	}
	
	// If original URL isn't in the dump (common with --dump-single-json), set it from what was used.
	if channelDump.OriginalURL == "" {
		channelDump.OriginalURL = channelURL
	}

	slog.Info("Successfully fetched and parsed channel videos", "channelTitle", channelDump.Title, "entryCount", len(channelDump.Entries), "subscriptionID", subscriptionID)
	return &channelDump, nil
}

// --- Stub implementations for other domain.Service methods ---

func (s *service) Submit(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	slog.Info("Service.Submit called (stub)", "subscriptionURL", sub.URL)
	dataSub := &data.Subscription{
		Id:       uuid.NewString(), // Placeholder ID generation. Repo should ideally handle ID.
		URL:      sub.URL,
		Params:   sub.Params,
		CronExpr: sub.CronExpr,
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
	if dataSubs == nil || len(*dataSubs) == 0 { // Ensure dataSubs is not nil before checking length
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
	
	var firstCursor int64 = 0 // Placeholder
	var nextCursor int64 = 0  // Placeholder
	// Actual cursor logic would require more info from repo.List or how IDs are structured.
	// For example, if 'start' is an offset and results are ordered:
	// firstCursor = start 
	// if len(domainSubs) == limit { nextCursor = start + int64(limit) }
	// If 'start' is an ID (which it is for this repo), this logic is more complex
	// and depends on the ordering of IDs and whether they are sequential.
	// The current repo.List seems to take a 'start' ID.
	if len(domainSubs) > 0 {
		// This is a guess. The repo.List method needs to provide actual cursor info or be based on offset.
		// If start is the ID of the first item, then firstCursor could be that ID (if convertible to int64).
		// The `start` parameter for `repo.List` is an int64, but subscription IDs are strings.
		// This highlights a mismatch that needs to be resolved in repository/service layers.
		// For now, returning 0,0 as placeholders.
	}

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
	// This method retrieves a numerical cursor for a given string ID.
	// The current repo.GetCursor already does this.
	return s.repo.GetCursor(ctx, id)
}
