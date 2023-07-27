package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/iamthen0ise/faux/internal/api"
	"github.com/iamthen0ise/faux/internal/applogger"
	"github.com/iamthen0ise/faux/internal/args"

	"golang.org/x/term"
)

func terminalSizeOK() bool {
	rows, cols, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// If there's an error, assume the terminal is big enough.
		return true
	}

	// Check if the terminal size is 50x80 or larger
	return rows >= 50 && cols >= 70
}

func printWelcomeMessage(port int) {
	asciiArt := `
A FRIENDLY HTTP MOCKING SERVER FOR DEVELOPMENT AND TESTING
	
▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄	▄▄▄▄▄▄▄  ▄  ▄▄▄ ▄  ▄  ▄▄▄▄▄▄▄  
███████╗░█████╗░██╗░░░██╗██╗░░██╗	█ ▄▄▄ █ ▄▄█▀█▄██▀▀█▀▀ █ ▄▄▄ █  
██╔════╝██╔══██╗██║░░░██║╚██╗██╔╝	█ ███ █  ▀█▀ ▀█▄▄▄█▄  █ ███ █  
██╔════╝██╔══██╗██║░░░██║╚██╗██╔╝	█▄▄▄▄▄█ ▄▀▄ ▄▀▄▀▄▀▄▀█ █▄▄▄▄▄█  
█████╗░░███████║██║░░░██║░╚███╔╝░	▄▄▄ ▄▄▄▄█▀█▄▀▀██ █ ▀▄▄▄   ▄    
█████╗░░███████║██║░░░██║░╚███╔╝░	█▀  ▀▄▄█▄▄▀▀ ▄▀ ▀▀  ▀▄█ ▄▀▄▄█  
██╔══╝░░██╔══██║██║░░░██║░██╔██╗░	▀  █▄▀▄▀█  ▀▄  ▀▀█▀█▄▄▄▀▀▄ █▄  
██╔══╝░░██╔══██║██║░░░██║░██╔██╗░	 ▀▄ ▄█▄▄▀█▄██▀▄ ▀▄  ▀▀█ ▄█ ▄█  
██║░░░░░██║░░██║╚██████╔╝██╔╝╚██╗	▄▄ ▀▄▀▄ ██▀▄▀▀█▀▄█▄▄█▀█▄ █ █▄  
╚═╝░░░░░╚═╝░░╚═╝░╚═════╝░╚═╝░░╚═╝	▄▀▀▀ █▄▀ █▄▀ ▄▀ █▀█  █▀▄▄▀▀▄█  
╚═╝░░░░░╚═╝░░╚═╝░╚═════╝░╚═╝░░╚═╝	▄▀▄███▄  ██▀▄   ██ ██████▀ ▀   
					▄▄▄▄▄▄▄ ██▀██▀▄ ▀ ▄██ ▄ █▄▀██  
					█ ▄▄▄ █ ███▄▀▀█▄██▀ █▄▄▄█▀ █▄  
					█ ███ █ ▄ ▀▀ ▄▀  ██▄▀  ██▄▀▀█  
					█▄▄▄▄▄█ █▄ ▀▄  ▀██▀███ ▄█  █▄ 

RUN WITH --quiet-start TO HIDE THIS^^^
	`
	fmt.Println(asciiArt)
	fmt.Println("Welcome to FAUX, your friendly mock server.")
	fmt.Println("To call Magic Route, use the following format:")
	fmt.Println("http://<host>:<port>/status/<statusCode>?responseTime=<delayInMilliseconds>")

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	// Generate a random method
	rand.Seed(time.Now().UnixNano())
	randomMethod := methods[rand.Intn(len(methods))]

	// Generate a random status code between 200 and 500
	statusCode := rand.Intn(301) + 200 // random integer between 200 and 500

	// Set example headers and body
	examlePayload := `{"response_headers":{"Content-Type":"application/json"},"response_body":"{\"message\": \"Hello from Faux\"}"}`

	fmt.Printf("\nExample curl command:\n")
	fmt.Printf("curl -X %s -d '%s' http://localhost:%d/status/%d\n", randomMethod, examlePayload, port, statusCode)
}

func main() {
	appConfig := &args.AppConfig{}
	args.ParseInput(appConfig)

	if !appConfig.QuietStart && terminalSizeOK() {
		printWelcomeMessage(appConfig.Port)
	}

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
