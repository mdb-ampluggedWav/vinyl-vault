package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var audioFileValidExtensions = map[string]bool{
	".wav":  true,
	".aiff": true,
	".flac": true,
	".alac": true,
	".mp3":  true,
	".m4a":  true,
	".opus": true,
}

func (f *FileService) SaveTrackAudioFile(
	file *multipart.FileHeader, albumID uint64, trackNumber int, trackTitle string,
) (*FileUploadResult, error) {
	if trackTitle == "" {
		return nil, fmt.Errorf("track title cannot be empty")
	}

	if err := f.ValidateAudioFile(file); err != nil {
		return nil, err
	}

	ext := f.GetAudioFileExtension(file.Filename)
	safeTitle := SanitizeFilename(trackTitle)
	filename := fmt.Sprintf("%d_%02d_%s%s", albumID, trackNumber, safeTitle, ext)
	destPath := filepath.Join(f.audioDir, filename)

	if err := os.MkdirAll(f.audioDir, dirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if err = dst.Chmod(filePermissions); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to set file permissions: %w", err)
	}

	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &FileUploadResult{
		Path:     destPath,
		Filename: filename,
		Size:     uint64(file.Size),
	}, nil
}

func (f *FileService) DeleteAudioFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	if err := f.ValidateFilePath(filePath); err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (f *FileService) ValidateAudioFile(file *multipart.FileHeader) error {
	ext := f.GetAudioFileExtension(file.Filename)
	if !f.IsValidAudioExtension(ext) {
		return fmt.Errorf("unsupported audio format: %s", ext)
	}
	if file.Size > f.maxAudioFileSize {
		maxMB := f.maxAudioFileSize / (1 << 20)
		return fmt.Errorf("file too large: maximum size is %dMB", maxMB)
	}
	return nil
}

func (f *FileService) GetAudioFileExtension(filename string) string {
	return filepath.Ext(filename)
}

func (f *FileService) IsValidAudioExtension(ext string) bool {
	return audioFileValidExtensions[ext]
}
