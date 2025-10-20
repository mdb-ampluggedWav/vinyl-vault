package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"vinyl-vault/internal/services"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService  *services.FileService
	trackService *services.TrackService
	albumService *services.AlbumService
}

func NewFileHandler(
	fileService *services.FileService,
	trackService *services.TrackService,
	albumService *services.AlbumService,
) *FileHandler {
	return &FileHandler{
		fileService:  fileService,
		trackService: trackService,
		albumService: albumService,
	}
}

func (h *FileHandler) RegisterFileRoutes(router *gin.RouterGroup) {
	router.GET("/track/:id/stream", h.StreamTrack)
	router.GET("/track/:id/download", h.DownloadTrack)
	router.GET("/album/:id/cover", h.ServeCoverArt)
}

func (h *FileHandler) StreamTrack(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	trackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid track id"})
		return
	}

	// Get track
	track, err := h.trackService.GetTrack(c.Request.Context(), uint64(trackID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		return
	}

	// Get album
	album, err := h.albumService.GetAlbum(c.Request.Context(), track.AlbumID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	_ = userID
	_ = album

	fullPath := h.fileService.GetFullPath(track.FilePath)

	if !h.fileService.FileExists(fullPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio file not found"})
		return
	}

	// Get file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to access file"})
		return
	}

	// Set headers for streaming
	c.Header("Content-Type", getContentType(fullPath))
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "no-cache")

	// Handle range requests for seeking
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// Serve partial content for range requests
		c.File(fullPath)
		return
	}

	// Serve full file
	c.File(fullPath)
}

func (h *FileHandler) DownloadTrack(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	trackID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid track id"})
		return
	}

	// Get track
	track, err := h.trackService.GetTrack(c.Request.Context(), uint64(trackID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		return
	}

	// Get album to verify access
	album, err := h.albumService.GetAlbum(c.Request.Context(), track.AlbumID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	// Check access (allowing all authenticated users per project goal)
	_ = userID
	_ = album

	fullPath := h.fileService.GetFullPath(track.FilePath)

	// Verify file exists
	if !h.fileService.FileExists(fullPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio file not found"})
		return
	}

	// Create a safe filename for download
	ext := filepath.Ext(fullPath)
	downloadName := services.SanitizeFilename(track.Title) + ext

	// Serve file as attachment
	c.FileAttachment(fullPath, downloadName)
}

func (h *FileHandler) ServeCoverArt(c *gin.Context) {
	albumID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid album id"})
		return
	}

	// Get album
	album, err := h.albumService.GetAlbum(c.Request.Context(), uint64(albumID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	// Check if album has cover art
	if album.Metadata.CoverArtPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover art available"})
		return
	}

	fullCoverArtPath := h.fileService.GetFullPath(album.Metadata.CoverArtPath)

	// Verify file exists
	if !h.fileService.FileExists(fullCoverArtPath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover art file not found"})
		return
	}

	// Determine content type
	contentType := getImageContentType(fullCoverArtPath)

	// Set cache headers for images (they don't change often)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day

	// Serve the image file
	c.File(fullCoverArtPath)
}

func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/mp4"
	case ".aiff":
		return "audio/aiff"
	case ".alac":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
}

func getImageContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
