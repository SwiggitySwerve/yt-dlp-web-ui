package data

import "time" // Ensure time is imported

type ArchiveEntry struct {
	RowId     int64     `json:"-"` // Internal cursor ID, not for client JSON
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	Path      string    `json:"path"`
	Thumbnail string    `json:"thumbnail"`
	Source    string    `json:"source"`
	Metadata  string    `json:"metadata"`  // JSON string of full metadata
	CreatedAt time.Time `json:"created_at"`
	Duration  int64     `json:"duration,omitempty"` // New, in seconds
	Format    string    `json:"format,omitempty"`   // New, e.g., "mp4", "webm"
	Uploader  string    `json:"uploader,omitempty"` // New field for uploader
}
