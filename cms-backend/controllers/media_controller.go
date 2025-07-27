package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMedia(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)
	var media []models.Media

	// Support filtering by type
	mediaType := c.Query("type")

	query := db
	if mediaType != "" {
		query = query.Where("type = ?", mediaType)
	}

	// Retrieve all media with optional filtering
	if err := query.Find(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

func GetMediaByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Get ID parameter from URL
	id := c.Param("id")

	// Find media by ID
	var media models.Media
	if err := db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Media not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

func CreateMedia(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Define media variable to store incoming data
	var media models.Media

	// Parse JSON request body into media struct
	if err := c.ShouldBindJSON(&media); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Validate required fields
	if media.URL == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "URL is required",
		})
		return
	}
	if media.Type == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Type is required",
		})
		return
	}

	// Start database transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the media
	if err := tx.Create(&media).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return created media
	c.JSON(http.StatusCreated, media)
}

func DeleteMedia(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Get ID parameter from URL
	id := c.Param("id")

	// Check if media exists
	var media models.Media
	if err := db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Media not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete the media
	if err := tx.Delete(&media).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Media deleted successfully",
	})
}
