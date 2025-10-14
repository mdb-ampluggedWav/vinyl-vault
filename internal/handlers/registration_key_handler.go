package handlers

import (
	"net/http"
	"strconv"
	"vinyl-vault/internal/services"

	"github.com/gin-gonic/gin"
)

type GenerateKeyRequest struct {
	ExpirationHours int `json:"expiration_hours" binding:"required,min=1,max=8760"`
}

type ValidateKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

type RegistrationKeyHandler struct {
	keyService *services.RegistrationKeyService
}

func NewRegistrationKeyHandler(keyService *services.RegistrationKeyService) *RegistrationKeyHandler {
	return &RegistrationKeyHandler{
		keyService: keyService,
	}
}

func (h *RegistrationKeyHandler) RegisterKeyRoutes(router *gin.Engine) {

	router.POST("/admin/registration-key", h.GenerateKey)
	router.GET("/admin/registration-keys", h.GetMyKeys)
	router.DELETE("/admin/registration-key/:id", h.DeleteKey)

	router.POST("/validate-key", h.ValidateKey)
}

func (h *RegistrationKeyHandler) ValidateKey(c *gin.Context) {
	var req ValidateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.keyService.ValidateKey(c.Request.Context(), req.Key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"expires_at": key.ExpiresAt,
		"message":    "Registration key is valid. You can proceed with registration.",
	})
}

// GenerateKey creates a new one-time registration key (admin only)
func (h *RegistrationKeyHandler) GenerateKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req GenerateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.keyService.GenerateKey(c.Request.Context(), userID.(uint64), req.ExpirationHours)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":        key.Key,
		"expires_at": key.ExpiresAt,
		"message":    "Registration key generated successfully. Share this key with the user.",
	})
}

// GetMyKeys returns all keys created by admin
func (h *RegistrationKeyHandler) GetMyKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	keys, err := h.keyService.GetKeysByCreator(c.Request.Context(), userID.(uint64))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// DeleteKey removes an unused registration key (admin only)
func (h *RegistrationKeyHandler) DeleteKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	if err = h.keyService.DeleteKey(c.Request.Context(), uint64(keyID), userID.(uint64)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
