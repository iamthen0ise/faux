package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadOpenAPIFromFile loads the OpenAPI schema from a file.
func LoadOpenAPIFromFile(filepath string) (*openapi3.T, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return ParseOpenAPIFromText(string(data))
}

// LoadOpenAPIFromURL loads the OpenAPI schema from a given URL.
func LoadOpenAPIFromURL(url string) (*openapi3.T, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseOpenAPIFromText(string(data))
}

// ParseOpenAPIFromText parses the OpenAPI schema from raw text input.
func ParseOpenAPIFromText(input string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	return loader.LoadFromData([]byte(input))
}

func ParseOpenAPISchema(input string) (*openapi3.T, error) {
	if isURL(input) {
		return LoadOpenAPIFromURL(input)
	} else if isFile(input) {
		return LoadOpenAPIFromFile(input)
	}
	return ParseOpenAPIFromText(input)
}

// isURL checks if the given string is a URL.
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// isFile checks if the given string is a valid file path.
func isFile(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}
func MapOpenAPIRoutes(swagger *openapi3.T) ([]Route, error) {
	var routes []Route

	for path, pathItem := range swagger.Paths {
		for method := range pathItem.Operations() {
			route := Route{
				Path:       path,
				Method:     method,
				StatusCode: 200,
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}
func (r *Router) LoadRoutesFromOpenAPI(document string) error {
	swagger, err := ParseOpenAPISchema(document)
	if err != nil {
		return err
	}

	routes, err := MapOpenAPIRoutes(swagger)
	if err != nil {
		return err
	}

	for _, route := range routes {
		r.AddRoute(&route)
	}

	return nil
}
