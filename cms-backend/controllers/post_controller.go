package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPosts retrieves all posts with optional filtering
func GetPosts(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var posts []models.Post

	title := c.Query("title")
	author := c.Query("author")

	query := db
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author = ?", author)
	}

	// Use proper preloading for media relationships
	if err := query.Preload("Media").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// GetPost retrieves a specific post by ID
func GetPost(c *gin.Context) {
	// Get database instance from Gin context
	db := c.MustGet("db").(*gorm.DB)
	
	// Get the ID from URL parameter
	id := c.Param("id")
	
	// Define post variable and query database
	var post models.Post
	if err := db.Preload("Media").First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Post not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	
	// Return the post
	c.JSON(http.StatusOK, post)
}

// CreatePost creates a new post
func CreatePost(c *gin.Context) {
	// Get database instance from Gin context
	db := c.MustGet("db").(*gorm.DB)
	
	// Define post variable to store incoming data
	var post models.Post
	
	// Parse JSON request body into post struct
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	
	// Validate required fields
	if post.Title == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Title is required",
		})
		return
	}
	if post.Content == "" {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Content is required",
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
	
	// Create the post
	if err := tx.Create(&post).Error; err != nil {
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
	
	// Return created post
	c.JSON(http.StatusCreated, post)
}

// UpdatePost updates an existing post
func UpdatePost(c *gin.Context) {
	// Get database instance from Gin context
	db := c.MustGet("db").(*gorm.DB)
	
	// Get ID from URL parameter
	id := c.Param("id")
	
	// Find existing post
	var existingPost models.Post
	if err := db.First(&existingPost, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Post not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	
	// Define variable for update input
	var updateData models.Post
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	
	// Update only the fields that are allowed to be updated
	if updateData.Title != "" {
		existingPost.Title = updateData.Title
	}
	if updateData.Content != "" {
		existingPost.Content = updateData.Content
	}
	if updateData.Author != "" {
		existingPost.Author = updateData.Author
	}
	
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Save the updated post
	if err := tx.Save(&existingPost).Error; err != nil {
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
	
	// Return updated post
	c.JSON(http.StatusOK, existingPost)
}

// DeletePost deletes a post
func DeletePost(c *gin.Context) {
	// Get database instance from Gin context
	db := c.MustGet("db").(*gorm.DB)
	
	// Get ID from URL parameter
	id := c.Param("id")
	
	// Find existing post
	var post models.Post
	if err := db.First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Post not found",
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
	
	// Delete the post (soft delete if GORM's DeletedAt is configured, otherwise hard delete)
	if err := tx.Delete(&post).Error; err != nil {
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
		"message": "Post deleted successfully",
	})
}