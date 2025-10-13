package services

import (
	"context"
	"fmt"
	"time"

	"vinyl-vault/pkg"
)

type Track struct {
	ID           uint64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AlbumID      uint64           `json:"album_id" gorm:"not null"`
	TrackNumber  int              `json:"track_number" gorm:"not null"`
	Title        string           `json:"title" gorm:"not null"`
	Duration     int              `json:"duration"`
	FilePath     string           `json:"file_path" gorm:"not null"`
	AudioQuality pkg.AudioQuality `json:"audio_quality" gorm:"embedded;embeddedPrefix:audio_"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type TrackRepository interface {
	FindByID(ctx context.Context, id uint64) (*Track, error)
	FindByAlbumID(ctx context.Context, albumID uint64) ([]*Track, error)
	Save(ctx context.Context, track *Track) error
	Delete(ctx context.Context, id uint64) error
}

type TrackService struct {
	trackRepository TrackRepository
	albumRepository AlbumRepository
}

func NewTrackService(trackRepository TrackRepository, albumRepository AlbumRepository) *TrackService {
	return &TrackService{
		trackRepository: trackRepository,
		albumRepository: albumRepository,
	}
}

func (t *TrackService) CreateTrack(
	ctx context.Context, userID, albumID uint64, trackNumber int, title string,
	duration int, filePath string, audioQuality pkg.AudioQuality) (*Track, error) {

	album, err := t.albumRepository.FindByID(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("album not found: %w", err)
	}
	if album.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this album")
	}

	track := &Track{
		AlbumID:      albumID,
		TrackNumber:  trackNumber,
		Title:        title,
		Duration:     duration,
		FilePath:     filePath,
		AudioQuality: audioQuality,
	}
	if err := t.trackRepository.Save(ctx, track); err != nil {
		return nil, fmt.Errorf("failed to create track: %w", err)
	}
	return track, nil
}

func (t *TrackService) GetTrack(ctx context.Context, id uint64) (*Track, error) {

	track, err := t.trackRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}
	return track, nil
}

func (t *TrackService) GetTracksByAlbum(ctx context.Context, albumID uint64) ([]*Track, error) {
	tracks, err := t.trackRepository.FindByAlbumID(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks: %w", err)
	}
	return tracks, nil
}

func (t *TrackService) UpdateTrack(
	ctx context.Context, userID, trackID uint64,
	trackNumber *int, title *string, duration *int, audioQuality *pkg.AudioQuality,
) (*Track, error) {

	track, err := t.trackRepository.FindByID(ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	album, err := t.albumRepository.FindByID(ctx, track.AlbumID)
	if err != nil {
		return nil, fmt.Errorf("album not found: %w", err)
	}
	if album.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this track's album")
	}

	if trackNumber != nil {
		track.TrackNumber = *trackNumber
	}

	if title != nil {
		track.Title = *title
	}

	if duration != nil {
		track.Duration = *duration
	}
	if audioQuality != nil {
		track.AudioQuality = *audioQuality
	}

	if err = t.trackRepository.Save(ctx, track); err != nil {
		return nil, fmt.Errorf("failed to update track: %w", err)
	}
	return track, nil
}

func (t *TrackService) DeleteTrack(ctx context.Context, userID, trackID uint64) error {

	track, err := t.trackRepository.FindByID(ctx, trackID)
	if err != nil {
		return fmt.Errorf("track not found: %w", err)
	}

	album, err := t.albumRepository.FindByID(ctx, track.AlbumID)
	if err != nil {
		return fmt.Errorf("album not found: %w", err)
	}
	if album.UserID != userID {
		return fmt.Errorf("unauthorized: you don't own this track's album")
	}

	if err := t.trackRepository.Delete(ctx, trackID); err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}
	return nil
}
