package integration

import (
	"cms-backend/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TODO: Import required packages for:
// - JSON handling
// - HTTP testing
// - URL manipulation
// - String formatting
// - Your application models
// - Testing package

/*
POST INTEGRATION TESTS

These tests verify the complete flow of post operations through the API,
including relationships with media items.
Each test should:
1. Start with a clean database
2. Set up required relationships (media)
3. Perform post operations
4. Verify responses and relationships
*/

func TestPostIntegration(t *testing.T) {
	// Clear Database before starting tests
	clearTables()

	// Create Test Media for use in posts
	mediaID := createTestMedia(t)

	t.Run("Create Post with Media", func(t *testing.T) {
		// Test creating a post with media relationship
		body := `{
			"title": "Test Post with Media",
			"content": "This is a test post with media attachment",
			"author": "Test Author",
			"media": [{"id": ` + fmt.Sprintf("%d", mediaID) + `}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/posts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful creation
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
			return
		}

		// Parse response
		var response models.Post
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify post data
		if response.Title != "Test Post with Media" {
			t.Errorf("Expected title 'Test Post with Media', got '%s'", response.Title)
		}
		if response.Author != "Test Author" {
			t.Errorf("Expected author 'Test Author', got '%s'", response.Author)
		}
		if len(response.Media) != 1 {
			t.Errorf("Expected 1 media item, got %d", len(response.Media))
		}
	})

	t.Run("Get Posts with Filter", func(t *testing.T) {
		// Test retrieving posts with author filter
		req := httptest.NewRequest("GET", "/api/v1/posts?author=Test+Author", nil)
		w := httptest.NewRecorder()

		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful retrieval
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		// Parse response
		var posts []models.Post
		if err := json.Unmarshal(w.Body.Bytes(), &posts); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify filtered results
		if len(posts) != 1 {
			t.Errorf("Expected 1 post, got %d", len(posts))
		}
		if len(posts) > 0 && posts[0].Author != "Test Author" {
			t.Errorf("Expected author 'Test Author', got '%s'", posts[0].Author)
		}
	})
}

// Helper function to create test media
func createTestMedia(t *testing.T) uint {
	body := `{
		"url": "http://example.com/test.jpg",
		"type": "image"
	}`

	req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := GetTestRouter()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test media, status: %d, body: %s", w.Code, w.Body.String())
	}

	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to create test media: %v", err)
	}

	return response.ID
}

/*
TESTING HINTS:
1. Request Creation:
   - Use proper JSON formatting for relationships
   - Handle URL encoding for query parameters
   - Set appropriate headers

2. Response Validation:
   - Check both status codes and response content
   - Verify relationship data is correct
   - Validate filtered results carefully

3. Test Data:
   - Create meaningful test data
   - Handle relationships properly
   - Clean up between tests

4. Error Cases to Consider:
   - Invalid media IDs
   - Missing required fields
   - Invalid filter parameters
   - Non-existent relationships
*/
