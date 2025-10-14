package services

import (
	"context"
	"fmt"
	"time"
	"vinyl-vault/pkg"
)

type Album struct {
	ID        uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64       `json:"user_id" gorm:"not null;index"`
	Metadata  pkg.Metadata `json:"metadata" gorm:"embedded;embeddedPrefix:metadata_"`
	Tracks    []Track      `json:"tracks" gorm:"foreignKey:AlbumID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type AlbumRepository interface {
	FindByID(ctx context.Context, id uint64) (*Album, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*Album, error)
	FindByArtist(ctx context.Context, artist string) ([]*Album, error)
	Save(ctx context.Context, track *Album) error
	Delete(ctx context.Context, id uint64) error
}

type AlbumService struct {
	albumRepository AlbumRepository
	fileService     *FileService
}

func NewAlbumService(albumRepository AlbumRepository, fileService *FileService) *AlbumService {
	return &AlbumService{
		albumRepository: albumRepository,
		fileService:     fileService,
	}
}

func (a *AlbumService) CreateAlbum(ctx context.Context, userID uint64, metadata pkg.Metadata) (*Album, error) {
	album := &Album{
		UserID:   userID,
		Metadata: metadata,
	}
	if err := a.albumRepository.Save(ctx, album); err != nil {
		return nil, fmt.Errorf("failed to create album: %w", err)
	}
	return album, nil
}

func (a *AlbumService) GetAlbum(ctx context.Context, id uint64) (*Album, error) {

	album, err := a.albumRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}
	return album, nil
}

func (a *AlbumService) GetAlbumsByArtist(ctx context.Context, artist string) ([]*Album, error) {

	albums, err := a.albumRepository.FindByArtist(ctx, artist)
	if err != nil {
		return nil, fmt.Errorf("albums not found or empty: %w", err)
	}
	return albums, nil
}

func (a *AlbumService) GetAlbumsByUser(ctx context.Context, userID uint64) ([]*Album, error) {
	albums, err := a.albumRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user albums: %w", err)
	}
	return albums, nil
}

func (a *AlbumService) IsOwner(ctx context.Context, albumID, userID uint64) (bool, error) {
	album, err := a.albumRepository.FindByID(ctx, albumID)
	if err != nil {
		return false, err
	}
	return album.UserID == userID, nil
}

func (a *AlbumService) UpdateAlbumInfo(ctx context.Context, userID, albumID uint64, metadata pkg.Metadata) (*Album, error) {

	album, err := a.albumRepository.FindByID(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("album not found: %w", err)
	}

	if album.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this album")
	}

	// If updating cover art, delete old one
	if metadata.CoverArtPath != "" && metadata.CoverArtPath != album.Metadata.CoverArtPath {
		if album.Metadata.CoverArtPath != "" {
			a.fileService.DeleteCoverArt(album.Metadata.CoverArtPath)
		}
	}

	album.Metadata = metadata

	if err = a.albumRepository.Save(ctx, album); err != nil {
		return nil, fmt.Errorf("failed to update album's metadata: %w", err)
	}
	return album, nil
}

func (a *AlbumService) DeleteAlbum(ctx context.Context, userID, albumID uint64) error {

	album, err := a.albumRepository.FindByID(ctx, albumID)
	if err != nil {
		return fmt.Errorf("album not found: %w", err)
	}

	if album.UserID != userID {
		return fmt.Errorf("unauthorized: you don't own this album")
	}

	if album.Metadata.CoverArtPath != "" {
		a.fileService.DeleteCoverArt(album.Metadata.CoverArtPath)
	}

	if err = a.albumRepository.Delete(ctx, albumID); err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}
	return nil
}
