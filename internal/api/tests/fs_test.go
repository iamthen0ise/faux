package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/iamthen0ise/faux/internal/api"
)

func TestLoadRoutesFromFile(t *testing.T) {
	r := api.NewRouter()

	// create a temporary file
	tmpFile, err := ioutil.TempFile("", "testRoutes*.json")
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
	r := api.NewRouter()

	// create a temporary directory
	tmpDir, err := ioutil.TempDir("", "testRoutes")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// create a temporary file in the directory
	tmpFile, err := ioutil.TempFile(tmpDir, "*.json")
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

	// load the routes from the directory
	err = r.LoadRoutesFromDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load routes from directory: %v", err)
	}

	// delete the temp directory
	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatalf("Failed to delete temp directory: %v", err)
	}

	// check the routes
	if len(r.Routes) != 1 {
		t.Errorf("Expected one route, got %v", len(r.Routes))
	}
}
