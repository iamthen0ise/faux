package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OpenAPISpec struct {
	OpenAPI string                     `json:"openapi"`
	Info    OpenAPIInfo                `json:"info"`
	Paths   map[string]OpenAPIPathItem `json:"paths"`
}

type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
}

type OpenAPIResponse struct {
	Description string `json:"description"`
}

func (r *Router) GenerateOpenAPI() OpenAPISpec {
	spec := OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:   "Magic Mock API",
			Version: "1.0",
		},
		Paths: make(map[string]OpenAPIPathItem),
	}

	for _, route := range r.Routes {
		operation := OpenAPIOperation{
			Summary:     "Auto-generated mock route",
			Description: fmt.Sprintf("Handles %s requests for %s", route.Method, route.Path),
			Responses: map[string]OpenAPIResponse{
				strconv.Itoa(route.StatusCode): {
					Description: http.StatusText(route.StatusCode),
				},
			},
		}

		pathItem := OpenAPIPathItem{}
		switch strings.ToLower(route.Method) {
		case "get":
			pathItem.Get = &operation
		case "post":
			pathItem.Post = &operation
		case "put":
			pathItem.Put = &operation
		case "delete":
			pathItem.Delete = &operation
		}

		spec.Paths[route.Path] = pathItem
	}

	// Add MagicRoute specifics
	spec.Paths["/status/{statusCode}"] = OpenAPIPathItem{
		Get: &OpenAPIOperation{
			Summary:     "MagicRoute for dynamic responses",
			Description: "Generates a response dynamically based on request content. The status code can be any value between 100 and 599.",
			Responses: map[string]OpenAPIResponse{
				"default": {
					Description: "Dynamic response based on provided request content.",
				},
			},
		},
	}

	return spec
}

func (r *Router) OpenAPIHandler(w http.ResponseWriter, req *http.Request) {
	openAPISpec := r.GenerateOpenAPI()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openAPISpec)
}
