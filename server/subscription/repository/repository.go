package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/data"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
)

type Repository struct {
	db *sql.DB
}

// Delete implements domain.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, "DELETE FROM subscriptions WHERE id = ?", id)

	return err
}

// GetCursor implements domain.Repository.
func (r *Repository) GetCursor(ctx context.Context, id string) (int64, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return -1, err
	}

	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT rowid FROM subscriptions WHERE id = ?", id)

	var rowId int64

	if err := row.Scan(&rowId); err != nil {
		return -1, err
	}

	return rowId, nil
}

// List implements domain.Repository.
func (r *Repository) List(ctx context.Context, start int64, limit int) (*[]data.Subscription, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var elements []data.Subscription

	rows, err := conn.QueryContext(ctx, "SELECT rowid, * FROM subscriptions WHERE rowid > ? LIMIT ?", start, limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var rowId int64
		var element data.Subscription

		if err := rows.Scan(
			&rowId,
			&element.Id,
			&element.URL,
			&element.Params,
			&element.CronExpr,
		); err != nil {
			return &elements, err
		}

		elements = append(elements, element)
	}

	return &elements, nil
}

// Get implements domain.Repository.
func (r *Repository) Get(ctx context.Context, id string) (*data.Subscription, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT id, url, params, cron FROM subscriptions WHERE id = ?", id)

	var sub data.Subscription
	err = row.Scan(&sub.Id, &sub.URL, &sub.Params, &sub.CronExpr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Standard way to indicate "not found" without it being an application error yet
		}
		return nil, err // Other errors
	}
	return &sub, nil
}

// Submit implements domain.Repository.
func (r *Repository) Submit(ctx context.Context, sub *data.Subscription) (*data.Subscription, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO subscriptions (id, url, params, cron) VALUES (?, ?, ?, ?)",
		uuid.NewString(),
		sub.URL,
		sub.Params,
		sub.CronExpr,
	)

	return sub, err
}

// UpdateByExample implements domain.Repository.
func (r *Repository) UpdateByExample(ctx context.Context, example *data.Subscription) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		"UPDATE subscriptions SET url = ?, params = ?, cron = ? WHERE id = ? OR url = ?",
		example.URL,
		example.Params,
		example.CronExpr,
		example.Id,
		example.URL,
	)

	return err
}

func New(db *sql.DB) domain.Repository {
	return &Repository{
		db: db,
	}
}

// --- Implementation of new methods for SubscriptionVideoUpdate ---

func (r *Repository) InsertSubscriptionUpdate(ctx context.Context, update *data.SubscriptionVideoUpdate) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if update.Id == "" {
		update.Id = uuid.NewString()
	}
	if update.DetectedAt.IsZero() {
		update.DetectedAt = time.Now()
	}
	if update.Status == "" {
		update.Status = "new" // Default status
	}


	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO subscription_video_updates (id, subscription_id, video_url, video_title, thumbnail_url, published_at, detected_at, is_seen, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		update.Id,
		update.SubscriptionID,
		update.VideoURL,
		update.VideoTitle,
		update.ThumbnailURL,
		update.PublishedAt,
		update.DetectedAt,
		update.IsSeen,
		update.Status,
	)
	return err
}

func (r *Repository) GetUnseenUpdatesCount(ctx context.Context, subscriptionIDs []string) (int, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	query := "SELECT COUNT(*) FROM subscription_video_updates WHERE is_seen = FALSE"
	args := []interface{}{}

	if len(subscriptionIDs) > 0 {
		placeholders := strings.Repeat("?,", len(subscriptionIDs)-1) + "?"
		query += " AND subscription_id IN (" + placeholders + ")"
		for _, id := range subscriptionIDs {
			args = append(args, id)
		}
	}

	var count int
	err = conn.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *Repository) ListUnseenUpdates(ctx context.Context, limit int, offset int, subscriptionIDs []string) ([]data.SubscriptionVideoUpdate, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var updates []data.SubscriptionVideoUpdate
	query := "SELECT id, subscription_id, video_url, video_title, thumbnail_url, published_at, detected_at, is_seen, status FROM subscription_video_updates WHERE is_seen = FALSE"
	args := []interface{}{}

	if len(subscriptionIDs) > 0 {
		placeholders := strings.Repeat("?,", len(subscriptionIDs)-1) + "?"
		query += " AND subscription_id IN (" + placeholders + ")"
		for _, id := range subscriptionIDs {
			args = append(args, id)
		}
	}
	query += " ORDER BY detected_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var update data.SubscriptionVideoUpdate
		if err := rows.Scan(
			&update.Id,
			&update.SubscriptionID,
			&update.VideoURL,
			&update.VideoTitle,
			&update.ThumbnailURL,
			&update.PublishedAt,
			&update.DetectedAt,
			&update.IsSeen,
			&update.Status,
		); err != nil {
			return updates, err
		}
		updates = append(updates, update)
	}
	return updates, rows.Err()
}

func (r *Repository) MarkUpdateAsSeen(ctx context.Context, updateID string, seen bool) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "UPDATE subscription_video_updates SET is_seen = ? WHERE id = ?", seen, updateID)
	return err
}

func (r *Repository) MarkAllUpdatesAsSeen(ctx context.Context, subscriptionIDs []string, seen bool) (int64, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	query := "UPDATE subscription_video_updates SET is_seen = ? WHERE is_seen = ?" // Default: mark unseen (FALSE) as seen (TRUE)
	args := []interface{}{seen, !seen} // If seen is true, we want to update where is_seen = false. If seen is false, update where is_seen = true.

	if len(subscriptionIDs) > 0 {
		placeholders := strings.Repeat("?,", len(subscriptionIDs)-1) + "?"
		query += " AND subscription_id IN (" + placeholders + ")"
		for _, id := range subscriptionIDs {
			args = append(args, id)
		}
	}

	res, err := conn.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) UpdateSubscriptionUpdateStatus(ctx context.Context, updateID string, status string) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "UPDATE subscription_video_updates SET status = ? WHERE id = ?", status, updateID)
	return err
}

func (r *Repository) GetSubscriptionUpdateByVideoURL(ctx context.Context, videoURL string) (*data.SubscriptionVideoUpdate, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT id, subscription_id, video_url, video_title, thumbnail_url, published_at, detected_at, is_seen, status FROM subscription_video_updates WHERE video_url = ?", videoURL)
	var update data.SubscriptionVideoUpdate
	err = row.Scan(
		&update.Id,
		&update.SubscriptionID,
		&update.VideoURL,
		&update.VideoTitle,
		&update.ThumbnailURL,
		&update.PublishedAt,
		&update.DetectedAt,
		&update.IsSeen,
		&update.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &update, nil
}

func (r *Repository) DeleteSubscriptionUpdate(ctx context.Context, updateID string) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "DELETE FROM subscription_video_updates WHERE id = ?", updateID)
	return err
}

func (r *Repository) GetSubscriptionUpdate(ctx context.Context, updateID string) (*data.SubscriptionVideoUpdate, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT id, subscription_id, video_url, video_title, thumbnail_url, published_at, detected_at, is_seen, status FROM subscription_video_updates WHERE id = ?", updateID)
	var update data.SubscriptionVideoUpdate
	err = row.Scan(
		&update.Id,
		&update.SubscriptionID,
		&update.VideoURL,
		&update.VideoTitle,
		&update.ThumbnailURL,
		&update.PublishedAt,
		&update.DetectedAt,
		&update.IsSeen,
		&update.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &update, nil
}
