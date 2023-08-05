package api

import (
	"reflect"
	"testing"
)

func TestGenerateOpenAPI(t *testing.T) {
	tests := []struct {
		name         string
		routes       []*Route
		expectedSpec OpenAPISpec
	}{
		{
			name: "single route",
			routes: []*Route{
				{
					Path:       "/test",
					Method:     "GET",
					StatusCode: 200,
				},
			},
			expectedSpec: OpenAPISpec{
				OpenAPI: "3.0.0",
				Info: OpenAPIInfo{
					Title:   "Magic Mock API",
					Version: "1.0",
				},
				Paths: map[string]OpenAPIPathItem{
					"/test": {
						Get: &OpenAPIOperation{
							Summary:     "Auto-generated mock route",
							Description: "Handles GET requests for /test",
							Responses: map[string]OpenAPIResponse{
								"200": {
									Description: "OK",
								},
							},
						},
					},
					"/status/{statusCode}": {
						Get: &OpenAPIOperation{
							Summary:     "MagicRoute for dynamic responses",
							Description: "Generates a response dynamically based on request content. The status code can be any value between 100 and 599.",
							Responses: map[string]OpenAPIResponse{
								"default": {
									Description: "Dynamic response based on provided request content.",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			for _, route := range tt.routes {
				router.AddRoute(route)
			}

			gotSpec := router.GenerateOpenAPI()

			if !reflect.DeepEqual(gotSpec, tt.expectedSpec) {
				t.Errorf("GenerateOpenAPI() got:\n%v\nwant:\n%v", gotSpec, tt.expectedSpec)
			}
		})
	}
}
