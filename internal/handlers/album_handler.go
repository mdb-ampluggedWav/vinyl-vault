package handlers

import (
	"net/http"
	"strconv"
	"vinyl-vault/internal/services"
	"vinyl-vault/pkg"

	"github.com/gin-gonic/gin"
)

type CreateAlbumRequest struct {
	Metadata pkg.Metadata `json:"metadata"`
}

type UpdateAlbumRequest struct {
	Metadata pkg.Metadata `json:"metadata"`
}

type AlbumHandler struct {
	albumService *services.AlbumService
}

func NewAlbumHandler(albumService *services.AlbumService) *AlbumHandler {
	return &AlbumHandler{
		albumService: albumService,
	}
}

func (h *AlbumHandler) RegisterAlbumRoutes(router *gin.Engine) {
	router.POST("/album", h.CreateAlbum)
	router.GET("/album/:id", h.GetAlbum)
	router.GET("/albums/me", h.GetMyAlbums)
	router.PUT("/album/:id", h.UpdateAlbum)
	router.DELETE("/album/:id", h.DeleteAlbum)
}

func (h *AlbumHandler) CreateAlbum(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	var req CreateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	album, err := h.albumService.CreateAlbum(c.Request.Context(), userID.(uint64), req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
