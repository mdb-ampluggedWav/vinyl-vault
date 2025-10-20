package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"vinyl-vault/pkg"
)

type mockTrackRepository struct {
	tracks            map[uint64]*Track
	findByIDFunc      func(ctx context.Context, id uint64) (*Track, error)
	findByAlbumIDFunc func(ctx context.Context, albumID uint64) ([]*Track, error)
	saveFunc          func(ctx context.Context, track *Track) error
	deleteFunc        func(ctx context.Context, id uint64) error
}

func (m *mockTrackRepository) FindByID(ctx context.Context, id uint64) (*Track, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	track, exists := m.tracks[id]
	if !exists {
		return nil, errors.New("track not found")
	}
	return track, nil
}

func (m *mockTrackRepository) FindByAlbumID(ctx context.Context, albumID uint64) ([]*Track, error) {
	if m.findByAlbumIDFunc != nil {
		return m.findByAlbumIDFunc(ctx, albumID)
	}
	var tracks []*Track
	for _, track := range m.tracks {
		if track.AlbumID == albumID {
			tracks = append(tracks, track)
		}
	}
	return tracks, nil
}

func (m *mockTrackRepository) Save(ctx context.Context, track *Track) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, track)
	}
	if track.ID == 0 {
		track.ID = uint64(len(m.tracks) + 1)
	}
	m.tracks[track.ID] = track
	return nil
}

func (m *mockTrackRepository) Delete(ctx context.Context, id uint64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	delete(m.tracks, id)
	return nil
}

type mockAlbumRepository struct {
	albums       map[uint64]*Album
	findByIDFunc func(ctx context.Context, id uint64) (*Album, error)
}

func (m *mockAlbumRepository) FindByID(ctx context.Context, id uint64) (*Album, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	album, exists := m.albums[id]
	if !exists {
		return nil, errors.New("album not found")
	}
	return album, nil
}

func (m *mockAlbumRepository) FindByUserID(ctx context.Context, userID uint64) ([]*Album, error) {
	return nil, nil
}

func (m *mockAlbumRepository) FindByArtist(ctx context.Context, artist string) ([]*Album, error) {
	return nil, nil
}

func (m *mockAlbumRepository) Save(ctx context.Context, album *Album) error {
	return nil
}

func (m *mockAlbumRepository) Delete(ctx context.Context, id uint64) error {
	return nil
}

type mockFileDeleter struct {
	deleteAudioFileFunc func(filePath string) error
}

func (m *mockFileDeleter) DeleteAudioFile(filePath string) error {
	if m.deleteAudioFileFunc != nil {
		return m.deleteAudioFileFunc(filePath)
	}
	return nil
}

func TestTrackService_CreateTrack(t *testing.T) {

	tests := []struct {
		name        string
		userID      uint64
		albumID     uint64
		trackNumber int
		title       string
		duration    pkg.Duration
		filePath    string
		setupMocks  func(*mockTrackRepository, *mockAlbumRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful track creation",
			userID:      1,
			albumID:     1,
			trackNumber: 1,
			title:       "Test Track",
			duration:    pkg.Duration(180),
			filePath:    "/uploads/track.flac",
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1,
				}
			},
			wantErr: false,
		},

		{
			name:        "unauthorized - user doesn't own album",
			userID:      2,
			albumID:     1,
			trackNumber: 1,
			title:       "Test Track",
			duration:    pkg.Duration(180),
			filePath:    "/uploads/track.flac",
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1,
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},

		{
			name:        "album not found",
			userID:      1,
			albumID:     999,
			trackNumber: 1,
			title:       "Test Track",
			duration:    pkg.Duration(180),
			filePath:    "/uploads/track.flac",
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				// No album setup
			},
			wantErr:     true,
			errContains: "album not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			trackRepo := &mockTrackRepository{tracks: make(map[uint64]*Track)}
			albumRepo := &mockAlbumRepository{albums: make(map[uint64]*Album)}
			fileService := &mockFileDeleter{}

			if tt.setupMocks != nil {
				tt.setupMocks(trackRepo, albumRepo)
			}

			service := NewTrackService(trackRepo, albumRepo, fileService)

			track, err := service.CreateTrack(
				context.Background(),
				tt.userID,
				tt.albumID,
				tt.trackNumber,
				tt.title,
				tt.duration,
				tt.filePath,
				pkg.AudioQuality{},
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if track == nil {
					t.Errorf("expected track to be created")
					return
				}
				if track != nil && track.Title != tt.title {
					t.Errorf("expected title '%s', got '%s'", tt.title, track.Title)
				}
			}
		})
	}
}

func TestTrackService_GetTrack(t *testing.T) {
	trackRepo := &mockTrackRepository{
		tracks: map[uint64]*Track{
			1: {
				ID:          1,
				AlbumID:     1,
				TrackNumber: 1,
				Title:       "Test Track",
			},
		},
	}
	albumRepo := &mockAlbumRepository{albums: make(map[uint64]*Album)}
	fileService := &mockFileDeleter{}

	service := NewTrackService(trackRepo, albumRepo, fileService)

	t.Run("get existing track", func(t *testing.T) {
		track, err := service.GetTrack(context.Background(), 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if track.ID != 1 {
			t.Errorf("expected track ID 1, got %d", track.ID)
		}
	})

	t.Run("get non-existent track", func(t *testing.T) {
		_, err := service.GetTrack(context.Background(), 999)
		if err == nil {
			t.Errorf("expected error for non-existent track")
		}
	})
}

func TestTrackService_GetTracksByAlbum(t *testing.T) {
	tests := []struct {
		name        string
		albumID     uint64
		setupMocks  func(*mockTrackRepository, *mockAlbumRepository)
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:    "get tracks for album with multiple tracks",
			albumID: 1,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:          1,
					AlbumID:     1,
					TrackNumber: 1,
					Title:       "Track 1",
				}
				trackRepo.tracks[2] = &Track{
					ID:          2,
					AlbumID:     1,
					TrackNumber: 2,
					Title:       "Track 2",
				}
				trackRepo.tracks[3] = &Track{
					ID:          3,
					AlbumID:     1,
					TrackNumber: 3,
					Title:       "Track 3",
				}
				// Track from different album - should not be returned
				trackRepo.tracks[4] = &Track{
					ID:          4,
					AlbumID:     2,
					TrackNumber: 1,
					Title:       "Different Album Track",
				}
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "get tracks for album with one track",
			albumID: 2,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:          1,
					AlbumID:     2,
					TrackNumber: 1,
					Title:       "Single Track",
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:    "get tracks for album with no tracks",
			albumID: 3,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				// No tracks for album 3
				trackRepo.tracks[1] = &Track{
					ID:          1,
					AlbumID:     1,
					TrackNumber: 1,
					Title:       "Other Album Track",
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "tracks with same album ID but different track numbers",
			albumID: 5,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[10] = &Track{
					ID:          10,
					AlbumID:     5,
					TrackNumber: 3, // Out of order
					Title:       "Track 3",
				}
				trackRepo.tracks[11] = &Track{
					ID:          11,
					AlbumID:     5,
					TrackNumber: 1,
					Title:       "Track 1",
				}
				trackRepo.tracks[12] = &Track{
					ID:          12,
					AlbumID:     5,
					TrackNumber: 2,
					Title:       "Track 2",
				}
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "repository error",
			albumID: 999,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {

				trackRepo.findByAlbumIDFunc = func(ctx context.Context, albumID uint64) ([]*Track, error) {
					return nil, errors.New("database error")
				}
			},
			wantCount:   0,
			wantErr:     true,
			errContains: "failed to get tracks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackRepo := &mockTrackRepository{tracks: make(map[uint64]*Track)}
			albumRepo := &mockAlbumRepository{albums: make(map[uint64]*Album)}
			fileService := &mockFileDeleter{}

			if tt.setupMocks != nil {
				tt.setupMocks(trackRepo, albumRepo)
			}

			service := NewTrackService(trackRepo, albumRepo, fileService)

			tracks, err := service.GetTracksByAlbum(context.Background(), tt.albumID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(tracks) != tt.wantCount {
				t.Errorf("expected %d tracks, got %d", tt.wantCount, len(tracks))
			}

			for i, track := range tracks {
				if track.AlbumID != tt.albumID {
					t.Errorf("track %d has albumID %d, expected %d", i, track.AlbumID, tt.albumID)
				}
			}

			if tt.wantCount > 0 && len(tracks) > 0 {
				if tracks[0].Title == "" {
					t.Errorf("first track has empty title")
				}
			}
		})
	}
}

func TestTrackService_UpdateTrack(t *testing.T) {
	newTitle := "Updated Title"
	newTrackNumber := 2

	tests := []struct {
		name        string
		trackID     uint64
		userID      uint64
		title       *string
		trackNumber *int
		setupMocks  func(*mockTrackRepository, *mockAlbumRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful update",
			trackID:     1,
			userID:      1,
			title:       &newTitle,
			trackNumber: &newTrackNumber,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:          1,
					AlbumID:     1,
					TrackNumber: 1,
					Title:       "Old Title",
				}
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1,
				}
			},
			wantErr: false,
		},
		{
			name:    "unauthorized update",
			trackID: 1,
			userID:  2,
			title:   &newTitle,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:      1,
					AlbumID: 1,
				}
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1, // Different user, unauthorized
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackRepo := &mockTrackRepository{tracks: make(map[uint64]*Track)}
			albumRepo := &mockAlbumRepository{albums: make(map[uint64]*Album)}
			fileService := &mockFileDeleter{}

			tt.setupMocks(trackRepo, albumRepo)

			service := NewTrackService(trackRepo, albumRepo, fileService)

			track, err := service.UpdateTrack(
				context.Background(),
				tt.userID,
				tt.trackID,
				tt.trackNumber,
				tt.title,
				nil,
				nil,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tt.title != nil && track.Title != *tt.title {
					t.Errorf("expected title '%s', got '%s'", *tt.title, track.Title)
				}
			}
		})
	}
}

func TestTrackService_DeleteTrack(t *testing.T) {
	tests := []struct {
		name        string
		trackID     uint64
		userID      uint64
		setupMocks  func(*mockTrackRepository, *mockAlbumRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful delete",
			trackID: 1,
			userID:  1,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:      1,
					AlbumID: 1,
				}
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1,
				}
			},
			wantErr: false,
		},
		{
			name:    "unauthorized delete",
			trackID: 1,
			userID:  2,
			setupMocks: func(trackRepo *mockTrackRepository, albumRepo *mockAlbumRepository) {
				trackRepo.tracks[1] = &Track{
					ID:      1,
					AlbumID: 1,
				}
				albumRepo.albums[1] = &Album{
					ID:     1,
					UserID: 1,
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackRepo := &mockTrackRepository{tracks: make(map[uint64]*Track)}
			albumRepo := &mockAlbumRepository{albums: make(map[uint64]*Album)}
			fileService := &mockFileDeleter{}

			tt.setupMocks(trackRepo, albumRepo)

			service := NewTrackService(trackRepo, albumRepo, fileService)

			err := service.DeleteTrack(context.Background(), tt.userID, tt.trackID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
