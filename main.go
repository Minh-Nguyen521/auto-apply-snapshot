package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auto-apply-snapshot/snapshot"
)

func main() {
	// Parse command line flags
	action := flag.String("action", "service", "Action to perform: service, create, restore, or list")
	snapshotName := flag.String("snapshot", "", "Snapshot name for restore action")
	flag.Parse()

	// Create snapshot manager
	manager, err := snapshot.NewManager()
	if err != nil {
		log.Fatalf("Failed to create snapshot manager: %v", err)
	}

	// Execute the requested action
	switch *action {
	case "service":
		runService(manager)
	case "create":
		if err := manager.CreateSnapshot(); err != nil {
			log.Fatalf("Failed to create snapshot: %v", err)
		}
		fmt.Println("Snapshot created successfully")
	case "restore":
		if *snapshotName == "" {
			log.Fatal("Snapshot name is required for restore action")
		}
		if err := manager.RestoreSnapshot(*snapshotName); err != nil {
			log.Fatalf("Failed to restore snapshot: %v", err)
		}
		fmt.Printf("Snapshot %s restored successfully\n", *snapshotName)
	case "list":
		snapshots, err := manager.ListSnapshots()
		if err != nil {
			log.Fatalf("Failed to list snapshots: %v", err)
		}
		if len(snapshots) == 0 {
			fmt.Println("No snapshots found")
		} else {
			fmt.Println("\nAvailable snapshots:")
			for _, s := range snapshots {
				fmt.Printf("- %s\n", s)
			}
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func runService(manager *snapshot.Manager) {
	fmt.Println("MongoDB Snapshot Service started")

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
