package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPages retrieves all pages
func GetPages(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Declare pages slice variable
	var pages []models.Page

	// Query all pages from database
	title := c.Query("title")
	author := c.Query("author")

	query := db
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author = ?", author)
	}

	// Handle potential database errors
	if err := query.Find(&pages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return success response with pages
	c.JSON(http.StatusOK, pages)
}

// GetPage retrieves a specific page by ID
func GetPage(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Get ID parameter and convert to uint
	id := c.Param("id")

	// Declare page variable
	var page models.Page

	// Query page from database
	if err := db.First(&page, id).Error; err != nil {
		// Handle potential database errors
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Page not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return success response with page
	c.JSON(http.StatusOK, page)
}

// CreatePage creates a new page
func CreatePage(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Declare page variable
	var page models.Page

	// Bind JSON request body
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Validate required fields
	if page.Title == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Title is required",
		})
		return
	}
	if page.Content == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Content is required",
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

	// Create page in database
	if err := tx.Create(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Commit transaction and return response
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, page)
}

// UpdatePage updates an existing page by ID
func UpdatePage(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Convert string ID to uint
	id := c.Param("id")

	// Find existing page
	var existingPage models.Page
	if err := db.First(&existingPage, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Page not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Bind JSON update data
	var updateData models.Page
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Update page fields
	if updateData.Title != "" {
		existingPage.Title = updateData.Title
	}
	if updateData.Content != "" {
		existingPage.Content = updateData.Content
	}

	// Start transaction and save
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&existingPage).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, existingPage)
}

// DeletePage deletes a page by ID
func DeletePage(c *gin.Context) {
	// Get database instance from context
	db := c.MustGet("db").(*gorm.DB)

	// Convert string ID to uint
	id := c.Param("id")

	// Check if page exists
	var page models.Page
	if err := db.First(&page, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Page not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Start transaction and delete
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Page deleted successfully",
	})
}
