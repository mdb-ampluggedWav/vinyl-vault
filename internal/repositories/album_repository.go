package repositories

import (
	"context"
	"errors"
	"fmt"
	"vinyl-vault/internal/services"

	"gorm.io/gorm"
)

type GormAlbumRepository struct {
	db *gorm.DB
}

func NewGormAlbumRepository(db *gorm.DB) services.AlbumRepository {
	return &GormAlbumRepository{
		db: db,
	}
}

func (r *GormAlbumRepository) FindByUserID(ctx context.Context, userID uint64) ([]*services.Album, error) {

	var albums []*services.Album

	if result := r.db.WithContext(ctx).Preload("Tracks").Where("user_id = ?", userID).Find(&albums); result.Error != nil {
		return nil, fmt.Errorf("failed to find albums: %w", result.Error)
	}
	return albums, nil
}

func (r *GormAlbumRepository) FindByID(ctx context.Context, id uint64) (*services.Album, error) {
	var album services.Album

	result := r.db.WithContext(ctx).Preload("Tracks").First(&album, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("album with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to find album: %w", result.Error)
	}
	return &album, nil
}

func (r *GormAlbumRepository) FindByArtist(ctx context.Context, artist string) ([]*services.Album, error) {

	var albums []*services.Album

	result := r.db.WithContext(ctx).Where("metadata_artist = ?", artist).Find(&albums)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find albums: %w", result.Error)
	}
	return albums, nil
}

func (r *GormAlbumRepository) Save(ctx context.Context, album *services.Album) error {

	result := r.db.WithContext(ctx).Save(album)
	if result.Error != nil {
		return fmt.Errorf("failed to save album: %w", result.Error)
	}
	return nil
}

func (r *GormAlbumRepository) Delete(ctx context.Context, id uint64) error {

	result := r.db.WithContext(ctx).Delete(&services.Album{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete album: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("album with id %d not found", id)
	}
	return nil
}
