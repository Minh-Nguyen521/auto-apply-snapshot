# MongoDB Auto Snapshot (Python)

This program automatically creates and downloads snapshots of your MongoDB databases on a scheduled basis, with the ability to restore snapshots when needed.

## Features

- Automated daily snapshots (configurable schedule)
- Exports all databases and collections
- Detailed logging
- Error handling and retry mechanism
- Configurable backup directory
- Snapshot restoration capability
- Command-line interface for easy management

## Setup

1. Install Python 3.8 or later

2. Clone the repository:
```bash
git clone https://github.com/yourusername/auto-apply-snapshot.git
cd auto-apply-snapshot/python
```

3. Install the required dependencies:
```bash
pip install -r requirements.txt
```

4. Copy the environment template and configure your settings:
```bash
cp .env.example .env
```

5. Edit the `.env` file with your MongoDB connection details:
- Set your `MONGODB_URI` with proper credentials
- Optionally configure `BACKUP_DIR` for custom backup location

## Usage

### Automated Snapshot Service

Run the automated snapshot service:
```bash
python mongodb_snapshot.py
```

The service will:
- Create snapshots daily at 2 AM (configurable in the code)
- Store backups in the specified directory
- Log all activities to `mongodb_snapshot.log`

### Command Line Interface

Use the CLI to manage snapshots:

1. List available snapshots:
```bash
python mongodb_cli.py list
```

2. Create a snapshot manually:
```bash
python mongodb_cli.py create
```

3. Restore a snapshot:
```bash
python mongodb_cli.py restore --snapshot YYYYMMDD_HHMMSS
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

Logs are written to both:
- Console output
- `mongodb_snapshot.log` file

## Error Handling

The script includes:
- Connection error handling
- Automatic retry mechanism
- Detailed error logging
- Safe snapshot restoration with validation 