package args

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYaml(t *testing.T) {
	testCases := []struct {
		name        string
		configYaml  string
		expected    *AppConfig
		expectError bool
	}{
		{
			name: "Valid YAML",
			configYaml: `
authToken: s3cr3tt0k3n
routesPath: /path/to/routes.json
colorize: true
logFormat: "[{{.Time}}] {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}\n"
host: localhost
port: 8080`,
			expected: &AppConfig{
				AuthToken:  "s3cr3tt0k3n",
				RoutesPath: "/path/to/routes.json",
				Colorize:   true,
				LogFormat:  "[{{.Time}}] {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}\n",
				Host:       "localhost",
				Port:       8080,
			},
			expectError: false,
		},
		{
			name:        "Invalid YAML",
			configYaml:  `invalid_yaml`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appConfig := NewAppConfig()

			// Create temporary file
			tempFile, err := ioutil.TempFile(os.TempDir(), "*.yaml")
			if err != nil {
				t.Fatalf("Cannot create temporary file: %s", err)
			}

			// Write YAML to temporary file
			_, err = tempFile.WriteString(tc.configYaml)
			if err != nil {
				t.Fatalf("Failed to write to temporary file: %s", err)
			}

			// Defer the removal of the temporary file
			defer func() {
				err := os.Remove(tempFile.Name())
				if err != nil {
					t.Logf("Warning: Error removing temporary file: %s", err)
				}
			}()

			err = ParseYaml(tempFile.Name(), appConfig)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, appConfig)
			}
		})

	}
}

func TestParseInput(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name             string
		args             []string
		configFilePath   string
		configYaml       string
		expected         *AppConfig
		expectParseError bool
	}{
		{
			name: "CLI Flags",
			args: []string{"-token=s3cr3tt0k3n", "-routes=/path/to/routes.json", "-colorize=true", "-log-format='{{.Time}} {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}'", "-host=localhost", "-port=8080"},
			expected: &AppConfig{
				AuthToken:  "s3cr3tt0k3n",
				RoutesPath: "/path/to/routes.json",
				Colorize:   true,
				LogFormat:  "'{{.Time}} {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}'",
				Host:       "localhost",
				Port:       8080,
			},
			expectParseError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock command line arguments
			os.Args = append([]string{"cmd"}, tc.args...)
			// Create an instance of AppConfig
			appConfig := NewAppConfig()

			// If a YAML file is provided in the test case, use it
			if tc.configFilePath != "" {
				os.Args = append(os.Args, "-config="+tc.configFilePath)
			}

			// Parse command line arguments or YAML file into the AppConfig
			err := ParseInput(appConfig)

			// Check if the function returns an error as expected
			if tc.expectParseError {
				assert.Error(t, err)
			} else {
				// If no error is expected, check that the function does not return an error and that AppConfig is as expected
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, appConfig)
			}
		})
	}
}
