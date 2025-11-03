package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"vinyl-vault/internal/config"
)

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

type FileUploadResult struct {
	Path, Filename string
	Size           uint64
}

type FileService struct {
	uploadDir, coverArtDir, audioDir  string
	maxAudioFileSize, maxCoverArtSize int64
}

func NewFileService(uploadDir, coverArtDir, audioDir string) *FileService {
	return &FileService{
		uploadDir:        uploadDir,
		coverArtDir:      coverArtDir,
		audioDir:         audioDir,
		maxAudioFileSize: 500 << 20, // 500mb
		maxCoverArtSize:  10 << 20,  // 10mb
	}
}

func NewFileServiceWithConfig(uploadDir, coverArtDir, audioDir string, cfg *config.Config) *FileService {
	return &FileService{
		uploadDir:        uploadDir,
		coverArtDir:      coverArtDir,
		audioDir:         audioDir,
		maxAudioFileSize: cfg.MaxAudioFileSize,
		maxCoverArtSize:  cfg.MaxCoverArtSize,
	}
}

func (f *FileService) GetRelativePath(fullPath string) (string, error) {
	if fullPath == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	realFullPath, err := filepath.EvalSymlinks(absFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	absUploadDir, err := filepath.Abs(f.uploadDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute upload dir: %w", err)
	}
	realUploadDir, err := filepath.EvalSymlinks(absUploadDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve upload dir: %w", err)
	}
	rel, err := filepath.Rel(realUploadDir, realFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path is outside upload dir")
	}
	return rel, nil
}

func (f *FileService) GetFullPath(relativePath string) string {
	return filepath.Join(f.uploadDir, relativePath)
}

func (f *FileService) GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

func (f *FileService) FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (f *FileService) EnsureDirectoriesExist() error {
	dirs := []string{f.uploadDir, f.coverArtDir, f.audioDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, dirPermissions); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (f *FileService) ValidateFilePath(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Resolve symlinks to ensure the real path stays within allowed roots
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	allowedDirs := []string{f.uploadDir, f.audioDir, f.coverArtDir}
	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		realDir, err := filepath.EvalSymlinks(absDir)
		if err != nil {
			continue
		}
		if rel, err := filepath.Rel(realDir, realPath); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return nil
		}
	}

	return fmt.Errorf("file path is outside allowed directories")
}

func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", " ", "_",
	)
	sanitized := replacer.Replace(name)

	const maxLen = 100
	if len(sanitized) > maxLen {
		runes := []rune(sanitized)
		if len(runes) > maxLen {
			sanitized = string(runes[:maxLen])
		}
	}
	return sanitized
}
