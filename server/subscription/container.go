package subscription

import (
	"database/sql"

	archiveDomain "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain" // Added import
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/task"
)

func Container(db *sql.DB, runner task.TaskRunner, archiveRepo archiveDomain.Repository) domain.RestHandler { // Signature changed
	var (
		r = provideRepository(db)
		s = provideService(r, runner, archiveRepo) // archiveRepo passed here
		h = provideHandler(s)
	)
	return h
}
