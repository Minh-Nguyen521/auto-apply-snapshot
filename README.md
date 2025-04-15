# MongoDB Auto Snapshot (Go)

This program automatically creates and downloads snapshots of your MongoDB databases on a scheduled basis, with the ability to restore snapshots when needed. Written in Go for better performance and concurrency.

## Setup

1. Install Go 1.16 or later

2. Clone the repository:
```bash
git clone https://github.com/yourusername/auto-apply-snapshot.git
cd auto-apply-snapshot
```

3. Install dependencies:
```bash
go mod download
```

4. Copy the configuration template and configure your settings:
```bash
cp config.yaml.example config.yaml
```

5. Edit the `config.yaml` file with your MongoDB connection details:
- Set your `mongodb_uri` with proper credentials
- Optionally configure `backup_dir` for custom backup location

## Usage

### Automated Snapshot Service

Run the automated snapshot service:
```bash
go run main.go -action service
```

The service will:
- Create snapshots daily at 2 AM (configurable in the code)
- Store backups in the specified directory
- Log all activities to the console

### Command Line Interface

Use the CLI to manage snapshots:

1. List available snapshots:
```bash
go run main.go -action list
```

2. Create a snapshot manually:
```bash
go run main.go -action create
```

3. Restore a snapshot:
```bash
go run main.go -action restore -snapshot YYYYMMDD_HHMMSS
```

## Backup Structure

Backups are organized as follows:
```
backups/
└── YYYYMMDD_HHMMSS/
    └── database_name/
        └── collection_name.json
```

## Logging

Logs are written to the console with timestamps and log levels.

## Error Handling

The program includes:
- Connection error handling
- Automatic retry mechanism
- Detailed error logging
- Safe snapshot restoration with validation
- Context-based timeouts for operations

## Building

To build the executable:
```bash
go build -o mongodb-snapshot
```

Then run it:
```bash
./mongodb-snapshot -action service
```
