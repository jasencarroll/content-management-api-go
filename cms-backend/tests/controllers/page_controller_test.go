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

// TODO: Import required packages for:
// - HTTP testing (net/http, httptest)
// - JSON handling (encoding/json)
// - Database mocking (sqlmock)
// - Time handling
// - Your application packages (models, utils)

func TestGetPages(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "First Page", "Content 1", time.Now(), time.Now()).
		AddRow(2, "Second Page", "Content 2", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "pages"`).WillReturnRows(rows)

	router.GET("/pages", controllers.GetPages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response []models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if len(response) != 2 {
		t.Fatalf("Expected 2 pages, but got %d", len(response))
	}
}

func TestGetPage(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Mock Data Creation
	now := time.Now()
	row := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Test Page", "Test Content", now, now)

	// STEP 3: Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(row)

	// STEP 4: HTTP Test Setup
	router.GET("/pages/:id", controllers.GetPage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages/1", nil)
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.ID != 1 {
		t.Fatalf("Expected page ID 1, but got %d", response.ID)
	}
	if response.Title != "Test Page" {
		t.Fatalf("Expected title 'Test Page', but got '%s'", response.Title)
	}
}

func TestCreatePage(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "pages" \("title","content","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4\) RETURNING "id"`).
		WithArgs("New Page", "New Content", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// STEP 3: Request Preparation
	page := models.Page{
		Title:   "New Page",
		Content: "New Content",
	}
	jsonData, _ := json.Marshal(page)

	// STEP 4: HTTP Test Setup
	router.POST("/pages", controllers.CreatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, but got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.Title != "New Page" {
		t.Fatalf("Expected title 'New Page', but got '%s'", response.Title)
	}
	if response.Content != "New Content" {
		t.Fatalf("Expected content 'New Content', but got '%s'", response.Content)
	}
}

func TestUpdatePage(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	now := time.Now()
	existingRow := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Old Title", "Old Content", now, now)

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(existingRow)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "pages" SET "title"=\$1,"content"=\$2,"created_at"=\$3,"updated_at"=\$4 WHERE "id" = \$5`).
		WithArgs("Updated Title", "Updated Content", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// STEP 3: Request Preparation
	updateData := map[string]string{
		"title":   "Updated Title",
		"content": "Updated Content",
	}
	jsonData, _ := json.Marshal(updateData)

	// STEP 4: HTTP Test Setup
	router.PUT("/pages/:id", controllers.UpdatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/pages/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.Title != "Updated Title" {
		t.Fatalf("Expected title 'Updated Title', but got '%s'", response.Title)
	}
	if response.Content != "Updated Content" {
		t.Fatalf("Expected content 'Updated Content', but got '%s'", response.Content)
	}
}

func TestDeletePage(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	now := time.Now()
	existingRow := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Page to Delete", "Content to Delete", now, now)

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(existingRow)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "pages" WHERE "pages"\."id" = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// STEP 3: HTTP Test Setup
	router.DELETE("/pages/:id", controllers.DeletePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/pages/1", nil)
	router.ServeHTTP(w, req)

	// STEP 4: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	expectedMessage := "Page deleted successfully"
	if response["message"] != expectedMessage {
		t.Fatalf("Expected message '%s', but got '%s'", expectedMessage, response["message"])
	}
}

/*
TESTING HINTS:
1. Use sqlmock.AnyArg() for timestamp fields
2. Remember to escape special characters in SQL patterns
3. Each database operation needs proper error handling
4. Content-Type header is required for POST/PUT requests
5. Transaction tests need Begin/Commit expectations
6. Use proper argument matching in mock expectations
7. Consider testing error cases:
   - Invalid IDs
   - Missing required fields
   - Database errors
*/
