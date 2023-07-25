package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iamthen0ise/faux/internal/api"
)

func TestAddRoute(t *testing.T) {
	router := api.NewRouter()
	route := &api.Route{
		Path:       "/test",
		Method:     "GET",
		StatusCode: 200,
	}

	router.AddRoute(route)

	if _, ok := router.Routes["/test"]; !ok {
		t.Fatalf("Route not added correctly.")
	}
}

func TestServeHTTP_JsonBody(t *testing.T) {
	router := api.NewRouter()
	route := &api.Route{
		Path:       "/test",
		Method:     "POST",
		StatusCode: 200,
	}

	router.AddRoute(route)

	jsonBody := strings.NewReader(`{"response_headers":{"Content-Type":"application/json"},"response_body":"{\"message\":\"Hello, World!\"}"}`)
	req := httptest.NewRequest("POST", "/test", jsonBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the header
	expectedHeader := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedHeader {
		t.Errorf("Content type header does not match: got %v want %v", ctype, expectedHeader)
	}

	// Check the body
	expected := `"{\"message\":\"Hello, World!\"}"` // Notice the escaped quotes here
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServeHTTP_NoArgs(t *testing.T) {
	router := api.NewRouter()
	route := &api.Route{
		Path:       "/test",
		Method:     "GET",
		StatusCode: 200,
	}

	router.AddRoute(route)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the header
	expectedHeader := ""
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedHeader {
		t.Errorf("Content type header does not match: got %v want %v", ctype, expectedHeader)
	}

	// Check the body
	// expected := ""
	// if rr.Body.String() != expected {
	// 	t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	// }
}

func TestServeHTTP_QueryParams(t *testing.T) {
	router := api.NewRouter()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/200?response_headers.Content-Type=text%2Fhtml&response_body=Hello+World", nil)

	router.ServeHTTP(rec, req)

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if ctype := rec.Header().Get("Content-Type"); ctype != "text/html" {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, "text/html")
	}

	if body := rec.Body.String(); body != `"Hello World"` {
		t.Errorf("handler returned unexpected body: got %v want %v",
			body, "Hello World")
	}
}

func TestServeHTTP_QueryAndJson(t *testing.T) {
	router := api.NewRouter()

	jsonBody := strings.NewReader(`{"response_headers": {"Content-Type": "application/json"}, "response_body": "{\"message\": \"Hello, World!\"}"}`)
	req := httptest.NewRequest("GET", "/status/200", jsonBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the header
	expectedHeader := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expectedHeader {
		t.Errorf("Content type header does not match: got %v want %v", ctype, expectedHeader)
	}

	// Check the body
	expected := `"{\"message\": \"Hello, World!\"}"`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
