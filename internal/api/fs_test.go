package api

import (
	"os"
	"testing"
	"time"
)

func TestLoadRoutesFromFile(t *testing.T) {
	r := NewRouter()

	// create a temporary file
	tmpFile, err := os.CreateTemp("", "testRoutes*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// write some data to it
	_, err = tmpFile.WriteString(`[
		{
			"path": "/test",
			"method": "GET",
			"status_code": 200,
			"response_headers": {"Content-Type": "application/json"},
			"response_body": "Hello, World!"
		}
	]`)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// close the file
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// load the routes from the file
	err = r.LoadRoutesFromFiles([]string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("Failed to load routes from file: %v", err)
	}

	// delete the temp file
	err = os.Remove(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to delete temp file: %v", err)
	}

	// check the routes
	if len(r.Routes) != 1 {
		t.Errorf("Expected one route, got %v", len(r.Routes))
	}
}

func TestLoadRoutesFromDir(t *testing.T) {
	router := NewRouter()
	testCases := []struct {
		name          string
		setup         func() string
		teardown      func(string)
		expectedError bool
	}{
		{
			name: "non-existent directory",
			setup: func() string {
				return "nonexistentdir"
			},
			expectedError: true,
		},
		{
			name: "empty directory",
			setup: func() string {
				os.Mkdir("emptydir", os.ModePerm)
				return "emptydir"
			},
			teardown: func(dir string) {
				os.Remove(dir)
			},
		},
		{
			name: "directory with invalid json",
			setup: func() string {
				os.Mkdir("invalidjsondir", os.ModePerm)
				file, _ := os.Create("invalidjsondir/invalid.json")
				file.WriteString("invalid json")
				file.Close()
				return "invalidjsondir"
			},
			teardown: func(dir string) {
				os.RemoveAll(dir)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.setup()
			if tc.teardown != nil {
				defer tc.teardown(dir)
			}
			err := router.LoadRoutesFromDir(dir)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWatchRoutes(t *testing.T) {
	routesFilePath := "test_routes.json"

	// Write initial routes to file
	err := os.WriteFile(routesFilePath, []byte(`[
		{
			"path": "/test",
			"method": "GET",
			"status_code": 200
		}
	]`), 0o644)
	if err != nil {
		t.Fatalf("Failed to write routes to file: %v", err)
	}

	router := NewRouter()

	// Start watching routes in a separate goroutine
	go WatchRoutes(router, routesFilePath)

	// Wait for a while to ensure the watcher has started
	time.Sleep(1 * time.Second)

	// Modify the routes file
	err = os.WriteFile(routesFilePath, []byte(`[
		{
			"path": "/test",
			"method": "GET",
			"status_code": 201
		}
	]`), 0o644)
	if err != nil {
		t.Fatalf("Failed to write routes to file: %v", err)
	}

	// Wait for a while to ensure the watcher has processed the event
	time.Sleep(1 * time.Second)

	// Check if the route has been updated
	route, ok := router.Routes["/test"]
	if !ok || route.StatusCode != 201 {
		t.Errorf("Routes were not updated after file modification")
	}

	// Clean up
	err = os.Remove(routesFilePath)
	if err != nil {
		t.Errorf("os.Remove failed after test.")
	}
}
