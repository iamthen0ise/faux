package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestAddRoute(t *testing.T) {
	router := NewRouter()
	route := &Route{
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
	router := NewRouter()
	route := &Route{
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
	router := NewRouter()
	route := &Route{
		Path:       "/test",
		Method:     "GET",
		StatusCode: 200,
	}

	router.AddRoute(route)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
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
	router := NewRouter()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/200?response_headers.Content-Type=text%2Fhtml&response_body=Hello+World", http.NoBody)

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
	router := NewRouter()

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

func TestParseRequestIntoMagicReq(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc           string
		reqBody        string
		expectedErr    bool
		expectedOutput MagicRequest
	}{
		{
			desc:        "Normal case",
			reqBody:     `{"response_headers": {"header1":"value1", "header2":"value2"}, "response_body": "Hello, World!"}`,
			expectedErr: false,
			expectedOutput: MagicRequest{
				ResponseHeaders: map[string]string{"header1": "value1", "header2": "value2"},
				ResponseBody:    "Hello, World!",
			},
		},
		{
			desc:           "Bad JSON case",
			reqBody:        `{"response_headers": "header1":"value1", "header2":"value2", "response_body": "Hello, World!"}`,
			expectedErr:    true,
			expectedOutput: MagicRequest{},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			router := NewRouter()

			req, _ := http.NewRequest(http.MethodGet, "/status/200", strings.NewReader(tC.reqBody))
			req.Header.Set("Content-Type", "application/json")

			var magicReq MagicRequest

			err := router.parseRequestIntoMagicReq(req, &magicReq)
			gotErr := err != nil

			if tC.expectedErr != gotErr {
				t.Errorf("Expected error status %v, but got %v", tC.expectedErr, gotErr)
			}

			if !tC.expectedErr {
				if magicReq.ResponseBody != tC.expectedOutput.ResponseBody {
					t.Errorf("Expected response body %v, but got %v", tC.expectedOutput.ResponseBody, magicReq.ResponseBody)
				}

				for key, value := range tC.expectedOutput.ResponseHeaders {
					if magicReq.ResponseHeaders[key] != value {
						t.Errorf("Expected response header %s to be %s, but got %s", key, value, magicReq.ResponseHeaders[key])
					}
				}
			}
		})
	}
}

func TestParseRequestIntoMagicReq_GET(t *testing.T) {
	req := httptest.NewRequest("GET", "/magic/200?response_headers.Content-Type=application%2Fjson&response_body=%7B%22message%22%3A%20%22Hello%2C%20World%21%22%7D", http.NoBody)
	router := NewRouter()

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	contentType, ok := magicReq.ResponseHeaders["Content-Type"]
	if !ok || contentType != "application/json" {
		t.Errorf("ResponseHeaders not parsed correctly: got %v, want 'application/json'", contentType)
	}

	if magicReq.ResponseBody != `{"message": "Hello, World!"}` {
		t.Errorf("ResponseBody not parsed correctly: got %v, want {\"message\": \"Hello, World!\"}", magicReq.ResponseBody)
	}
}

func TestParseRequestIntoMagicReq_POST(t *testing.T) {
	jsonBody := strings.NewReader(`{"response_headers":{"Content-Type":"application/json"},"response_body":"{\"message\": \"Hello, World!\"}"}`)
	req := httptest.NewRequest("POST", "/magic/200", jsonBody)
	req.Header.Set("Content-Type", "application/json")
	router := NewRouter()

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	contentType, ok := magicReq.ResponseHeaders["Content-Type"]
	if !ok || contentType != "application/json" {
		t.Errorf("ResponseHeaders not parsed correctly: got %v, want 'application/json'", contentType)
	}

	if magicReq.ResponseBody != `{"message": "Hello, World!"}` {
		t.Errorf("ResponseBody not parsed correctly: got %v, want {\"message\": \"Hello, World!\"}", magicReq.ResponseBody)
	}
}

func TestParseRequestIntoMagicReq_QueryParamsPresent(t *testing.T) {
	req := httptest.NewRequest("GET", "/magic/200?response_headers.Content-Type=application%2Fjson&response_body=%7B%22message%22%3A%20%22Hello%2C%20World%21%22%7D", http.NoBody)
	router := NewRouter()

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	contentType, ok := magicReq.ResponseHeaders["Content-Type"]
	if !ok || contentType != "application/json" {
		t.Errorf("ResponseHeaders not parsed correctly: got %v, want 'application/json'", contentType)
	}

	if magicReq.ResponseBody != `{"message": "Hello, World!"}` {
		t.Errorf("ResponseBody not parsed correctly: got %v, want {\"message\": \"Hello, World!\"}", magicReq.ResponseBody)
	}
}

func TestParseRequestIntoMagicReq_QueryParamsNotPresent(t *testing.T) {
	req := httptest.NewRequest("GET", "/magic/200", http.NoBody)
	router := NewRouter()

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if len(magicReq.ResponseHeaders) != 0 {
		t.Errorf("ResponseHeaders should be empty: got %v", magicReq.ResponseHeaders)
	}

	// if magicReq.ResponseBody != "" {
	// 	t.Errorf("ResponseBody should be empty: got %v", magicReq.ResponseBody)
	// }
}

func TestParseRequestIntoMagicReq_JsonPresent(t *testing.T) {
	jsonBody := strings.NewReader(`{"response_headers":{"Content-Type":"application/json"},"response_body":"{\"message\": \"Hello, World!\"}"}`)
	req := httptest.NewRequest("POST", "/magic/200", jsonBody)
	req.Header.Set("Content-Type", "application/json")
	router := NewRouter()

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	contentType, ok := magicReq.ResponseHeaders["Content-Type"]
	if !ok || contentType != "application/json" {
		t.Errorf("ResponseHeaders not parsed correctly: got %v, want 'application/json'", contentType)
	}

	if magicReq.ResponseBody != `{"message": "Hello, World!"}` {
		t.Errorf("ResponseBody not parsed correctly: got %v, want {\"message\": \"Hello, World!\"}", magicReq.ResponseBody)
	}
}

func TestParseRequestIntoMagicReq_JsonNotPresent(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest("POST", "/magic/200", http.NoBody)
	req.Header.Set("Content-Type", "application/json")

	magicReq := &MagicRequest{}
	err := router.parseRequestIntoMagicReq(req, magicReq)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if len(magicReq.ResponseHeaders) != 0 {
		t.Errorf("ResponseHeaders should be empty: got %v", magicReq.ResponseHeaders)
	}

	if magicReq.ResponseBody != http.NoBody {
		t.Errorf("ResponseBody should be empty: got %v", magicReq.ResponseBody)
	}
}

func TestParseDotNotation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc           string
		input          url.Values
		expectedOutput map[string]interface{}
	}{
		{
			desc: "Single level",
			input: url.Values{
				"name":    []string{"John"},
				"country": []string{"USA"},
			},
			expectedOutput: map[string]interface{}{
				"name":    "John",
				"country": "USA",
			},
		},
		{
			desc: "Nested level",
			input: url.Values{
				"user.name":      []string{"John"},
				"user.country":   []string{"USA"},
				"user.address":   []string{"New York"},
				"company.name":   []string{"OpenAI"},
				"company.office": []string{"San Francisco"},
			},
			expectedOutput: map[string]interface{}{
				"user": map[string]interface{}{
					"name":    "John",
					"country": "USA",
					"address": "New York",
				},
				"company": map[string]interface{}{
					"name":   "OpenAI",
					"office": "San Francisco",
				},
			},
		},
		{
			desc: "Mixed single and nested levels",
			input: url.Values{
				"name":         []string{"John"},
				"country":      []string{"USA"},
				"company.name": []string{"OpenAI"},
			},
			expectedOutput: map[string]interface{}{
				"name":    "John",
				"country": "USA",
				"company": map[string]interface{}{
					"name": "OpenAI",
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			output := ParseDotNotation(tC.input)

			if !reflect.DeepEqual(output, tC.expectedOutput) {
				t.Errorf("Expected output %v, but got %v", tC.expectedOutput, output)
			}
		})
	}
}

func TestLoadRoutesFromJSON(t *testing.T) {
	t.Parallel()

	router := NewRouter()

	jsonData := []byte(`[
		{
			"path": "/path1",
			"method": "GET",
			"status_code": 200
		},
		{
			"path": "/path2",
			"method": "POST",
			"status_code": 201
		}
	]`)

	err := router.LoadRoutesFromJSON(jsonData)
	if err != nil {
		t.Errorf("LoadRoutesFromJSON returned error: %v", err)
	}

	expectedRoutes := []Route{
		{
			Path:       "/path1",
			Method:     "GET",
			StatusCode: 200,
		},
		{
			Path:       "/path2",
			Method:     "POST",
			StatusCode: 201,
		},
	}

	for _, expected := range expectedRoutes {
		route, ok := router.Routes[expected.Path]
		if !ok {
			t.Errorf("Route not added: %s", expected.Path)
		} else if !reflect.DeepEqual(route, &expected) {
			t.Errorf("Route not correctly added: got %+v, want %+v", route, &expected)
		}
	}
}

func TestAuthMiddleware(t *testing.T) {
	// create router
	router := NewRouter()

	// add route with auth required
	route := &Route{
		Path:         "/auth",
		Method:       "GET",
		StatusCode:   http.StatusOK,
		AuthRequired: true,
	}
	router.AddRoute(route)

	// create middleware
	middleware := &AuthMiddleware{
		Token: "mytoken",
		Next:  router,
	}

	// Test with no auth
	req := httptest.NewRequest("GET", "/auth", http.NoBody)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Test with auth
	req = httptest.NewRequest("GET", "/auth", http.NoBody)
	req.Header.Set("Authorization", "mytoken")
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test with incorrect auth
	req = httptest.NewRequest("GET", "/auth", http.NoBody)
	req.Header.Set("Authorization", "wrongtoken")
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Test with no auth required
	route = &Route{
		Path:         "/noauth",
		Method:       "GET",
		StatusCode:   http.StatusOK,
		AuthRequired: false,
	}
	router.AddRoute(route)

	req = httptest.NewRequest("GET", "/noauth", http.NoBody)
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
