package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamthen0ise/faux/internal/api"
	"github.com/iamthen0ise/faux/internal/applogger"
	"github.com/iamthen0ise/faux/internal/args"
)

// AppConfig encapsulates the command line arguments

func main() {
	appConfig := &args.AppConfig{}
	args.ParseInput(appConfig)

	// Initialize a new Logger.
	logger := applogger.NewLogger("[{{.Time}}] {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}\n", appConfig.Colorize)

	authMiddleware := &api.AuthMiddleware{Token: appConfig.AuthToken}
	router := api.NewRouter()
	authMiddleware.Next = router

	if appConfig.RoutesPath != "" {
		fileInfo, err := os.Stat(appConfig.RoutesPath)
		if err != nil {
			log.Fatalf("Failed to get the file info: %v", err)
		}

		if fileInfo.IsDir() {
			// Load routes from a directory
			err = router.LoadRoutesFromDir(appConfig.RoutesPath)
			if err != nil {
				log.Fatalf("Failed to load routes: %v", err)
			}
		} else {
			// Load routes from specific files
			err := router.LoadRoutesFromFiles([]string{appConfig.RoutesPath})
			if err != nil {
				log.Fatalf("Failed to load routes: %v", err)
			}
		}
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := statusRecorder{ResponseWriter: w}
		router.ServeHTTP(&rec, r)

		duration := time.Since(start)
		logger.LogRequest(r, rec.status, duration)
	}))

	// Start the HTTP server.
	log.Fatal(http.ListenAndServe(appConfig.Host+":"+fmt.Sprint(appConfig.Port), nil))
}

// statusRecorder is an HTTP ResponseWriter that captures the status code written to it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code written.
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
