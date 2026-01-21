package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"singbox-web/internal/api"
	"singbox-web/internal/storage"
)

func main() {
	// Get data directory
	dataDir := "data"

	// Use executable directory for production
	if os.Getenv("DEV") != "1" {
		execPath, err := os.Executable()
		if err == nil {
			dataDir = filepath.Join(filepath.Dir(execPath), "data")
		}
	}

	log.Printf("Using data directory: %s", dataDir)

	// Initialize database
	if err := storage.InitDatabase(dataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get port from settings
	port, err := storage.GetSetting("web_port")
	if err != nil {
		port = "60017"
	}

	// Setup router
	r := api.SetupRouter()

	// Start server
	log.Printf("singbox-web starting on http://localhost:%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}
