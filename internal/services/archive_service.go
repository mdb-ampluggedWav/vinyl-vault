package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxArchiveFiles = 100

func (f *FileService) ArchiveAudioFilesToZip(filePaths []string, zipName string) (string, error) {
	if len(filePaths) == 0 {
		return "", fmt.Errorf("no files to archive")
	}

	if len(filePaths) > maxArchiveFiles {
		return "", fmt.Errorf("too many files to archive (max %d)", maxArchiveFiles)
	}

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

	for _, filePath := range filePaths {
		if !f.FileExists(filePath) {
			err = os.Remove(zipPath)
			if err != nil {
				return "", err
			}
			return "", fmt.Errorf("file not found: %s", filePath)
		}
		if err = f.addFileToZip(zipWriter, filePath); err != nil {
			err = os.Remove(zipPath)
			if err != nil {
				return "", err
			}
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

	info, err := file.Stat()
	if err != nil {
		return err
	}

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

func (f *FileService) DeleteZipFile(zipPath string) error {
	if !strings.HasSuffix(strings.ToLower(zipPath), ".zip") {
		return fmt.Errorf("file is not a zip file")
	}
	return f.DeleteAudioFile(zipPath)
}
