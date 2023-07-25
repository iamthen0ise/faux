package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/iamthen0ise/faux/pkg/api"
	applog "github.com/iamthen0ise/faux/pkg/log"
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

	// Load routes from a JSON file.
	if *routesFilePath != "" {
		routesData, err := ioutil.ReadFile(*routesFilePath)
		if err != nil {
			log.Fatalf("Error reading routes file: %v", err)
		}

		err = router.LoadRoutesFromJSON(routesData)
		if err != nil {
			log.Fatalf("Error loading routes: %v", err)
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
