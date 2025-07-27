package integration

import (
	"cms-backend/models"
	"cms-backend/routes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Package-level variables for test database connection and router
var (
	testDB *gorm.DB
	router *gin.Engine
)

/*
INTEGRATION TEST SETUP GUIDE

This file sets up the integration test environment for your CMS backend.
It handles database connections, schema migrations, and cleanup.

Key Components:
1. Test database connection
2. Router setup
3. Schema migrations
4. Test cleanup
*/

func TestMain(m *testing.M) {
	// STEP 1: Environment Setup
	setup()

	// STEP 2: Run Tests
	code := m.Run()

	// STEP 3: Cleanup
	cleanup()

	// Exit with test result code
	os.Exit(code)
}

func setup() {
	// STEP 1: Configure Gin
	gin.SetMode(gin.TestMode)

	// STEP 2: Database Connection
	// Define test database connection string
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPassword := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	dbName := getEnvOrDefault("TEST_DB_NAME", "cms_test")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	// Connect to test database using GORM
	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// STEP 3: Schema Migration
	// Migrate all model schemas
	err = testDB.AutoMigrate(
		&models.Media{},
		&models.Page{},
		&models.Post{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}

	// STEP 4: Router Setup
	router = gin.New()
	routes.InitializeRoutes(router, testDB)
}

func cleanup() {
	// STEP 1: Database Cleanup
	sqlDB, err := testDB.DB()
	if err != nil {
		log.Printf("Error getting underlying SQL database: %v", err)
		return
	}

	// Drop all tables in correct order (foreign key constraints)
	// 1. Junction tables first
	testDB.Exec("DROP TABLE IF EXISTS post_media CASCADE")
	// 2. Main tables next
	testDB.Exec("DROP TABLE IF EXISTS posts CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS media CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS pages CASCADE")

	// STEP 2: Connection Cleanup
	err = sqlDB.Close()
	if err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
}

func clearTables() {
	// STEP 1: Data Cleanup
	// Delete all data from tables in correct order to maintain referential integrity
	// 1. Junction tables first
	testDB.Exec("DELETE FROM post_media")
	// 2. Main tables next
	testDB.Exec("DELETE FROM posts")
	testDB.Exec("DELETE FROM media")
	testDB.Exec("DELETE FROM pages")
}

// getEnvOrDefault returns the environment variable value or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetTestDB returns the test database instance for use in other test files
func GetTestDB() *gorm.DB {
	return testDB
}

// GetTestRouter returns the test router instance for use in other test files
func GetTestRouter() *gin.Engine {
	return router
}

/*
TESTING HINTS:
1. Database Connection:
   - Use a separate test database
   - Consider environment variables for credentials
   - Handle connection errors properly

2. Table Management:
   - Drop tables in correct order (foreign key constraints)
   - Clear data between tests
   - Consider using transactions for tests

3. Error Handling:
   - Log setup/cleanup errors
   - Ensure proper resource cleanup
   - Handle database operation errors

4. Best Practices:
   - Use constants for connection strings
   - Consider test helper functions
   - Add proper logging for debugging
   - Document any required setup steps

5. Environment Variables for Testing:
   - TEST_DB_HOST (default: localhost)
   - TEST_DB_PORT (default: 5432)
   - TEST_DB_USER (default: postgres)
   - TEST_DB_PASSWORD (default: password)
   - TEST_DB_NAME (default: cms_test)

6. Usage in Test Files:
   - Import this package in your integration test files
   - Use the `router` variable to make HTTP requests
   - Use the `testDB` variable for direct database operations
   - Call `clearTables()` in test setup to ensure clean state
*/

/*
TESTING HINTS:
1. Database Connection:
   - Use a separate test database
   - Consider environment variables for credentials
   - Handle connection errors properly

2. Table Management:
   - Drop tables in correct order (foreign key constraints)
   - Clear data between tests
   - Consider using transactions for tests

3. Error Handling:
   - Log setup/cleanup errors
   - Ensure proper resource cleanup
   - Handle database operation errors

4. Best Practices:
   - Use constants for connection strings
   - Consider test helper functions
   - Add proper logging for debugging
   - Document any required setup steps
*/
