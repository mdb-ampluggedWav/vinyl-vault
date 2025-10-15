package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"vinyl-vault/internal/services"
	"vinyl-vault/pkg"

	"github.com/gin-gonic/gin"
)

type CreateAlbumRequest struct {
	Metadata pkg.Metadata `form:"metadata"`
}

type UpdateAlbumRequest struct {
	Metadata pkg.Metadata `form:"metadata"`
}

type AlbumHandler struct {
	albumService *services.AlbumService
	fileService  *services.FileService
}

func NewAlbumHandler(albumService *services.AlbumService, fileService *services.FileService) *AlbumHandler {
	return &AlbumHandler{
		albumService: albumService,
		fileService:  fileService,
	}
}

func (h *AlbumHandler) RegisterAlbumRoutes(router *gin.RouterGroup) {
	router.POST("/album", h.CreateAlbum)
	router.GET("/album/:id", h.GetAlbum)
	router.GET("/albums/me", h.GetMyAlbums)
	router.PUT("/album/:id", h.UpdateAlbum)
	router.DELETE("/album/:id", h.DeleteAlbum)
	router.GET("/album/:id/download", h.DownloadAlbum) // download as zip
}

func (h *AlbumHandler) CreateAlbum(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateAlbumRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create album first without coverart
	album, err := h.albumService.CreateAlbum(c.Request.Context(), userID.(uint64), req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// then we handle coverart if provided
	coverFile, err := c.FormFile("cover_art")
	if err == nil {
		result, err := h.fileService.SaveCoverArt(coverFile, album.ID)
		if err != nil {
			c.JSON(http.StatusCreated, gin.H{
				"album":   album,
				"warning": "Album created but cover art failed",
			})
			return
		}
		relativePath, err := h.fileService.GetRelativePath(result.Path)
		if err != nil {
			h.fileService.DeleteCoverArt(result.Path)
		}

		req.Metadata.CoverArtPath = relativePath
		album, err = h.albumService.UpdateAlbumInfo(c.Request.Context(), userID.(uint64), album.ID, req.Metadata)
		if err != nil {
			h.fileService.DeleteCoverArt(result.Path)
		}
	}

	c.JSON(http.StatusCreated, album)
}

func (h *AlbumHandler) GetAlbum(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	album, err := h.albumService.GetAlbum(c.Request.Context(), uint64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, album)
}

func (h *AlbumHandler) GetMyAlbums(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	albums, err := h.albumService.GetAlbumsByUser(c.Request.Context(), userID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

// DownloadAlbum creates a zip of all tracks and returns it
func (h *AlbumHandler) DownloadAlbum(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	albumID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Get album to verify ownership
	album, err := h.albumService.GetAlbum(c.Request.Context(), uint64(albumID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	// Get all track file paths
	var trackPaths []string
	for _, track := range album.Tracks {
		fullPath := h.fileService.GetFullPath(track.FilePath)
		trackPaths = append(trackPaths, fullPath)
	}

	if len(trackPaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "album has no tracks"})
		return
	}

	// Create zip
	zipName := fmt.Sprintf("album_%d_%s.zip", albumID, sanitizeFilename(album.Metadata.Album))
	zipPath, err := h.fileService.ArchiveAudioFilesToZip(trackPaths, zipName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create archive"})
		return
	}

	// Send file and cleanup
	defer h.fileService.DeleteZipFile(zipPath)
	c.FileAttachment(zipPath, zipName)
}

func (h *AlbumHandler) UpdateAlbum(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateAlbumRequest
	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle optional new cover art
	coverFile, err := c.FormFile("cover_art")
	if err == nil {
		result, err := h.fileService.SaveCoverArt(coverFile, uint64(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		relativePath, err := h.fileService.GetRelativePath(result.Path)

		if err != nil {
			h.fileService.DeleteCoverArt(result.Path)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process file path"})
			return
		}
		req.Metadata.CoverArtPath = relativePath
	}

	album, err := h.albumService.UpdateAlbumInfo(c.Request.Context(), userID.(uint64), uint64(id), req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, album)
}

func (h *AlbumHandler) DeleteAlbum(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err = h.albumService.DeleteAlbum(c.Request.Context(), userID.(uint64), uint64(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func sanitizeFilename(name string) string {

	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", " ", "_",
	)
	sanitized := replacer.Replace(name)
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	return sanitized
}
