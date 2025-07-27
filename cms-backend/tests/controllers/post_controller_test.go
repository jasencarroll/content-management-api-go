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

func TestGetPosts(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Mock Data Creation
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "First Post", "Content 1", "Author 1", now, now).
		AddRow(2, "Second Post", "Content 2", "Author 2", now, now)

	// STEP 3: Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "posts"`).WillReturnRows(rows)
	
	// Mock the media preloading query
	mediaRows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at", "post_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(mediaRows)

	// STEP 4: HTTP Test Setup
	router.GET("/posts", controllers.GetPosts)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response []models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if len(response) != 2 {
		t.Fatalf("Expected 2 posts, but got %d", len(response))
	}
}

func TestGetPostsWithFilters(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Mock Data Creation
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Test Post", "Test Content", "TestAuthor", now, now)

	// STEP 3: Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE title ILIKE \$1 AND author = \$2`).
		WithArgs("%Test%", "TestAuthor").
		WillReturnRows(rows)
		
	// Mock the media preloading query
	mediaRows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at", "post_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(1).
		WillReturnRows(mediaRows)

	// STEP 4: HTTP Test Setup
	router.GET("/posts", controllers.GetPosts)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts?title=Test&author=TestAuthor", nil)
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response []models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if len(response) != 1 {
		t.Fatalf("Expected 1 post, but got %d", len(response))
	}
	if response[0].Title != "Test Post" {
		t.Fatalf("Expected title 'Test Post', but got '%s'", response[0].Title)
	}
}

func TestGetPost(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Mock Data Creation
	now := time.Now()
	row := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Test Post", "Test Content", "Test Author", now, now)

	// STEP 3: Database Expectations
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(row)
		
	// Mock the media preloading query
	mediaRows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at", "post_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(1).
		WillReturnRows(mediaRows)

	// STEP 4: HTTP Test Setup
	router.GET("/posts/:id", controllers.GetPost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.ID != 1 {
		t.Fatalf("Expected post ID 1, but got %d", response.ID)
	}
	if response.Title != "Test Post" {
		t.Fatalf("Expected title 'Test Post', but got '%s'", response.Title)
	}
}

func TestCreatePost(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "posts" \("title","content","author","created_at","updated_at"\) VALUES \(\$1,\$2,\$3,\$4,\$5\) RETURNING "id"`).
		WithArgs("New Post", "New Content", "New Author", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// STEP 3: Request Preparation
	post := models.Post{
		Title:   "New Post",
		Content: "New Content",
		Author:  "New Author",
	}
	jsonData, _ := json.Marshal(post)

	// STEP 4: HTTP Test Setup
	router.POST("/posts", controllers.CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, but got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.Title != "New Post" {
		t.Fatalf("Expected title 'New Post', but got '%s'", response.Title)
	}
	if response.Author != "New Author" {
		t.Fatalf("Expected author 'New Author', but got '%s'", response.Author)
	}
}

func TestUpdatePost(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	now := time.Now()
	existingRow := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Old Title", "Old Content", "Old Author", now, now)

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(existingRow)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "posts" SET "title"=\$1,"content"=\$2,"author"=\$3,"created_at"=\$4,"updated_at"=\$5 WHERE "id" = \$6`).
		WithArgs("Updated Title", "Updated Content", "Updated Author", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// STEP 3: Request Preparation
	updateData := map[string]string{
		"title":   "Updated Title",
		"content": "Updated Content",
		"author":  "Updated Author",
	}
	jsonData, _ := json.Marshal(updateData)

	// STEP 4: HTTP Test Setup
	router.PUT("/posts/:id", controllers.UpdatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/posts/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// STEP 5: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response.Title != "Updated Title" {
		t.Fatalf("Expected title 'Updated Title', but got '%s'", response.Title)
	}
	if response.Author != "Updated Author" {
		t.Fatalf("Expected author 'Updated Author', but got '%s'", response.Author)
	}
}

func TestDeletePost(t *testing.T) {
	// STEP 1: Test Setup
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	// STEP 2: Database Expectations
	now := time.Now()
	existingRow := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Post to Delete", "Content to Delete", "Author", now, now)

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(existingRow)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "posts" WHERE "posts"\."id" = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// STEP 3: HTTP Test Setup
	router.DELETE("/posts/:id", controllers.DeletePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	router.ServeHTTP(w, req)

	// STEP 4: Response Validation
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	expectedMessage := "Post deleted successfully"
	if response["message"] != expectedMessage {
		t.Fatalf("Expected message '%s', but got '%s'", expectedMessage, response["message"])
	}
}
