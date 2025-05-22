package rpc

import (
	"errors"
	"log/slog"
	"strings" // Added for sanitization

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/formats"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal/livestream"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/sys"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/updater"
)

type Service struct {
	db *internal.MemoryDB
	mq *internal.MessageQueue
	lm *livestream.Monitor
}

type Running []internal.ProcessResponse
type Pending []string

type NoArgs struct{}

// Exec spawns a Process.
// The result of the execution is the newly spawned process Id.
func (s *Service) Exec(args internal.DownloadRequest, result *string) error {
	// Sanitize ChannelFolder
	var sanitizedChannelFolder string
	if args.ChannelFolder != "" {
		tempFolder := args.ChannelFolder
		// Basic sanitization: replace potentially problematic characters or sequences
		tempFolder = strings.ReplaceAll(tempFolder, "/", "_")
		tempFolder = strings.ReplaceAll(tempFolder, "\\", "_") // For literal backslash
		tempFolder = strings.ReplaceAll(tempFolder, "..", "_") // Prevent path traversal

		// A more robust sanitization for common invalid filename characters.
		// Note: This is a simplified list. OS-specific rules can be more complex.
		// Consider a library for more thorough cross-platform sanitization if needed.
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"} 
		for _, char := range invalidChars {
			tempFolder = strings.ReplaceAll(tempFolder, char, "_")
		}
		
		// Ensure it's not empty after sanitization if it was originally non-empty.
		// This handles cases like args.ChannelFolder being just ".." or "//".
		if tempFolder == "" {
			sanitizedChannelFolder = "_" // Or a more descriptive default like "default_subfolder"
		} else {
			sanitizedChannelFolder = tempFolder
		}
		slog.Info("Channel folder requested for download", "original", args.ChannelFolder, "sanitized", sanitizedChannelFolder)
	}

	p := &internal.Process{
		Url:    args.URL,
		Params: args.Params,
		Output: internal.DownloadOutput{
			Path:          args.Path,    // Base path from request
			Filename:      args.Rename,  // Filename template from request
			ChannelFolder: sanitizedChannelFolder, // Assign sanitized folder name
		},
		PreferredFormats:   args.PreferredFormats,   // New
		PreferredQualities: args.PreferredQualities, // New
	}

	s.db.Set(p)
	s.mq.Publish(p)

	*result = p.Id
	return nil
}

// ExecPlaylist spawns a Process for each item in a playlist.
// The result of the execution is the newly spawned process Id. (This behavior might need adjustment for playlists)
func (s *Service) ExecPlaylist(args internal.DownloadRequest, result *string) error {
	// Note: The ChannelFolder from args would apply to all videos in this playlist.
	// The internal.PlaylistDetect function will need to be aware of this or
	// args passed to it should include the sanitizedChannelFolder.
	// For now, PlaylistDetect is not modified by this subtask, but it's a consideration.
	slog.Info("ExecPlaylist called", "url", args.URL, "channelFolder", args.ChannelFolder)

	// It's important that PlaylistDetect correctly uses the ChannelFolder.
	// If PlaylistDetect itself creates new internal.Process objects, it needs to
	// receive and use the sanitizedChannelFolder.
	// For this subtask, we only ensure ExecPlaylist *could* pass it if PlaylistDetect supported it.
	// The current internal.PlaylistDetect might not be expecting ChannelFolder in its args.
	// We are modifying `args` here, but `PlaylistDetect` might not use `args.ChannelFolder`.
	// This requires `PlaylistDetect` to be updated to use `args.ChannelFolder` when creating processes.
	// For now, we assume `PlaylistDetect` will be updated or this is a non-functional change for playlists
	// until `PlaylistDetect` is refactored.

	// Sanitize ChannelFolder for playlist (same logic as Exec)
	var sanitizedChannelFolderForPlaylist string
	if args.ChannelFolder != "" {
		tempFolder := args.ChannelFolder
		tempFolder = strings.ReplaceAll(tempFolder, "/", "_")
		tempFolder = strings.ReplaceAll(tempFolder, "\\", "_")
		tempFolder = strings.ReplaceAll(tempFolder, "..", "_")
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			tempFolder = strings.ReplaceAll(tempFolder, char, "_")
		}
		if tempFolder == "" {
			sanitizedChannelFolderForPlaylist = "_"
		} else {
			sanitizedChannelFolderForPlaylist = tempFolder
		}
		slog.Info("Using channel folder for playlist download", "original", args.ChannelFolder, "sanitized", sanitizedChannelFolderForPlaylist)
        args.ChannelFolder = sanitizedChannelFolderForPlaylist // Update args for PlaylistDetect
	}


	err := internal.PlaylistDetect(args, s.mq, s.db) // PlaylistDetect needs to use args.ChannelFolder
	if err != nil {
		return err
	}

	*result = "" // Typically for playlists, individual process IDs are handled, not one single ID.
	return nil
}

// ExecLivestream handles livestream monitoring requests.
func (s *Service) ExecLivestream(args internal.DownloadRequest, result *string) error {
	// Livestreams typically don't use ChannelFolder in the same way as direct downloads,
	// as their output path is usually fixed or handled differently by yt-dlp's live options.
	// If ChannelFolder were to be used, similar sanitization and Process setup would be needed.
	// For now, it's ignored for livestreams as per current structure.
	slog.Info("ExecLivestream called", "url", args.URL)
	s.lm.Add(args.URL)

	*result = args.URL
	return nil
}

// ProgressLivestream retrieves the status of monitored livestreams.
func (s *Service) ProgressLivestream(args NoArgs, result *livestream.LiveStreamStatus) error {
	*result = s.lm.Status()
	return nil
}

// KillLivestream stops monitoring a specific livestream.
func (s *Service) KillLivestream(args string, result *struct{}) error {
	slog.Info("killing livestream", slog.String("url", args))

	err := s.lm.Remove(args)
	if err != nil {
		slog.Error("failed killing livestream", slog.String("url", args), slog.Any("err", err))
		return err
	}

	return nil
}

// KillAllLivestream stops monitoring all livestreams.
func (s *Service) KillAllLivestream(args NoArgs, result *struct{}) error {
	return s.lm.RemoveAll()
}

// Progess retrieves the Progress of a specific Process given its Id
func (s *Service) Progess(args internal.DownloadRequest, progress *internal.DownloadProgress) error {
	proc, err := s.db.Get(args.Id)
	if err != nil {
		return err
	}

	*progress = proc.Progress
	return nil
}

// Formats retrieves available format for a given resource
func (s *Service) Formats(args internal.DownloadRequest, meta *formats.Metadata) error {
	var err error
	// If ChannelFolder is relevant for format fetching (e.g., if it implies different cookies or auth),
	// that would be a more advanced scenario. For now, it's not used here.
	metadata, err := formats.ParseURL(args.URL)
	if err != nil && metadata == nil { // metadata can be non-nil even with error (e.g. for playlists)
		return err
	}

	if metadata != nil && metadata.IsPlaylist() {
        // If PlaylistDetect is called here, it also needs to be aware of ChannelFolder if it's to be used.
        // The current args for PlaylistDetect might not include it, or PlaylistDetect might not use it.
        // For now, just passing original args.
		go internal.PlaylistDetect(args, s.mq, s.db)
	}

	if metadata == nil { // If error occurred and metadata is nil
		return errors.New("failed to retrieve formats and metadata was nil")
	}
	
	*meta = *metadata
	return nil
}

// Pending retrieves a slice of all Pending/Running processes ids
func (s *Service) Pending(args NoArgs, pending *Pending) error {
	*pending = *s.db.Keys()
	return nil
}

// Running retrieves a slice of all Processes progress
func (s *Service) Running(args NoArgs, running *Running) error {
	*running = *s.db.All()
	return nil
}

// Kill kills a process given its id and remove it from the memoryDB
func (s *Service) Kill(args string, killed *string) error {
	slog.Info("Trying killing process with id", slog.String("id", args))

	proc, err := s.db.Get(args)
	if err != nil {
		return err
	}

	if proc == nil {
		return errors.New("nil process")
	}

	if err := proc.Kill(); err != nil {
		slog.Info("failed killing process", slog.String("id", proc.Id), slog.Any("err", err))
		return err
	}

	s.db.Delete(proc.Id)
	slog.Info("succesfully killed process", slog.String("id", proc.Id))
	*killed = proc.Id // Return the ID of the killed process
	return nil
}

// KillAll kills all process unconditionally and removes them from
// the memory db
func (s *Service) KillAll(args NoArgs, killed *string) error { // result type change for consistency? (e.g. stream of IDs)
	slog.Info("Killing all spawned processes")

	var (
		keys       = s.db.Keys()
		removeFunc = func(p *internal.Process) error {
			defer s.db.Delete(p.Id)
			return p.Kill()
		}
	)
    var killedIDs []string // To store IDs of processes attempted to be killed

	for _, key := range *keys {
		proc, err := s.db.Get(key)
		if err != nil {
			// Log error but continue to try killing others
			slog.Error("Failed to get process for KillAll", "id", key, "error", err)
			continue
		}

		if proc == nil {
			s.db.Delete(key) // Clean up nil process entry
			continue
		}

		if err := removeFunc(proc); err != nil {
			slog.Info(
				"failed killing process during KillAll",
				slog.String("id", proc.Id),
				slog.Any("err", err),
			)
			continue // Continue to next process
		}
        killedIDs = append(killedIDs, proc.Id)
		slog.Info("succesfully killed process during KillAll", slog.String("id", proc.Id))
		proc = nil // gc helper
	}
    *killed = strings.Join(killedIDs, ", ") // Return comma-separated list of killed IDs
	return nil
}

// Clear a process from the db rendering it unusable if active
func (s *Service) Clear(args string, killed *string) error {
	slog.Info("Clearing process with id", slog.String("id", args))
	s.db.Delete(args)
    *killed = args // Return the ID of the cleared process
	return nil
}

// ClearCompleted removes completed processes
func (s *Service) ClearCompleted(args NoArgs, clearedCount *int) error { // Changed result to count
	var (
		keys       = s.db.Keys()
		count      = 0
		removeFunc = func(p *internal.Process) error {
			if p.Progress.Status == internal.StatusCompleted {
				s.db.Delete(p.Id) // Just delete from DB, don't try to kill a completed process
                count++
			}
			return nil // No error if not completed or if deletion is fine
		}
	)

	for _, key := range *keys {
		proc, err := s.db.Get(key)
		if err != nil {
            slog.Error("Failed to get process for ClearCompleted", "id", key, "error", err)
			continue // Skip if error getting process
		}
        if proc == nil {
            s.db.Delete(key) // Clean up nil process entry
            continue
        }
		if err := removeFunc(proc); err != nil {
			// This error path should ideally not be hit if removeFunc is just deleting.
            slog.Error("Error in removeFunc for ClearCompleted", "id", proc.Id, "error", err)
			// Depending on error, might continue or return
		}
	}
    *clearedCount = count
	slog.Info("ClearCompleted finished", "clearedCount", count)
	return nil
}

// FreeSpace gets the available from package sys util
func (s *Service) FreeSpace(args NoArgs, free *uint64) error {
	freeSpace, err := sys.FreeSpace()
	if err != nil {
		return err
	}

	*free = freeSpace
	return nil
}

// DirectoryTree returns a flattned tree of the download directory
func (s *Service) DirectoryTree(args NoArgs, tree *[]string) error {
	dfsTree, err := sys.DirectoryTree()

	if err != nil {
		*tree = nil
		return err
	}

	if dfsTree != nil {
		*tree = *dfsTree
	} else {
        *tree = []string{} // Ensure non-nil response
    }

	return nil
}

// UpdateExecutable updates the yt-dlp binary using its builtin function
func (s *Service) UpdateExecutable(args NoArgs, updated *bool) error {
	slog.Info("Updating yt-dlp executable to the latest release")

	if err := updater.UpdateExecutable(); err != nil {
		slog.Error("Failed updating yt-dlp", "error", err)
		*updated = false
		return err
	}

	*updated = true
	slog.Info("Succesfully updated yt-dlp")

	return nil
}
