package service

import (
	"context"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/data"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/archive/domain"
)

type Service struct {
	repository domain.Repository
}

func New(repository domain.Repository) domain.Service {
	return &Service{
		repository: repository,
	}
}

// Archive implements domain.Service.
func (s *Service) Archive(ctx context.Context, entity *domain.ArchiveEntry) error {
	return s.repository.Archive(ctx, &data.ArchiveEntry{
		// RowId is not set here as it's DB-generated
		Id:        entity.Id,
		Title:     entity.Title,
		Path:      entity.Path,
		Thumbnail: entity.Thumbnail,
		Source:    entity.Source,
		Metadata:  entity.Metadata,
		CreatedAt: entity.CreatedAt,
	})
}

// HardDelete implements domain.Service.
func (s *Service) HardDelete(ctx context.Context, id string) (*domain.ArchiveEntry, error) {
	res, err := s.repository.HardDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil { // Handle case where entry might not be found
		return nil, nil 
	}
	return &domain.ArchiveEntry{
		Id:        res.Id,
		Title:     res.Title,
		Path:      res.Path,
		Thumbnail: res.Thumbnail,
		Source:    res.Source,
		Metadata:  res.Metadata,
		CreatedAt: res.CreatedAt,
	}, nil
}

// SoftDelete implements domain.Service.
func (s *Service) SoftDelete(ctx context.Context, id string) (*domain.ArchiveEntry, error) {
	res, err := s.repository.SoftDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil { // Handle case where entry might not be found
		return nil, nil
	}
	return &domain.ArchiveEntry{
		Id:        res.Id,
		Title:     res.Title,
		Path:      res.Path,
		Thumbnail: res.Thumbnail,
		Source:    res.Source,
		Metadata:  res.Metadata,
		CreatedAt: res.CreatedAt,
	}, nil
}

// List implements domain.Service.
func (s *Service) List(
	ctx context.Context,
	startRowId int,
	limit int,
	sortBy string, 
	filterByUploader string,
) (*domain.PaginatedResponse[[]domain.ArchiveEntry], error) {
	// Call repository's updated List method
	archiveEntries, err := s.repository.List(ctx, startRowId, limit, sortBy, filterByUploader)
	if err != nil {
		return nil, err
	}

	respEntries := make([]domain.ArchiveEntry, len(*archiveEntries))
	var firstCursor, nextCursor int64

	if len(*archiveEntries) > 0 {
		firstCursor = (*archiveEntries)[0].RowId // Use RowId from data.ArchiveEntry

		// If the number of entries returned is equal to the limit,
		// it's likely there are more entries. Set nextCursor to the RowId of the last entry.
		if len(*archiveEntries) == limit {
			nextCursor = (*archiveEntries)[len(*archiveEntries)-1].RowId
		} else {
			nextCursor = 0 // Or a clear indicator that there are no more pages
		}
		
		for i, entry := range *archiveEntries {
			respEntries[i] = domain.ArchiveEntry{ // Map data.ArchiveEntry to domain.ArchiveEntry
				Id:        entry.Id,
				Title:     entry.Title,
				Path:      entry.Path,
				Thumbnail: entry.Thumbnail,
				Source:    entry.Source,
				Metadata:  entry.Metadata,
				CreatedAt: entry.CreatedAt,
				// RowId is not part of domain.ArchiveEntry, so not mapped here for response to client
			}
		}
	} else {
		// No entries found
		// firstCursor and nextCursor will remain 0, which is fine.
		// Alternatively, could set firstCursor to startRowId if that's meaningful for an empty result.
	}

	return &domain.PaginatedResponse[[]domain.ArchiveEntry]{
		First: firstCursor,
		Next:  nextCursor,
		Data:  respEntries,
	}, nil
}

// GetCursor implements domain.Service.
func (s *Service) GetCursor(ctx context.Context, id string) (int64, error) {
	return s.repository.GetCursor(ctx, id)
}
