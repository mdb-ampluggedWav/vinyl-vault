package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxAudioFileSize = int64(500 << 20)
	maxCoverArtSize  = int64(500 << 11)
)

type FileUploadResult struct {
	Path, Filename string
	Size           uint64
}

type FileValidationError struct {
	Field, Message string
}

type FileService struct {
	uploadDir, coverArtDir, audioDir string
}

func NewFileService(uploadDir, coverArtDir, audioDir string) *FileService {
	return &FileService{
		uploadDir:   uploadDir,
		coverArtDir: coverArtDir,
		audioDir:    audioDir,
	}
}

//***********************************************
//************ AUDIO FILE operations*************
//***********************************************

func (f *FileService) DeleteAudioFile(filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (f *FileService) ValidateAudioFile(file *multipart.FileHeader) error {

	ext := strings.ToLower(filepath.Ext(file.Filename))

	validExts := map[string]bool{
		".wav":  true,
		".aiff": true,
		".flac": true,
		".alac": true,
		".mp3":  true,
		".m4a":  true,
	}

	if !validExts[ext] {
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	if file.Size > maxAudioFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", file.Size, maxAudioFileSize)
	}

	return nil
}

// COVER ART operations
