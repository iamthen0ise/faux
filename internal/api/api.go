package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/iamthen0ise/faux/internal/throttling"
)

type Route struct {
	Path            string            `json:"path"`
	Method          string            `json:"method"`
	StatusCode      int               `json:"status_code"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseBody    interface{}       `json:"response_body,omitempty"`
	Lambda          int               `json:"-"`
	AuthRequired    bool              `json:"auth_required,omitempty"`
	ThrottlingLow   int               `json:"throttling_low,omitempty"`
	ThrottlingHigh  int               `json:"throttling_hi,omitempty"`
	RateLimitPerMin float32           `json:"rate_limit_per_min,omitempty"`
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
	Lambda          int               `json:"-"`
	AuthRequired    bool              `json:"auth_required,omitempty"`
	ThrottlingLow   int               `json:"throttling_low,omitempty"`
	ThrottlingHigh  int               `json:"throttling_hi,omitempty"`
	RateLimitPerMin float32           `json:"rate_limit_per_min,omitempty"`
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
	// Check for the specific /openapi route
	if req.URL.Path == "/openapi" {
		r.OpenAPIHandler(w, req)
		return
	}

	route, ok := r.Routes[req.URL.Path]
	if !ok && !strings.HasPrefix(req.URL.Path, "/status/") {
		http.NotFound(w, req)
		return
	}

	var magicReq MagicRequest

	if ok && route.Method == req.Method {
		throttlingMiddleware := throttling.ThrottlingMiddleware(route.ThrottlingLow, route.ThrottlingHigh)
		rateLimitMiddleware := throttling.RateLimitMiddleware(route.RateLimitPerMin)
		routeHandler := r.handleDefinedRoute(route, &magicReq)
		handler := throttlingMiddleware(rateLimitMiddleware(routeHandler))
		handler.ServeHTTP(w, req)
	} else {
		r.handleMagicRoute(w, req, &magicReq)
	}
}

func (r *Router) handleDefinedRoute(route *Route, magicReq *MagicRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") == "application/json" {
			if err := json.NewDecoder(req.Body).Decode(magicReq); err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}
			defer req.Body.Close()
		}

		setHeaders(w, magicReq.ResponseHeaders)
		writeResponse(w, route.StatusCode, magicReq.ResponseBody)
	})
}

func (r *Router) handleMagicRoute(w http.ResponseWriter, req *http.Request, magicReq *MagicRequest) {
	statusCode, err := r.parseMagicRoute(req.URL.Path)
	if err != nil {
		http.Error(w, "Invalid magic route", http.StatusBadRequest)
		return
	}

	if err := r.parseRequestIntoMagicReq(req, magicReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setHeaders(w, magicReq.ResponseHeaders)
	writeResponse(w, statusCode, magicReq.ResponseBody)
}

func (r *Router) parseMagicRoute(path string) (int, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, errors.New("Invalid magic route")
	}
	return strconv.Atoi(parts[2])
}
func setHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	w.WriteHeader(statusCode)

	if responseBody != nil {
		body, err := json.Marshal(responseBody)
		if err != nil {
			http.Error(w, "Error processing response body", http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(body); err != nil {
			http.Error(w, "Error writing response body", http.StatusInternalServerError)
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
