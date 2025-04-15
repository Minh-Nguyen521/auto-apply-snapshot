import os
import logging
import schedule
import time
from datetime import datetime
from pymongo import MongoClient
from dotenv import load_dotenv

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('mongodb_snapshot.log'),
        logging.StreamHandler()
    ]
)

# Load environment variables
load_dotenv()

class MongoDBSnapshot:
    def __init__(self):
        self.mongo_uri = os.getenv('MONGODB_URI')
        self.backup_dir = os.getenv('BACKUP_DIR', 'backups')
        self.client = None
        
        # Create backup directory if it doesn't exist
        if not os.path.exists(self.backup_dir):
            os.makedirs(self.backup_dir)

    def connect(self):
        """Establish connection to MongoDB"""
        try:
            self.client = MongoClient(self.mongo_uri)
            # Test the connection
            self.client.server_info()
            logging.info("Successfully connected to MongoDB")
        except Exception as e:
            logging.error(f"Failed to connect to MongoDB: {str(e)}")
            raise

    def create_snapshot(self):
        """Create and download MongoDB snapshot"""
        try:
            if not self.client:
                self.connect()

            # Get list of databases
            databases = self.client.list_database_names()
            
            # Create timestamp for backup folder
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            backup_path = os.path.join(self.backup_dir, timestamp)
            
            if not os.path.exists(backup_path):
                os.makedirs(backup_path)

            for db_name in databases:
                if db_name not in ['admin', 'local']:  # Skip system databases
                    logging.info(f"Creating snapshot for database: {db_name}")
                    
                    # Create database backup directory
                    db_backup_path = os.path.join(backup_path, db_name)
                    if not os.path.exists(db_backup_path):
                        os.makedirs(db_backup_path)

                    # Get collections in the database
                    db = self.client[db_name]
                    collections = db.list_collection_names()

                    for collection in collections:
                        # Export collection data
                        collection_data = list(db[collection].find())
                        
                        # Save to file
                        output_file = os.path.join(db_backup_path, f"{collection}.json")
                        with open(output_file, 'w') as f:
                            for document in collection_data:
                                f.write(str(document) + '\n')
                        
                        logging.info(f"Exported {len(collection_data)} documents from {db_name}.{collection}")

            logging.info(f"Snapshot completed successfully at {timestamp}")
            
        except Exception as e:
            logging.error(f"Error creating snapshot: {str(e)}")
            raise
        finally:
            if self.client:
                self.client.close()

    def list_snapshots(self):
        """List all available snapshots"""
        try:
            if not os.path.exists(self.backup_dir):
                return []
            
            snapshots = []
            for snapshot_dir in os.listdir(self.backup_dir):
                if os.path.isdir(os.path.join(self.backup_dir, snapshot_dir)):
                    snapshots.append(snapshot_dir)
            
            return sorted(snapshots, reverse=True)
        except Exception as e:
            logging.error(f"Error listing snapshots: {str(e)}")
            raise

    def restore_snapshot(self, snapshot_name):
        """Restore a snapshot to MongoDB"""
        try:
            if not self.client:
                self.connect()

            snapshot_path = os.path.join(self.backup_dir, snapshot_name)
            if not os.path.exists(snapshot_path):
                raise ValueError(f"Snapshot {snapshot_name} does not exist")

            # Get list of databases in the snapshot
            databases = [d for d in os.listdir(snapshot_path) 
                        if os.path.isdir(os.path.join(snapshot_path, d))]

            for db_name in databases:
                db_path = os.path.join(snapshot_path, db_name)
                logging.info(f"Restoring database: {db_name}")
                
                # Create database backup directory
                db = self.client[db_name]

                # Get list of collection files
                collection_files = [f for f in os.listdir(db_path) 
                                  if f.endswith('.json')]

                for collection_file in collection_files:
                    collection_name = collection_file[:-5]  # Remove .json extension
                    collection_path = os.path.join(db_path, collection_file)
                    
                    logging.info(f"Restoring collection: {collection_name}")

                    # Clear existing collection
                    db[collection_name].delete_many({})

                    # Read and restore documents
                    with open(collection_path, 'r') as f:
                        documents = []
                        for line in f:
                            try:
                                # Convert string representation to dict
                                doc_str = line.strip()
                                if doc_str.startswith("{"):
                                    doc = eval(doc_str)  # Safe here as we created the file
                                    documents.append(doc)
                            except Exception as e:
                                logging.error(f"Error parsing document: {str(e)}")
                                continue

                        if documents:
                            db[collection_name].insert_many(documents)
                            logging.info(f"Restored {len(documents)} documents to {db_name}.{collection_name}")

            logging.info(f"Snapshot {snapshot_name} restored successfully")

        except Exception as e:
            logging.error(f"Error restoring snapshot: {str(e)}")
            raise
        finally:
            if self.client:
                self.client.close()

def main():
    snapshot_manager = MongoDBSnapshot()
    
    # Schedule snapshot creation (every day at 2 AM)
    schedule.every().day.at("02:00").do(snapshot_manager.create_snapshot)
    
    logging.info("MongoDB Snapshot Manager started")
    
    while True:
        try:
            schedule.run_pending()
            time.sleep(60)
        except Exception as e:
            logging.error(f"Error in main loop: {str(e)}")
            time.sleep(300)  # Wait 5 minutes before retrying

if __name__ == "__main__":
    main() 