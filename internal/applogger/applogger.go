package applogger

import (
	"bytes"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	logFormat *template.Template
	colorize  bool
}

func NewLogger(format string, colorize bool) *Logger {
	logTemplate, err := template.New("log").Parse(format)
	if err != nil {
		panic(err)
	}

	return &Logger{
		logFormat: logTemplate,
		colorize:  colorize,
	}
}

func (l *Logger) LogRequest(r *http.Request, statusCode int, responseTime time.Duration) {
	var colorPrinter *color.Color
	switch {
	case statusCode >= 100 && statusCode < 200:
		colorPrinter = color.New(color.FgWhite)
	case statusCode >= 200 && statusCode < 300:
		colorPrinter = color.New(color.FgGreen)
	case statusCode >= 300 && statusCode < 400:
		colorPrinter = color.New(color.FgYellow)
	case statusCode >= 400 && statusCode < 500:
		colorPrinter = color.New(color.FgHiRed) // High-intensity red as orange.
	case statusCode >= 500:
		colorPrinter = color.New(color.FgRed)
	default:
		colorPrinter = color.New(color.FgWhite)
	}

	logData := struct {
		Time         string
		Method       string
		StatusCode   int
		Path         string
		ResponseTime time.Duration
	}{
		Time:         time.Now().Format("2006-01-02 15:04:05"),
		Method:       r.Method,
		StatusCode:   statusCode,
		Path:         r.URL.Path,
		ResponseTime: responseTime,
	}

	var logBuffer bytes.Buffer
	if err := l.logFormat.Execute(&logBuffer, logData); err != nil {
		log.Fatalf("Error executing log template: %v", err)
	}

	logStr := logBuffer.String()

	if l.colorize {
		colorPrinter.Println(logStr)
	} else {
		log.Println(logStr)
	}
}
