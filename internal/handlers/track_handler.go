package handlers

import (
	"net/http"
	"strconv"
	"vinyl-vault/internal/services"
	"vinyl-vault/pkg"

	"github.com/gin-gonic/gin"
)

type CreateTrackRequest struct {
	AlbumID      uint64           `json:"album_id"`
	TrackNumber  int              `json:"track_number"`
	Title        string           `json:"title"`
	Duration     int              `json:"duration"`
	FilePath     string           `json:"file_path"`
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
}

func NewTrackHandler(trackService *services.TrackService) *TrackHandler {
	return &TrackHandler{
		trackService: trackService,
	}
}

func (h *TrackHandler) RegisterTrackRoutes(router *gin.Engine) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	track, err := h.trackService.CreateTrack(
		c.Request.Context(), userID.(uint64), req.AlbumID, req.TrackNumber, req.Title, req.Duration, req.FilePath, req.AudioQuality,
	)
	if err != nil {
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
		c.Request.Context(), userID.(uint64), uint64(id), req.TrackNumber, req.Title, req.Duration, req.AudioQuality,
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
