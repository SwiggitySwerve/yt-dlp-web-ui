package repository

import (
	"context"
	"database/sql"
	"os"
	"strings" // Added import
	"log/slog"  // Added import

	"github.com/google/uuid"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/data"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) domain.Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Archive(ctx context.Context, entry *data.ArchiveEntry) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO archive (id, title, path, thumbnail, source, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		uuid.NewString(),
		entry.Title,
		entry.Path,
		entry.Thumbnail,
		entry.Source,
		entry.Metadata,
		entry.CreatedAt,
	)
	return err
}

func (r *Repository) SoftDelete(ctx context.Context, id string) (*data.ArchiveEntry, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var model data.ArchiveEntry
	row := tx.QueryRowContext(ctx, "SELECT id, title, path, thumbnail, source, metadata, created_at FROM archive WHERE id = ?", id)
	if err := row.Scan(
		&model.Id,
		&model.Title,
		&model.Path,
		&model.Thumbnail,
		&model.Source,
		&model.Metadata,
		&model.CreatedAt,
	); err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM archive WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *Repository) HardDelete(ctx context.Context, id string) (*data.ArchiveEntry, error) {
	entry, err := r.SoftDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(entry.Path); err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *Repository) List(ctx context.Context, startRowId int, limit int, sortBy string, filterByUploader string) (*[]data.ArchiveEntry, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT rowid, id, title, path, thumbnail, source, metadata, created_at FROM archive")

	var args []interface{}
	var conditions []string

	if filterByUploader != "" {
		// Simplified: uploader is part of source URL. Using LIKE for broader matching.
		// In a real scenario, this might be metadata.uploader or similar.
		conditions = append(conditions, "LOWER(source) LIKE LOWER(?)")
		args = append(args, "%"+filterByUploader+"%")
	}

	// Pagination condition
	if startRowId > 0 { // Assuming 0 or negative means no previous rowid cursor / fetch from beginning
		conditions = append(conditions, "rowid > ?")
		args = append(args, startRowId)
	}
	
	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	switch sortBy {
	case "title_asc":
		queryBuilder.WriteString(" ORDER BY title ASC")
	case "title_desc":
		queryBuilder.WriteString(" ORDER BY title DESC")
	case "date_asc":
		queryBuilder.WriteString(" ORDER BY created_at ASC")
	case "date_desc":
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	// TODO: Add duration sort later if duration is easily queryable (e.g., own column or JSON function)
	// For now, ensure metadata has a 'duration' field and that SQLite JSON functions are available/performant
	// or extract duration to its own column for efficient sorting.
	// Example for duration if metadata is JSON: ORDER BY json_extract(metadata, '$.duration') ASC
	default:
		// Default sort if sortBy is empty or unrecognized
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	}

	queryBuilder.WriteString(" LIMIT ?")
	args = append(args, limit)

	finalQuery := queryBuilder.String()
	slog.Debug("Executing archive list query", "query", finalQuery, "args", args)

	var entries []data.ArchiveEntry
	rows, err := conn.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		// var rowId int64 // To consume the selected rowid - REMOVED
		var entry data.ArchiveEntry
		if err := rows.Scan(
			&entry.RowId, // Scan directly into the struct field
			&entry.Id,
			&entry.Title,
			&entry.Path,
			&entry.Thumbnail,
			&entry.Source,
			&entry.Metadata,
			&entry.CreatedAt,
		); err != nil {
			// It's often better to return accumulated entries and the error
			return &entries, err
		}
		entries = append(entries, entry)
	}
	
	if err = rows.Err(); err != nil {
        return &entries, err
    }

	return &entries, nil
}

func (r *Repository) GetCursor(ctx context.Context, id string) (int64, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT rowid FROM archive WHERE id = ?", id)
	var rowId int64
	if err := row.Scan(&rowId); err != nil {
		return -1, err
	}
	return rowId, nil
}

func (r *Repository) IsSourceDownloaded(ctx context.Context, sourceURL string) (bool, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM archive WHERE source = ? LIMIT 1)"
	err = conn.QueryRowContext(ctx, query, sourceURL).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
