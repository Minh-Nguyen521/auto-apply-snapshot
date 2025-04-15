package snapshot

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

// Config holds the application configuration
type Config struct {
	MongoDBURI string `yaml:"mongodb_uri"`
	BackupDir  string `yaml:"backup_dir"`
}

// Manager handles MongoDB snapshot operations
type Manager struct {
	config Config
	client *mongo.Client
}

// NewManager creates a new snapshot manager
func NewManager() (*Manager, error) {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create manager
	manager := &Manager{
		config: config,
	}

	// Connect to MongoDB
	if err := manager.connect(); err != nil {
		return nil, err
	}

	return manager, nil
}

// loadConfig loads the configuration from config.yaml
func loadConfig() (Config, error) {
	var config Config

	// Default values
	config.BackupDir = "backups"

	// Try to load from file
	data, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables if set
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		config.MongoDBURI = uri
	}
	if dir := os.Getenv("BACKUP_DIR"); dir != "" {
		config.BackupDir = dir
	}

	// Validate configuration
	if config.MongoDBURI == "" {
		return config, fmt.Errorf("MongoDB URI is required")
	}

	return config, nil
}

// connect establishes a connection to MongoDB
func (m *Manager) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.config.MongoDBURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	log.Println("Successfully connected to MongoDB")
	return nil
}

// CreateSnapshot creates a snapshot of all databases
func (m *Manager) CreateSnapshot() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Create timestamp for backup folder
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(m.config.BackupDir, timestamp)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get list of databases
	databases, err := m.client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	for _, dbName := range databases {
		// Skip system databases
		if dbName == "admin" || dbName == "local" {
			continue
		}

		log.Printf("Creating snapshot for database: %s", dbName)

		// Create database backup directory
		dbBackupPath := filepath.Join(backupPath, dbName)
		if err := os.MkdirAll(dbBackupPath, 0755); err != nil {
			return fmt.Errorf("failed to create database backup directory: %w", err)
		}

		// Get collections in the database
		db := m.client.Database(dbName)
		collections, err := db.ListCollectionNames(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to list collections: %w", err)
		}

		for _, collectionName := range collections {
			collection := db.Collection(collectionName)

			// Find all documents in the collection
			cursor, err := collection.Find(ctx, bson.M{})
			if err != nil {
				return fmt.Errorf("failed to find documents: %w", err)
			}
			defer cursor.Close(ctx)

			// Read all documents
			var documents []bson.M
			if err := cursor.All(ctx, &documents); err != nil {
				return fmt.Errorf("failed to read documents: %w", err)
			}

			// Save to file
			outputFile := filepath.Join(dbBackupPath, collectionName+".json")
			file, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()

			// Write documents to file
			for _, doc := range documents {
				docBytes, err := bson.MarshalExtJSON(doc, true, true)
				if err != nil {
					return fmt.Errorf("failed to marshal document: %w", err)
				}

				if _, err := file.Write(docBytes); err != nil {
					return fmt.Errorf("failed to write document: %w", err)
				}
				if _, err := file.WriteString("\n"); err != nil {
					return fmt.Errorf("failed to write newline: %w", err)
				}
			}

			log.Printf("Exported %d documents from %s.%s", len(documents), dbName, collectionName)
		}
	}

	log.Printf("Snapshot completed successfully at %s", timestamp)
	return nil
}

// ListSnapshots returns a list of available snapshots
func (m *Manager) ListSnapshots() ([]string, error) {
	entries, err := ioutil.ReadDir(m.config.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, entry.Name())
		}
	}

	// Sort snapshots in descending order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(snapshots)))
	return snapshots, nil
}

// RestoreSnapshot restores a snapshot to MongoDB
func (m *Manager) RestoreSnapshot(snapshotName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	snapshotPath := filepath.Join(m.config.BackupDir, snapshotName)
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %s does not exist", snapshotName)
	}

	// Get list of databases in the snapshot
	entries, err := ioutil.ReadDir(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dbName := entry.Name()
		dbPath := filepath.Join(snapshotPath, dbName)
		log.Printf("Restoring database: %s", dbName)

		// Get or create database
		db := m.client.Database(dbName)

		// Get list of collection files
		files, err := ioutil.ReadDir(dbPath)
		if err != nil {
			return fmt.Errorf("failed to read database directory: %w", err)
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			collectionName := strings.TrimSuffix(file.Name(), ".json")
			collectionPath := filepath.Join(dbPath, file.Name())

			log.Printf("Restoring collection: %s", collectionName)

			// Clear existing collection
			collection := db.Collection(collectionName)
			if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
				return fmt.Errorf("failed to clear collection: %w", err)
			}

			// Read file
			data, err := ioutil.ReadFile(collectionPath)
			if err != nil {
				return fmt.Errorf("failed to read collection file: %w", err)
			}

			// Parse documents
			lines := strings.Split(string(data), "\n")
			var documents []interface{}

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				var doc bson.M
				if err := bson.UnmarshalExtJSON([]byte(line), true, &doc); err != nil {
					log.Printf("Error parsing document: %v", err)
					continue
				}

				documents = append(documents, doc)
			}

			// Insert documents
			if len(documents) > 0 {
				if _, err := collection.InsertMany(ctx, documents); err != nil {
					return fmt.Errorf("failed to insert documents: %w", err)
				}
				log.Printf("Restored %d documents to %s.%s", len(documents), dbName, collectionName)
			}
		}
	}

	log.Printf("Snapshot %s restored successfully", snapshotName)
	return nil
}

// Close closes the MongoDB connection
func (m *Manager) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}
