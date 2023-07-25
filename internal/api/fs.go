package api

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func (r *Router) LoadRoutesFromFiles(files []string) error {
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if err := r.LoadRoutesFromJSON(data); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) LoadRoutesFromDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}

		if err := r.LoadRoutesFromJSON(data); err != nil {
			return err
		}
	}
	return nil
}

// WatchRoutes sets up a watcher on the routes file or directory.
func WatchRoutes(router *Router, routesFilePath string) {
	// Initialize watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Create done channel to signal events.
	done := make(chan bool)

	// Create event handler.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Check if event is caused by a file write.
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Modified file:", event.Name, " reloading...")

					// Load routes again.
					err := router.LoadRoutesFromFiles([]string{event.Name})
					if err != nil {
						log.Printf("Could not load routes: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Add file or directory to watcher.
	err = watcher.Add(routesFilePath)
	if err != nil {
		log.Printf("Could not reload routes: %v\n", err)
	}

	// Block until done.
	<-done
}
