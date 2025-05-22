package subscription

import (
	"database/sql"
	"sync"

	archiveDomain "github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain" // Added import
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/domain"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/repository"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/rest"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/service"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/subscription/task"
)

var (
	repo domain.Repository
	svc  domain.Service
	hand domain.RestHandler

	repoOnce sync.Once
	svcOnce  sync.Once
	handOnce sync.Once
)

func provideRepository(db *sql.DB) domain.Repository {
	repoOnce.Do(func() {
		repo = repository.New(db)
	})
	return repo
}

func provideService(r domain.Repository, runner task.TaskRunner, archiveRepo archiveDomain.Repository) domain.Service { // Signature changed
	svcOnce.Do(func() {
		// Order of args for service.New: subscriptionRepo, archiveRepo, taskRunner
		svc = service.New(r, archiveRepo, runner) // archiveRepo passed here
	})
	return svc
}

func provideHandler(s domain.Service) domain.RestHandler {
	handOnce.Do(func() {
		hand = rest.New(s)
	})
	return hand
}
