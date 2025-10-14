package handlers

import (
	"net/http"
	"strconv"
	"vinyl-vault/internal/services"
	"vinyl-vault/pkg"

	"github.com/gin-gonic/gin"
)

type CreateTrackRequest struct {
	AlbumID      uint64           `json:"album_id" binding:"required"`
	TrackNumber  int              `json:"track_number" binding:"required"`
	Title        string           `json:"title" binding:"required"`
	Duration     int              `json:"duration"`
	AudioQuality pkg.AudioQuality `json:"audio_quality"`
}

type UpdateTrackRequest struct {
	TrackNumber  *int              `json:"track_number,omitempty"`
	Title        *string           `json:"title,omitempty"`
	Duration     *int              `json:"duration,omitempty"`
	AudioQuality *pkg.AudioQuality `json:"audio_quality,omitempty"`
}

type TrackHandler struct {
	trackService *services.TrackService
	fileService  *services.FileService
}

func NewTrackHandler(trackService *services.TrackService, fileService *services.FileService) *TrackHandler {
	return &TrackHandler{
		trackService: trackService,
		fileService:  fileService,
	}
}

func (h *TrackHandler) RegisterTrackRoutes(router *gin.RouterGroup) {
	router.POST("/track", h.CreateTrack)
	router.GET("/track/:id", h.GetTrack)
	router.PUT("/track/:id", h.UpdateTrack)
	router.DELETE("/track/:id", h.DeleteTrack)
}

func (h *TrackHandler) CreateTrack(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateTrackRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get uploaded audio file
	file, err := c.FormFile("audio_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file required"})
		return
	}

	// save audio file
	result, err := h.fileService.SaveTrackAudioFile(file, req.AlbumID, req.TrackNumber, req.Title)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	track, err := h.trackService.CreateTrack(
		c.Request.Context(),
		userID.(uint64),
		req.AlbumID,
		req.TrackNumber,
		req.Title,
		req.Duration,
		result.Path,
		req.AudioQuality,
	)
	if err != nil {
		// cleanup file if track creation fails
		h.fileService.DeleteAudioFile(result.Path)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, track)
}

func (h *TrackHandler) GetTrack(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	track, err := h.trackService.GetTrack(c.Request.Context(), uint64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) GetAlbumTracks(c *gin.Context) {
	albumID, err := strconv.ParseInt(c.Param("album_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid album id"})
		return
	}

	tracks, err := h.trackService.GetTracksByAlbum(c.Request.Context(), uint64(albumID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tracks)
}

func (h *TrackHandler) UpdateTrack(c *gin.Context) {

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

	var req UpdateTrackRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	track, err := h.trackService.UpdateTrack(
		c.Request.Context(),
		userID.(uint64),
		uint64(id),
		req.TrackNumber,
		req.Title,
		req.Duration,
		req.AudioQuality,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) DeleteTrack(c *gin.Context) {

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
	err = h.trackService.DeleteTrack(c.Request.Context(), userID.(uint64), uint64(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
