package args

import (
	"flag"
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ConfigFile string
	AuthToken  string `yaml:"authToken"`
	RoutesPath string `yaml:"routesPath"`
	Colorize   bool   `yaml:"colorize"`
	LogFormat  string `yaml:"logFormat"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	QuietStart bool
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func ParseInput(appConfig *AppConfig) error {
	flag.StringVar(&appConfig.ConfigFile, "config", "", "Path to YAML config file")
	flag.StringVar(&appConfig.AuthToken, "token", "", "Authentication token")
	flag.StringVar(&appConfig.RoutesPath, "routes", "", "Path to JSON file containing routes")
	flag.BoolVar(&appConfig.Colorize, "colorize", true, "Enable log colorizing")
	flag.StringVar(&appConfig.LogFormat, "log-format", "{{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}", "Log format")
	flag.StringVar(&appConfig.Host, "host", "localhost", "Application host")
	flag.IntVar(&appConfig.Port, "port", 8080, "Application port")
	flag.BoolVar(&appConfig.QuietStart, "quiet-start", false, "Mute any welcome messages")

	flag.Parse()

	// If config file provided, then parse it and update the appConfig fields
	if appConfig.ConfigFile != "" {
		err := ParseYaml(appConfig.ConfigFile, appConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func ParseYaml(configFilePath string, appConfig *AppConfig) error {
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configData, appConfig)
	if err != nil {
		return err
	}

	return nil
}
