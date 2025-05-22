// server/subscription/domain/channel_info.go
package domain

// YtdlpVideoInfo represents a single video entry from yt-dlp.
type YtdlpVideoInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"` // yt-dlp often provides a direct thumbnail URL
	// Thumbnails  []struct { // Or sometimes it's an array of thumbnails
	//	URL      string `json:"url"`
	//	Height   int    `json:"height,omitempty"`
	//	Width    int    `json:"width,omitempty"`
	//	Resolution string `json:"resolution,omitempty"`
	//	ID       string `json:"id,omitempty"`
	//} `json:"thumbnails,omitempty"`
	Duration        float64 `json:"duration,omitempty"` // Duration in seconds
	WebpageURL      string `json:"webpage_url,omitempty"`
	Uploader        string `json:"uploader,omitempty"`
	UploaderID      string `json:"uploader_id,omitempty"`
	UploaderURL     string `json:"uploader_url,omitempty"`
	UploadDate      string `json:"upload_date,omitempty"` // Format YYYYMMDD
	ViewCount       int64  `json:"view_count,omitempty"`
	LikeCount       int64  `json:"like_count,omitempty"`
	AverageRating   float64 `json:"average_rating,omitempty"`
	IsLive          bool   `json:"is_live,omitempty"`
	PlaylistIndex   int    `json:"playlist_index,omitempty"` // Useful for series
	PlaylistID      string `json:"playlist_id,omitempty"`
	PlaylistTitle   string `json:"playlist_title,omitempty"`
	PlaylistUploader string `json:"playlist_uploader,omitempty"`
	Extractor       string `json:"extractor,omitempty"` // e.g., "youtube"
	ExtractorKey    string `json:"extractor_key,omitempty"` // e.g., "Youtube"
	IsDownloaded  bool   `json:"is_downloaded"` // New field
}

// YtdlpChannelDump represents the overall JSON structure when dumping a channel/playlist.
// This typically includes channel-level metadata and a list of video entries.
type YtdlpChannelDump struct {
	Entries         []YtdlpVideoInfo `json:"entries,omitempty"` // For playlists or channels
	ID              string           `json:"id"`                // Channel/Playlist ID
	Title           string           `json:"title"`             // Channel/Playlist Title
	Uploader        string           `json:"uploader,omitempty"`  // Channel uploader name (if channel URL)
	UploaderID      string           `json:"uploader_id,omitempty"` // Channel uploader ID (if channel URL)
	UploaderURL     string           `json:"uploader_url,omitempty"`
	Description     string           `json:"description,omitempty"`
	WebpageURL      string           `json:"webpage_url,omitempty"` // URL of the channel/playlist itself
	OriginalURL     string           `json:"original_url,omitempty"`// The URL that was passed to yt-dlp
	Extractor       string           `json:"extractor,omitempty"`
	ExtractorKey    string           `json:"extractor_key,omitempty"`
	// Include other channel-specific fields if necessary
}

// Note: The exact structure of the JSON can vary slightly based on the source (YouTube, Vimeo, etc.)
// and whether it's a channel or a playlist. These structs aim to be comprehensive.
// --flat-playlist should simplify 'entries' to be mostly video-like objects.
