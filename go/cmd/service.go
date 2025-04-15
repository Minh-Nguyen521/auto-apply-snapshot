package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auto-apply-snapshot/snapshot"
)

// runService starts the automated snapshot service
func runService(manager *snapshot.Manager) {
	log.Println("MongoDB Snapshot Service started")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a ticker for scheduled snapshots (daily at 2 AM)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run initial snapshot if it's time
	if shouldRunSnapshot() {
		if err := manager.CreateSnapshot(); err != nil {
			log.Printf("Failed to create initial snapshot: %v", err)
		} else {
			log.Println("Initial snapshot created successfully")
		}
	}

	// Main service loop
	for {
		select {
		case <-ticker.C:
			if shouldRunSnapshot() {
				if err := manager.CreateSnapshot(); err != nil {
					log.Printf("Failed to create scheduled snapshot: %v", err)
				} else {
					log.Println("Scheduled snapshot created successfully")
				}
			}
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down", sig)
			return
		}
	}
}

// shouldRunSnapshot checks if it's time to run a snapshot (2 AM)
func shouldRunSnapshot() bool {
	now := time.Now()
	return now.Hour() == 2 && now.Minute() < 5 // Run between 2:00 and 2:05
}
