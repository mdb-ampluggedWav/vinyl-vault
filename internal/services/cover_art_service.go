package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var coverArtValidExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func (f *FileService) ValidateCoverArt(file *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !coverArtValidExtensions[ext] {
		return fmt.Errorf("unsupported image format: %s (allowed: jpg, png, webp)", ext)
	}

	if file.Size > f.maxCoverArtSize {
		maxMB := f.maxCoverArtSize / (1 << 20)
		return fmt.Errorf("image too large: maximum size is %dMB", maxMB)
	}
	return nil
}

func (f *FileService) SaveCoverArt(file *multipart.FileHeader, albumID uint64) (*FileUploadResult, error) {
	if albumID == 0 {
		return nil, fmt.Errorf("invalid album ID")
	}

	if err := f.ValidateCoverArt(file); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%d_%s%s", albumID, generateRandomString(8), ext)
	destPath := filepath.Join(f.coverArtDir, filename)

	if err := os.MkdirAll(f.coverArtDir, dirPermissions); err != nil {
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

func (f *FileService) DeleteCoverArt(filePath string) error {
	if filePath == "" {
		return nil
	}

	if err := f.ValidateFilePath(filePath); err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cover art: %w", err)
	}
	return nil
}
