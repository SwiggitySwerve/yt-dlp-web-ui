package common

import "time"

// Used to deser the yt-dlp -J output
type DownloadInfo struct {
	URL            string    `json:"webpage_url"` // Often webpage_url is the primary URL field from -J
	OriginalURL    string    `json:"original_url,omitempty"` // The URL as initially requested
	Title          string    `json:"title"`
	Thumbnail      string    `json:"thumbnail"`
	Resolution     string    `json:"resolution,omitempty"`
	Vcodec         string    `json:"vcodec,omitempty"`
	Acodec         string    `json:"acodec,omitempty"`
	Ext            string    `json:"ext,omitempty"` 
	FormatID       string    `json:"format_id,omitempty"` // yt-dlp specific format id
	FormatNote     string    `json:"format_note,omitempty"`// e.g. "1080p"
	Duration       float64   `json:"duration,omitempty"`   // Duration in seconds, yt-dlp often gives float
	CreatedAt      time.Time `json:"-"` // Internal, not from yt-dlp JSON usually for this field name. This is set by our app.
	FilesizeApprox int64     `json:"filesize_approx,omitempty"` 
    // Removed FileName as it's not standard in yt-dlp -J for this, and can be derived.
    // Kept Size as FilesizeApprox, as per yt-dlp -J output.
}
