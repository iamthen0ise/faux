package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock OpenAPI Document
const mockOpenAPIDocument = `
openapi: 3.0.0
info:
  title: Sample API
  version: 0.1.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: Successful response
`

func TestParseOpenAPISchemaPositive(t *testing.T) {
	swagger, err := ParseOpenAPISchema(mockOpenAPIDocument)
	assert.NoError(t, err)
	assert.NotNil(t, swagger)
}

func TestParseOpenAPISchemaNegative(t *testing.T) {
	invalidDoc := `invalid: openapi: doc`
	_, err := ParseOpenAPISchema(invalidDoc)
	assert.Error(t, err)
}

func TestMapOpenAPIRoutes(t *testing.T) {
	swagger, _ := ParseOpenAPISchema(mockOpenAPIDocument)
	routes, err := MapOpenAPIRoutes(swagger)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, "/test", routes[0].Path)
	assert.Equal(t, "GET", routes[0].Method)
}

func TestLoadRoutesFromOpenAPI(t *testing.T) {
	router := NewRouter()
	err := router.LoadRoutesFromOpenAPI(mockOpenAPIDocument)

	assert.NoError(t, err)

	// Validate that the route exists in the router
	route, exists := router.Routes["/test"]
	assert.True(t, exists)
	assert.NotNil(t, route)
}
