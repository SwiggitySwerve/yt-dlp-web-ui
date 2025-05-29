package data

import "time" // Ensure time is imported if not already

// Existing Subscription struct might be here
type Subscription struct {
	Id       string
	URL      string
	Params   string
	CronExpr string
}

type SubscriptionVideoUpdate struct {
	Id             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	VideoURL       string    `json:"video_url"`
	VideoTitle     string    `json:"video_title"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"` // From video metadata
	DetectedAt     time.Time `json:"detected_at"`          // When this record was created
	IsSeen         bool      `json:"is_seen"`
	Status         string    `json:"status"` // e.g., 'new', 'downloaded', 'dismissed'
}
