package repository

import (
	"context"
	"database/sql"
	"os"
	"strconv" // Added for Atoi
	"strings" 
	"log/slog"  

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
		"INSERT INTO archive (id, title, path, thumbnail, source, metadata, created_at, duration, format, uploader) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.NewString(),
		entry.Title,
		entry.Path,
		entry.Thumbnail,
		entry.Source,
		entry.Metadata,
		entry.CreatedAt,
		entry.Duration,
		entry.Format,
		entry.Uploader, // Added entry.Uploader
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
	row := tx.QueryRowContext(ctx, "SELECT id, title, path, thumbnail, source, metadata, created_at, duration, format, uploader FROM archive WHERE id = ?", id) // Added uploader
	if err := row.Scan(
		&model.Id,
		&model.Title,
		&model.Path,
		&model.Thumbnail,
		&model.Source,
		&model.Metadata,
		&model.CreatedAt,
		&model.Duration,
		&model.Format,
		&model.Uploader, // Added &model.Uploader
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
	if entry == nil { 
		return nil, sql.ErrNoRows 
	}
	if err := os.Remove(entry.Path); err != nil {
		return entry, err 
	}
	return entry, nil
}

func (r *Repository) List(ctx context.Context, startRowId int, limit int, sortBy string, filters map[string]string, searchQuery string) (*[]data.ArchiveEntry, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	
	var finalQuerySb strings.Builder
	var args []interface{}
	var conditions []string // Used for non-FTS and FTS filter parts

	// Build ftsMatcherArg
	titleSearchTerm := searchQuery
	tagsSearchTerm, hasTagsFilter := filters["filter_tags"]
	var ftsMatcherArg string

	if titleSearchTerm != "" && hasTagsFilter && tagsSearchTerm != "" {
		ftsMatcherArg = "title:(" + titleSearchTerm + ") tags_fts:(" + tagsSearchTerm + ")"
	} else if titleSearchTerm != "" {
		ftsMatcherArg = "title:(" + titleSearchTerm + ")"
	} else if hasTagsFilter && tagsSearchTerm != "" {
		ftsMatcherArg = "tags_fts:(" + tagsSearchTerm + ")"
	}

	if ftsMatcherArg != "" {
		// FTS Search Path
		ftsSubQuery := "SELECT rowid FROM archive_fts WHERE archive_fts MATCH ?"
		args = append(args, ftsMatcherArg) // Use ftsMatcherArg for the MATCH clause

		finalQuerySb.WriteString("SELECT r.rowid, r.id, r.title, r.path, r.thumbnail, r.source, r.metadata, r.created_at, r.duration, r.format, r.uploader ")
		finalQuerySb.WriteString("FROM archive r JOIN (")
		finalQuerySb.WriteString(ftsSubQuery)
		finalQuerySb.WriteString(") fts_matches ON r.rowid = fts_matches.rowid ")

		// Apply other filters from the 'filters' map to the main query
		// Note: filter_tags is already handled by ftsMatcherArg, so it should be skipped here.
		for key, value := range filters {
			if value == "" { continue }
			if key == "filter_tags" { continue } // Skip filter_tags as it's in ftsMatcherArg

			switch key {
			case "uploader":
				conditions = append(conditions, "LOWER(r.uploader) LIKE ?")
				args = append(args, "%"+strings.ToLower(value)+"%")
			case "format":
				conditions = append(conditions, "LOWER(r.format) = ?")
				args = append(args, strings.ToLower(value))
			case "min_duration":
				if dur, errConv := strconv.Atoi(value); errConv == nil && dur >= 0 {
					conditions = append(conditions, "r.duration >= ?")
					args = append(args, dur)
				} else {
                    slog.Warn("Invalid min_duration filter value, skipping", "value", value, "error", errConv)
                }
			case "max_duration":
				if dur, errConv := strconv.Atoi(value); errConv == nil && dur >= 0 {
					conditions = append(conditions, "r.duration <= ?")
					args = append(args, dur)
				} else {
                    slog.Warn("Invalid max_duration filter value, skipping", "value", value, "error", errConv)
                }
			}
		}

		if len(conditions) > 0 {
			finalQuerySb.WriteString("WHERE ")
			finalQuerySb.WriteString(strings.Join(conditions, " AND "))
		}
		
		// Pagination (startRowId) for FTS results
		if startRowId > 0 {
			if len(conditions) > 0 {
				finalQuerySb.WriteString(" AND ")
			} else {
				finalQuerySb.WriteString("WHERE ")
			}
			finalQuerySb.WriteString("r.rowid > ?") // Apply to the joined table's rowid
			args = append(args, startRowId)
		}

	} else {
		// Non-FTS Path (existing logic)
		// This path is taken if ftsMatcherArg is empty (neither title nor tags search)
		finalQuerySb.WriteString("SELECT rowid, id, title, path, thumbnail, source, metadata, created_at, duration, format, uploader FROM archive ")
		
		// Apply filters. filter_tags would be ignored here correctly as ftsMatcherArg is empty.
		for key, value := range filters {
			if value == "" { continue }
			// if key == "filter_tags" { continue; } // Technically not needed here as ftsMatcherArg would be empty
			switch key {
			case "uploader":
				conditions = append(conditions, "LOWER(uploader) LIKE ?")
				args = append(args, "%"+strings.ToLower(value)+"%")
			case "format":
				conditions = append(conditions, "LOWER(format) = ?")
				args = append(args, strings.ToLower(value))
			case "min_duration":
				if dur, errConv := strconv.Atoi(value); errConv == nil && dur >= 0 {
					conditions = append(conditions, "duration >= ?")
					args = append(args, dur)
				} else {
                     slog.Warn("Invalid min_duration filter value, skipping", "value", value, "error", errConv)
                }
			case "max_duration":
				if dur, errConv := strconv.Atoi(value); errConv == nil && dur >= 0 {
					conditions = append(conditions, "duration <= ?")
					args = append(args, dur)
				} else {
                    slog.Warn("Invalid max_duration filter value, skipping", "value", value, "error", errConv)
                }
			}
		}

		var whereClause strings.Builder 
		if len(conditions) > 0 {
			whereClause.WriteString("WHERE ")
			whereClause.WriteString(strings.Join(conditions, " AND "))
		}

		if startRowId > 0 {
			if len(conditions) > 0 { // Check if conditions already added something to whereClause
				whereClause.WriteString(" AND ")
			} else {
				whereClause.WriteString("WHERE ") 
			}
			whereClause.WriteString("rowid > ?")
			args = append(args, startRowId)
		}
		finalQuerySb.WriteString(whereClause.String()) 
	}

	// Common Sorting and Limit for both FTS and Non-FTS paths
	orderByClause := "ORDER BY created_at DESC" // Default
	switch sortBy {
	case "title_asc": orderByClause = "ORDER BY title ASC"
	case "title_desc": orderByClause = "ORDER BY title DESC"
	case "date_asc": orderByClause = "ORDER BY created_at ASC"
	case "date_desc": orderByClause = "ORDER BY created_at DESC"
	case "duration_asc": orderByClause = "ORDER BY duration ASC, created_at DESC" // Added secondary sort
	case "duration_desc": orderByClause = "ORDER BY duration DESC, created_at DESC" // Added secondary sort
	// For FTS, sorting by relevance (rank) is often desired if searchQuery is present.
    // Example: if searchQuery != "" { orderByClause = "ORDER BY rank" } // rank is an FTS5 virtual column
    // This requires selecting rank from the FTS subquery and joining it.
    // For this iteration, we use standard column sorting for both paths.
	}
	finalQuerySb.WriteString(" ")
	finalQuerySb.WriteString(orderByClause)

	finalQuerySb.WriteString(" LIMIT ?")
	args = append(args, limit)

	finalQueryString := finalQuerySb.String()
	slog.Debug("Executing archive list query", "query", finalQueryString, "args", args, "searchQuery", searchQuery)

	var entries []data.ArchiveEntry
	rows, err := conn.QueryContext(ctx, finalQueryString, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry data.ArchiveEntry
		if err := rows.Scan(
			&entry.RowId, 
			&entry.Id,
			&entry.Title,
			&entry.Path,
			&entry.Thumbnail,
			&entry.Source,
			&entry.Metadata,
			&entry.CreatedAt,
			&entry.Duration,
			&entry.Format,
			&entry.Uploader, // Added &entry.Uploader
		); err != nil {
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
