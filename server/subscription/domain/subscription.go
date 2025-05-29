package domain

import (
	"context"
	"net/http"
	"time" // Added time import

	"github.com/go-chi/chi/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/data"
)

type Subscription struct {
	Id       string `json:"id"`
	URL      string `json:"url"`
	Params   string `json:"params"`
	CronExpr string `json:"cron_expression"`
}

type PaginatedResponse[T any] struct {
	First int64 `json:"first"`
	Next  int64 `json:"next"`
	Data  T     `json:"data"`
}

type Repository interface {
	Submit(ctx context.Context, sub *data.Subscription) (*data.Subscription, error)
	List(ctx context.Context, start int64, limit int) (*[]data.Subscription, error)
	Get(ctx context.Context, id string) (*data.Subscription, error) // New method
	UpdateByExample(ctx context.Context, example *data.Subscription) error
	Delete(ctx context.Context, id string) error
	GetCursor(ctx context.Context, id string) (int64, error)

	// Methods for SubscriptionVideoUpdate
	InsertSubscriptionUpdate(ctx context.Context, update *data.SubscriptionVideoUpdate) error
	GetUnseenUpdatesCount(ctx context.Context, subscriptionIDs []string) (int, error)
	ListUnseenUpdates(ctx context.Context, limit int, offset int, subscriptionIDs []string) ([]data.SubscriptionVideoUpdate, error)
	MarkUpdateAsSeen(ctx context.Context, updateID string, seen bool) error
	MarkAllUpdatesAsSeen(ctx context.Context, subscriptionIDs []string, seen bool) (int64, error)
	UpdateSubscriptionUpdateStatus(ctx context.Context, updateID string, status string) error
	GetSubscriptionUpdateByVideoURL(ctx context.Context, videoURL string) (*data.SubscriptionVideoUpdate, error)
	DeleteSubscriptionUpdate(ctx context.Context, updateID string) error
	GetSubscriptionUpdate(ctx context.Context, updateID string) (*data.SubscriptionVideoUpdate, error) // New
}

type Service interface {
	Submit(ctx context.Context, sub *Subscription) (*Subscription, error)
	List(ctx context.Context, start int64, limit int) (*PaginatedResponse[[]Subscription], error)
	UpdateByExample(ctx context.Context, example *Subscription) error
	Delete(ctx context.Context, id string) error
	GetCursor(ctx context.Context, id string) (int64, error)
	GetChannelVideos(ctx context.Context, subscriptionID string) (*YtdlpChannelDump, error) // New method

	// Methods for SubscriptionVideoUpdate
	CreateSubscriptionUpdate(ctx context.Context, update *SubscriptionVideoUpdate) (*SubscriptionVideoUpdate, error)
	GetUnseenUpdatesCount(ctx context.Context, subscriptionIDs []string) (int, error)
	ListUnseenUpdates(ctx context.Context, limit int, offset int, subscriptionIDs []string) ([]SubscriptionVideoUpdate, error)
	MarkUpdateAsSeen(ctx context.Context, updateID string, seen bool) error
	MarkAllUpdatesAsSeen(ctx context.Context, subscriptionIDs []string, seen bool) (int64, error)
	UpdateSubscriptionUpdateStatus(ctx context.Context, updateID string, status string) error
	DeleteSubscriptionUpdate(ctx context.Context, updateID string) error
	GetSubscriptionUpdate(ctx context.Context, updateID string) (*SubscriptionVideoUpdate, error) // Added for DownloadUpdate handler
}

type RestHandler interface {
	Submit() http.HandlerFunc
	List() http.HandlerFunc
	UpdateByExample() http.HandlerFunc
	Delete() http.HandlerFunc
	GetCursor() http.HandlerFunc
	GetChannelVideos() http.HandlerFunc // Should already exist

	// New methods for SubscriptionVideoUpdate
	ListUpdates() http.HandlerFunc
	GetUnseenUpdatesCount() http.HandlerFunc
	MarkUpdateSeen() http.HandlerFunc
	MarkAllUpdatesSeen() http.HandlerFunc
	DownloadUpdate() http.HandlerFunc
	DeleteUpdate() http.HandlerFunc

	ApplyRouter() func(chi.Router)
}

// SubscriptionVideoUpdate struct for the domain layer
type SubscriptionVideoUpdate struct {
	Id             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	VideoURL       string    `json:"video_url"`
	VideoTitle     string    `json:"video_title"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"`
	DetectedAt     time.Time `json:"detected_at"`
	IsSeen         bool      `json:"is_seen"`
	Status         string    `json:"status"` // e.g., 'new', 'downloaded', 'dismissed'
}

// YtdlpChannelDump and YtdlpVideoInfo are already here from previous work
// (assuming they were added in the file if not shown in the provided snippet)
// If not, they should be added as well if they are part of this domain file.
// For now, focusing on SubscriptionVideoUpdate as per explicit instruction.
// YtdlpChannelDump structure (if not already present)
type YtdlpChannelDump struct {
	Entries    []YtdlpVideoInfo `json:"entries"`
	ID         string           `json:"id"` // Channel/Playlist ID
	Title      string           `json:"title"` // Channel/Playlist Title
	Uploader   string           `json:"uploader,omitempty"`
	UploaderID string           `json:"uploader_id,omitempty"`
	// Add other fields from yt-dlp output as needed
}

// YtdlpVideoInfo structure (if not already present)
type YtdlpVideoInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	WebpageURL  string `json:"webpage_url"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	UploadDate  string `json:"upload_date,omitempty"` // YYYYMMDD
	// Add other fields from yt-dlp output as needed
}
