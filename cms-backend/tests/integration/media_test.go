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

/*
MEDIA INTEGRATION TESTS

These tests verify the complete flow of media operations through the API.
Each test should:
1. Start with a clean database state
2. Perform API operations
3. Verify the responses
4. Check database state if needed
*/

func TestMediaIntegration(t *testing.T) {
	// Clear Database before starting tests
	clearTables()

	t.Run("Create Media", func(t *testing.T) {
		// STEP 1: Prepare Test Data
		// Create JSON body with URL and type
		body := `{
			"url": "http://example.com/test.jpg",
			"type": "image"
		}`

		// STEP 2: Create HTTP Request
		// Create POST request to /api/v1/media with proper headers
		req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// STEP 3: Execute Request
		// Create response recorder and send request through router
		w := httptest.NewRecorder()
		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful creation
		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		// STEP 4: Verify Response
		// Parse response JSON and verify media properties
		var response models.Media
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify media properties
		if response.URL != "http://example.com/test.jpg" {
			t.Errorf("Expected URL 'http://example.com/test.jpg', got %s", response.URL)
		}
		if response.Type != "image" {
			t.Errorf("Expected type 'image', got %s", response.Type)
		}
		if response.ID == 0 {
			t.Error("Expected non-zero ID")
		}
	})

	t.Run("Get All Media", func(t *testing.T) {
		// STEP 1: Setup Test Data
		// Create test media entries
		createTestMediaItem(t, "http://example.com/image1.jpg", "image")
		createTestMediaItem(t, "http://example.com/video1.mp4", "video")

		// STEP 2: Create HTTP Request
		// Create GET request to /api/v1/media
		req := httptest.NewRequest("GET", "/api/v1/media", nil)

		// STEP 3: Execute Request
		// Create response recorder and send request through router
		w := httptest.NewRecorder()
		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful retrieval
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// STEP 4: Verify Response
		// Parse response JSON array and verify properties
		var mediaList []models.Media
		if err := json.Unmarshal(w.Body.Bytes(), &mediaList); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check number of items (should be at least 2 from setup + any from previous tests)
		if len(mediaList) < 2 {
			t.Errorf("Expected at least 2 media items, got %d", len(mediaList))
		}

		// Verify media properties exist
		for _, media := range mediaList {
			if media.ID == 0 {
				t.Error("Media ID should not be zero")
			}
			if media.URL == "" {
				t.Error("Media URL should not be empty")
			}
			if media.Type == "" {
				t.Error("Media Type should not be empty")
			}
		}
	})

	t.Run("Get Single Media by ID", func(t *testing.T) {
		// STEP 1: Create test media
		mediaID := createTestMediaItem(t, "http://example.com/single.jpg", "image")

		// STEP 2: Create HTTP Request
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/media/%d", mediaID), nil)

		// STEP 3: Execute Request
		w := httptest.NewRecorder()
		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful retrieval
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// STEP 4: Verify Response
		var response models.Media
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != mediaID {
			t.Errorf("Expected ID %d, got %d", mediaID, response.ID)
		}
		if response.URL != "http://example.com/single.jpg" {
			t.Errorf("Expected URL 'http://example.com/single.jpg', got %s", response.URL)
		}
		if response.Type != "image" {
			t.Errorf("Expected type 'image', got %s", response.Type)
		}
	})

	t.Run("Create Media with Invalid Data", func(t *testing.T) {
		// Test missing required URL field
		body := `{
			"type": "image"
		}`

		req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Should return bad request
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Delete Media", func(t *testing.T) {
		// STEP 1: Create test media to delete
		mediaID := createTestMediaItem(t, "http://example.com/delete-me.jpg", "image")

		// STEP 2: Delete the media
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/media/%d", mediaID), nil)
		w := httptest.NewRecorder()

		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Verify successful deletion
		if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
			t.Fatalf("Expected status 200 or 204, got %d: %s", w.Code, w.Body.String())
		}

		// STEP 3: Verify media is deleted by trying to get it
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/media/%d", mediaID), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return not found
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 after deletion, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Get Non-existent Media", func(t *testing.T) {
		// Try to get media with non-existent ID
		req := httptest.NewRequest("GET", "/api/v1/media/99999", nil)
		w := httptest.NewRecorder()

		router := GetTestRouter()
		router.ServeHTTP(w, req)

		// Should return not found
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// Helper function to create test media and return its ID
func createTestMediaItem(t *testing.T, url, mediaType string) uint {
	body := fmt.Sprintf(`{
		"url": "%s",
		"type": "%s"
	}`, url, mediaType)

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
		t.Fatalf("Failed to parse test media response: %v", err)
	}

	return response.ID
}

/*
TESTING HINTS:
1. Request Creation:
   - Use httptest.NewRequest for creating requests
   - Remember to set Content-Type for POST requests
   - Use strings.NewReader for request bodies

2. Response Handling:
   - Use httptest.NewRecorder for capturing responses
   - Parse JSON responses carefully
   - Check both status codes and response bodies

3. Test Data:
   - Use meaningful test data
   - Clean up between tests
   - Consider edge cases

4. Error Cases:
   - Test invalid inputs
   - Test missing required fields
   - Test invalid content types
*/
