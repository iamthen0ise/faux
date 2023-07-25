package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamthen0ise/faux/internal/api"
	applog "github.com/iamthen0ise/faux/internal/log"
)

func main() {
	authToken := flag.String("t", "", "authentication token")

	flag.Parse()

	routesFilePath := flag.String("routes", "", "Path to JSON file containing routes")
	// Initialize a new Logger.
	logger := applog.NewLogger("[{{.Time}}] {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}\n", true)

	authMiddleware := &api.AuthMiddleware{Token: *authToken}

	router := api.NewRouter()
	authMiddleware.Next = router

	if *routesFilePath != "" {
		fileInfo, err := os.Stat(*routesFilePath)
		if err != nil {
			log.Fatalf("Failed to get the file info: %v", err)
		}

		if fileInfo.IsDir() {
			// Load routes from a directory
			err = router.LoadRoutesFromDir(*routesFilePath)
			if err != nil {
				log.Fatalf("Failed to load routes: %v", err)
			}
		} else {
			// Load routes from specific files
			err := router.LoadRoutesFromFiles([]string{*routesFilePath})
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
	log.Fatal(http.ListenAndServe(":8080", nil))
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
