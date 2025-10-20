package services

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxAudioFileSize = int64(500 << 20) //   500mb
	maxCoverArtSize  = int64(10 << 20)  //   10mb for flexibility

	dirPermissions  = 0755
	filePermissions = 0644

	maxArchiveFiles = 100
)

var (
	audioFileValidExtensions = map[string]bool{
		".wav":  true,
		".aiff": true,
		".flac": true,
		".alac": true,
		".mp3":  true,
		".m4a":  true,
		".opus": true,
	}
	coverArtValidExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
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

//*********************************************** AUDIO FILE ********************************************************

func (f *FileService) SaveTrackAudioFile(
	file *multipart.FileHeader, albumID uint64, trackNumber int, trackTitle string,
) (*FileUploadResult, error) {

	if trackTitle == "" {
		return nil, fmt.Errorf("track title cannot be empty")
	}

	if err := f.ValidateAudioFile(file); err != nil {
		return nil, err
	}

	// get OG extension
	ext := f.GetAudioFileExtension(file.Filename)

	safeTitle := SanitizeFilename(trackTitle)
	filename := fmt.Sprintf("%d_%02d_%s%s", albumID, trackNumber, safeTitle, ext)
	destPath := filepath.Join(f.audioDir, filename)

	if err := os.MkdirAll(f.audioDir, dirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// open
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Set file permissions
	if err = dst.Chmod(filePermissions); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Copy file
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

func (f *FileService) SaveAudioFile(file *multipart.FileHeader, destPath string) error {

	// check if valid
	if err := f.ValidateAudioFile(file); err != nil {
		return err
	}

	// ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// open uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// create dest. file
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// set file permissions
	if err = dst.Chmod(filePermissions); err != nil {
		os.Remove(destPath)
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// copy file to destination
	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(destPath) // cleanup
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}

func (f *FileService) DeleteAudioFile(filePath string) error {
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
	if file.Size > maxAudioFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", file.Size, maxAudioFileSize)
	}
	return nil
}

//*********************************************** ZIP FILE ***********************************************************

// ArchiveAudioFilesToZip creates a zip archive from multiple audio files
// Returns the path to the created zip file
func (f *FileService) ArchiveAudioFilesToZip(filePaths []string, zipName string) (string, error) {

	if len(filePaths) == 0 {
		return "", fmt.Errorf("no files to archive")
	}

	if len(filePaths) > maxArchiveFiles {
		return "", fmt.Errorf("too many files to archive (max %d)", maxArchiveFiles)
	}

	// validate zip name
	if !strings.HasSuffix(strings.ToLower(zipName), ".zip") {
		zipName += ".zip"
	}

	zipPath := filepath.Join(f.uploadDir, zipName)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// add each file to zip
	for _, filePath := range filePaths {
		if !f.FileExists(filePath) {
			os.Remove(zipPath)
			return "", fmt.Errorf("file not found: %s", filePath)
		}
		if err = f.addFileToZip(zipWriter, filePath); err != nil {
			os.Remove(zipPath) // clean up if error
			return "", fmt.Errorf("failed to add file %s to zip: %w", filePath, err)
		}
	}

	return zipPath, nil
}

func (f *FileService) addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// file info for header
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// create zip header with JUST filename
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

//*************************************************COVER ART*********************************************************

func (f *FileService) ValidateCoverArt(file *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !coverArtValidExtensions[ext] {
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	if file.Size > maxCoverArtSize {
		return fmt.Errorf("image too large: %d bytes (max %d)", file.Size, maxCoverArtSize)
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

	// generate unique filename
	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%d_%s%s", albumID, generateRandomString(8), ext)
	destPath := filepath.Join(f.coverArtDir, filename)

	// ensure dir exists
	if err := os.MkdirAll(f.coverArtDir, dirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// create dest file
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// set file permissions
	if err = dst.Chmod(filePermissions); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to set file permissions: %w", err)
	}

	// copy file
	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(destPath) // cleanup
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

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cover art: %w", err)
	}
	return nil
}

//************************************************** UTIL *************************************************************

// GetRelativePath -> returns path relative to upload dir (for storing in DB)
func (f *FileService) GetRelativePath(fullPath string) (string, error) {
	if fullPath == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	absUploadDir, err := filepath.Abs(f.uploadDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute upload dir: %w", err)
	}

	rel, err := filepath.Rel(absUploadDir, absFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
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

func (f *FileService) DeleteZipFile(zipPath string) error {
	if !strings.HasSuffix(strings.ToLower(zipPath), ".zip") {
		return fmt.Errorf("file is not a zip file")
	}
	return f.DeleteAudioFile(zipPath)
}

func (f *FileService) GetAudioFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

func (f *FileService) IsValidAudioExtension(ext string) bool {
	return audioFileValidExtensions[strings.ToLower(ext)]
}

func (f *FileService) ValidateFilePath(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if path is within allowed directories
	allowedDirs := []string{
		f.uploadDir,
		f.audioDir,
		f.coverArtDir,
	}

	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absDir) {
			return nil
		}
	}

	return fmt.Errorf("file path is outside allowed directories")
}

// generateRandomString creates a random hex string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
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
