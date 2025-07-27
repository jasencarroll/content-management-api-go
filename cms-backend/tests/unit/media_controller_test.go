package controllers

import (
	"bytes"
	"cms-backend/controllers"
	"cms-backend/models"
	"cms-backend/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetMedia(t *testing.T) {
	// Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// Mock Data Creation
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "https://example.com/image1.jpg", "image", now, now).
		AddRow(2, "https://example.com/video1.mp4", "video", now, now)

	// Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "media"`).WillReturnRows(rows)

	// HTTP Test Setup
	router.GET("/media", controllers.GetMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media", nil)
	router.ServeHTTP(w, req)

	// Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response []models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if len(response) != 2 {
		t.Fatalf("Expected 2 media items, but got %d", len(response))
	}
}

func TestGetMediaByID(t *testing.T) {
	// Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// Mock Data Creation
	now := time.Now()
	row := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "https://example.com/test.jpg", "image", now, now)

	// Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(row)

	// HTTP Test Setup
	router.GET("/media/:id", controllers.GetMediaByID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media/1", nil)
	router.ServeHTTP(w, req)

	// Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.ID != 1 {
		t.Fatalf("Expected media ID 1, but got %d", response.ID)
	}
	if response.URL != "https://example.com/test.jpg" {
		t.Fatalf("Expected URL 'https://example.com/test.jpg', but got '%s'", response.URL)
	}
	if response.Type != "image" {
		t.Fatalf("Expected type 'image', but got '%s'", response.Type)
	}
}

func TestCreateMedia(t *testing.T) {
	// Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// Database Expectations
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "media" \("url","type","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4\) RETURNING "id"`).
		WithArgs("https://example.com/new-image.jpg", "image", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Request Preparation
	media := models.Media{
		URL:  "https://example.com/new-image.jpg",
		Type: "image",
	}
	jsonData, _ := json.Marshal(media)

	// HTTP Test Setup
	router.POST("/media", controllers.CreateMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/media", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Response Validation
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, but got %d", w.Code)
	}

	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.URL != "https://example.com/new-image.jpg" {
		t.Fatalf("Expected URL 'https://example.com/new-image.jpg', but got '%s'", response.URL)
	}
	if response.Type != "image" {
		t.Fatalf("Expected type 'image', but got '%s'", response.Type)
	}
}

func TestDeleteMedia(t *testing.T) {
	// Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// Database Expectations
	now := time.Now()
	existingRow := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "https://example.com/delete-me.jpg", "image", now, now)

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(existingRow)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "media" WHERE "media"\."id" = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// HTTP Test Setup
	router.DELETE("/media/:id", controllers.DeleteMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/media/1", nil)
	router.ServeHTTP(w, req)

	// Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	expectedMessage := "Media deleted successfully"
	if response["message"] != expectedMessage {
		t.Fatalf("Expected message '%s', but got '%s'", expectedMessage, response["message"])
	}
}
