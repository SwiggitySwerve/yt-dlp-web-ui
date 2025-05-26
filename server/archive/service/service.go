package service

import (
	"context"
	// Ensure time is imported if used by domain.ArchiveEntry mapping (it's used by CreatedAt)
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/data" // For data.ArchiveEntry
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
)

type service struct { // Renamed from Service to service to match convention
	repository domain.Repository
}

func NewService(repo domain.Repository) domain.Service { // Renamed from New to NewService
	return &service{
		repository: repo,
	}
}

// List implements domain.Service.
func (s *service) List(ctx context.Context, startRowId int, limit int, sortBy string, filters map[string]string, searchQuery string) (*domain.PaginatedResponse[[]domain.ArchiveEntry], error) { // Signature updated
	archiveEntries, err := s.repository.List(ctx, startRowId, limit, sortBy, filters, searchQuery) // searchQuery passed
	if err != nil {
		return nil, err
	}
	
	respEntries := make([]domain.ArchiveEntry, len(*archiveEntries))
	var firstCursor, nextCursor int64

	if len(*archiveEntries) > 0 {
		firstCursor = (*archiveEntries)[0].RowId // data.ArchiveEntry has RowId

		if len(*archiveEntries) == limit {
			nextCursor = (*archiveEntries)[len(*archiveEntries)-1].RowId
		} else {
			nextCursor = 0 
		}
		
		for i, entry := range *archiveEntries { // entry is data.ArchiveEntry
			respEntries[i] = domain.ArchiveEntry{
				Id:        entry.Id,
				Title:     entry.Title,
				Path:      entry.Path,
				Thumbnail: entry.Thumbnail,
				Source:    entry.Source,
				Metadata:  entry.Metadata,
				CreatedAt: entry.CreatedAt, 
				Duration:  entry.Duration, // Mapping new field
				Format:    entry.Format,   // Mapping new field
			}
		}
	} else {
		firstCursor = int64(startRowId) 
        nextCursor = 0
	}

	return &domain.PaginatedResponse[[]domain.ArchiveEntry]{
		First: firstCursor,
		Next:  nextCursor,
		Data:  respEntries,
	}, nil
}

// Archive implements domain.Service.
func (s *service) Archive(ctx context.Context, entity *domain.ArchiveEntry) error {
	// Map domain.ArchiveEntry to data.ArchiveEntry
	dataEntry := &data.ArchiveEntry{
		Id:        entity.Id, 
		Title:     entity.Title,
		Path:      entity.Path,
		Thumbnail: entity.Thumbnail,
		Source:    entity.Source,
		Metadata:  entity.Metadata,
		CreatedAt: entity.CreatedAt,
        Duration:  entity.Duration, // Mapping new field
        Format:    entity.Format,   // Mapping new field
	}
	return s.repository.Archive(ctx, dataEntry)
}

// SoftDelete implements domain.Service.
func (s *service) SoftDelete(ctx context.Context, id string) (*domain.ArchiveEntry, error) {
	deletedEntry, err := s.repository.SoftDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if deletedEntry == nil {
		return nil, nil // Not found
	}
	return &domain.ArchiveEntry{ // Map data to domain
		Id:        deletedEntry.Id,
		Title:     deletedEntry.Title,
		Path:      deletedEntry.Path,
		Thumbnail: deletedEntry.Thumbnail,
		Source:    deletedEntry.Source,
		Metadata:  deletedEntry.Metadata,
		CreatedAt: deletedEntry.CreatedAt,
        Duration:  deletedEntry.Duration, // Mapping new field
        Format:    deletedEntry.Format,   // Mapping new field
	}, nil
}

// HardDelete implements domain.Service.
func (s *service) HardDelete(ctx context.Context, id string) (*domain.ArchiveEntry, error) {
	deletedEntry, err := s.repository.HardDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if deletedEntry == nil {
		return nil, nil // Not found
	}
	return &domain.ArchiveEntry{ // Map data to domain
		Id:        deletedEntry.Id,
		Title:     deletedEntry.Title,
		Path:      deletedEntry.Path,
		Thumbnail: deletedEntry.Thumbnail,
		Source:    deletedEntry.Source,
		Metadata:  deletedEntry.Metadata,
		CreatedAt: deletedEntry.CreatedAt,
        Duration:  deletedEntry.Duration, // Mapping new field
        Format:    deletedEntry.Format,   // Mapping new field
	}, nil
}


// GetCursor implements domain.Service.
func (s *service) GetCursor(ctx context.Context, id string) (int64, error) {
	return s.repository.GetCursor(ctx, id)
}

// IsSourceDownloaded is not part of domain.Service, it's on domain.Repository
// func (s *service) IsSourceDownloaded(ctx context.Context, sourceURL string) (bool, error) {
// 	return s.repository.IsSourceDownloaded(ctx, sourceURL)
// }
