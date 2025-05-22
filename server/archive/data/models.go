package data

import "time" // Ensure time is imported if not already

type ArchiveEntry struct {
	RowId     int64 // New field
	Id        string
	Title     string
	Path      string
	Thumbnail string
	Source    string
	Metadata  string
	CreatedAt time.Time
}
