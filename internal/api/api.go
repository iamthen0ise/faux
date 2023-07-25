package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Route struct {
	Path            string            `json:"path"`
	Method          string            `json:"method"`
	StatusCode      int               `json:"status_code"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseBody    interface{}       `json:"response_body,omitempty"`
	Lambda          int               `json:"-"`
	AuthRequired    bool
}

const (
	MaxQuerySize = 512 // bytes
)

var (
	ErrQueryTooLarge = errors.New("query size exceeds the maximum allowed size")
)

type Router struct {
	Routes map[string]*Route
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[string]*Route),
	}
}

type MagicRequest struct {
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseBody    interface{}       `json:"response_body,omitempty"`
}

func (r *Router) parseRequestIntoMagicReq(req *http.Request, magicReq *MagicRequest) error {
	if req.Header.Get("Content-Type") == "application/json" {
		defer req.Body.Close()
		if req.Body != http.NoBody {
			err := json.NewDecoder(req.Body).Decode(magicReq)
			if err != nil {
				return errors.New("Invalid JSON payload")
			}
		} else {
			magicReq.ResponseBody = http.NoBody
		}
	} else {
		// Parse dot notation query parameters.
		query := req.URL.Query()
		for k, v := range query {
			if strings.Contains(k, ".") {
				parts := strings.Split(k, ".")
				// We are assuming that we only support one level of nested structure for simplicity.
				// For more levels, consider using a recursive function.
				if len(parts) == 2 {
					if parts[0] == "response_headers" {
						if magicReq.ResponseHeaders == nil {
							magicReq.ResponseHeaders = make(map[string]string)
						}
						magicReq.ResponseHeaders[parts[1]] = v[0]
					} else if parts[0] == "response_body" {
						// We assume that v[0] is a JSON string and unmarshal it into a map.
						var responseBodyMap map[string]interface{}
						if err := json.Unmarshal([]byte(v[0]), &responseBodyMap); err != nil {
							return errors.New("Invalid response body")
						}
						magicReq.ResponseBody = responseBodyMap
					}
				}
			} else {
				// Handle non-nested query parameters.
				if k == "response_body" {
					magicReq.ResponseBody = v[0]
				}
			}
		}
	}
	return nil
}

func ParseDotNotation(m url.Values) map[string]interface{} {
	result := make(map[string]interface{})

	for key, values := range m {
		keys := strings.Split(key, ".")
		lastKey := keys[len(keys)-1]
		keys = keys[:len(keys)-1]

		innerMap := result
		for _, innerKey := range keys {
			if _, ok := innerMap[innerKey]; !ok {
				innerMap[innerKey] = make(map[string]interface{})
			}

			innerMap = innerMap[innerKey].(map[string]interface{})
		}

		innerMap[lastKey] = values[0]
	}

	return result
}

func (r *Router) AddRoute(route *Route) {
	r.Routes[route.Path] = route
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, ok := r.Routes[req.URL.Path]

	if !ok && !strings.HasPrefix(req.URL.Path, "/status/") {
		http.NotFound(w, req)
		return
	}

	var magicReq MagicRequest

	// If it's a user-defined route, handle it
	if ok && route.Method == req.Method {
		// Try to parse JSON payload first.
		if req.Header.Get("Content-Type") == "application/json" {
			defer req.Body.Close()
			err := json.NewDecoder(req.Body).Decode(&magicReq)
			if err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}
		}

		for key, value := range magicReq.ResponseHeaders {
			w.Header().Set(key, value)
		}
		w.WriteHeader(route.StatusCode)

		responseBody, err := json.Marshal(magicReq.ResponseBody)
		if err != nil {
			http.Error(w, "Error processing response body", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(responseBody)
		if err != nil {
			http.Error(w, "Error writing response body", http.StatusInternalServerError)
			return
		}
		return
	}

	// Parse status code from the magic route.
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid magic route", http.StatusBadRequest)
		return
	}
	statusCode, err := strconv.Atoi(parts[2])
	if err != nil || statusCode < 100 || statusCode > 599 {
		http.Error(w, "Invalid status code", http.StatusBadRequest)
		return
	}

	// Try to parse query params or JSON payload into magicReq.
	if err := r.parseRequestIntoMagicReq(req, &magicReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle magic route.
	for key, value := range magicReq.ResponseHeaders {
		w.Header().Set(key, value)
	}
	w.WriteHeader(statusCode)
	if magicReq.ResponseBody != "" {
		responseBody, err := json.Marshal(magicReq.ResponseBody)
		if err != nil {
			http.Error(w, "Error processing response body", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(responseBody)
		if err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func (r *Router) LoadRoutesFromJSON(data []byte) error {
	var routes []Route
	if err := json.Unmarshal(data, &routes); err != nil {
		return err
	}

	for _, route := range routes {
		newRoute := route
		r.AddRoute(&newRoute)
	}

	return nil
}
