package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AudioFormat string

const (
	FormatWAV  AudioFormat = "wav"
	FormatAIFF AudioFormat = "aiff"
	FormatFLAC AudioFormat = "flac"
	FormatALAC AudioFormat = "alac"
	FormatMP3  AudioFormat = "mp3"
	FormatOpus AudioFormat = "opus"
)

type ConversionService struct {
	tempDir string
}

func NewConversionService(tempDir string) *ConversionService {
	return &ConversionService{
		tempDir: tempDir,
	}
}

func (c *ConversionService) ConvertAudio(inputPath string, targetFormat AudioFormat) (string, error) {
	// Validate ffmpeg is available
	if err := c.validateFFmpeg(); err != nil {
		return "", err
	}

	// Validate target format
	if !c.isValidFormat(targetFormat) {
		return "", fmt.Errorf("unsupported format: %s", targetFormat)
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate output filename
	baseFilename := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	outputPath := filepath.Join(c.tempDir, fmt.Sprintf("%s_converted.%s", nameWithoutExt, c.getFileExtension(targetFormat)))

	// Build ffmpeg command based on target format
	args, err := c.buildFFmpegArgs(inputPath, outputPath, targetFormat)
	if err != nil {
		return "", err
	}

	// Execute conversion
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	// Verify output file was created
	if _, err = os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("conversion failed: output file not created")
	}

	return outputPath, nil
}

func (c *ConversionService) isValidFormat(format AudioFormat) bool {
	switch format {
	case FormatAIFF, FormatFLAC, FormatALAC, FormatMP3, FormatWAV, FormatOpus:
		return true
	default:
		return false
	}
}

func (c *ConversionService) GetSupportedFormats() []AudioFormat {
	return []AudioFormat{FormatAIFF, FormatFLAC, FormatALAC, FormatMP3, FormatWAV, FormatOpus}
}

func (c *ConversionService) getFileExtension(format AudioFormat) string {
	switch format {
	case FormatAIFF:
		return "aiff"
	case FormatFLAC:
		return "flac"
	case FormatALAC:
		return "m4a"
	case FormatMP3:
		return "mp3"
	case FormatWAV:
		return "wav"
	case FormatOpus:
		return "opus"
	default:
		return "bin"
	}
}

// CleanupTempFile removes and deletes temporary converted file
func (c *ConversionService) CleanupTempFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	absTempDir, err := filepath.Abs(c.tempDir)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absPath, absTempDir) {
		return fmt.Errorf("refusing to delete file outside temp directory")
	}

	if err = os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup temp file: %w", err)
	}
	return nil
}

func (c *ConversionService) buildFFmpegArgs(inputPath, outputPath string, format AudioFormat) ([]string, error) {
	args := []string{"-i", inputPath, "-y"}

	switch format {
	case FormatAIFF:
		args = append(args, "-acodec", "pcm_s16be", "-f", "aiff")
	case FormatWAV:
		args = append(args, "-acodec", "pcm_s16le", "-f", "wav")
	case FormatFLAC:
		args = append(args, "-acodec", "flac", "-compression_level", "5") // sweet spot for both speed and compression

	case FormatALAC:
		args = append(args, "-acodec", "alac", "-f", "mp4") // ALAC is typically in MP4 container
	case FormatMP3:
		args = append(args, "-acodec", "libmp3lame", "-b:a", "320k", "-ar", "44100") // 320k, 44.1kHz sample rate

	case FormatOpus:
		args = append(args, "-acodec", "libopus", "-b:a", "192k", "-vbr", "on") // High quality Opus (transparent quality) + Variable bitrate

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	args = append(args, outputPath)
	return args, nil
}

func (c *ConversionService) validateFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found in system PATH. Please install ffmpeg to enable audio conversion")
	}
	return nil
}
