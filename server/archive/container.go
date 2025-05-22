package archive

import (
	"database/sql"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
)

func Container(db *sql.DB) (domain.RestHandler, domain.Service, domain.Repository) { // Signature changed
	var (
		r = provideRepository(db) // r is domain.Repository
		s = provideService(r)
		h = provideHandler(s)
	)
	return h, s, r // r added to return
}
