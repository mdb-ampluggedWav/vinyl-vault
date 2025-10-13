package repositories

import (
	"context"
	"errors"
	"fmt"

	"vinyl-vault/internal/services"

	"gorm.io/gorm"
)

type GormTrackRepository struct {
	db *gorm.DB
}

func NewGormTrackRepository(db *gorm.DB) services.TrackRepository {
	return &GormTrackRepository{
		db: db,
	}
}

func (r *GormTrackRepository) FindByID(ctx context.Context, id uint64) (*services.Track, error) {
	var track services.Track

	result := r.db.WithContext(ctx).First(&track, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("track with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to find track: %w", result.Error)
	}
	return &track, nil
}

func (r *GormTrackRepository) FindByAlbumID(ctx context.Context, albumID uint64) ([]*services.Track, error) {
	var tracks []*services.Track

	result := r.db.WithContext(ctx).Where("album_id = ?", albumID).Find(&tracks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find tracks: %w", result.Error)
	}
	return tracks, nil
}

func (r *GormTrackRepository) Save(ctx context.Context, track *services.Track) error {

	result := r.db.WithContext(ctx).Save(track)
	if result.Error != nil {
		return fmt.Errorf("failed to save track: %w", result.Error)
	}
	return nil
}

func (r *GormTrackRepository) Delete(ctx context.Context, id uint64) error {

	result := r.db.WithContext(ctx).Delete(&services.Track{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete track: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("track with id %d not found", id)
	}
	return nil
}
