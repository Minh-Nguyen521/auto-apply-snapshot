package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/auto-apply-snapshot/src/snapshot"
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
	defer manager.Close()

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
