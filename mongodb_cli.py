import argparse
from mongodb_snapshot import MongoDBSnapshot

def main():
    parser = argparse.ArgumentParser(description='MongoDB Snapshot Manager CLI')
    parser.add_argument('action', choices=['create', 'restore', 'list'],
                      help='Action to perform: create, restore, or list snapshots')
    parser.add_argument('--snapshot', help='Snapshot name for restore action')
    
    args = parser.parse_args()
    snapshot_manager = MongoDBSnapshot()
    
    if args.action == 'create':
        snapshot_manager.create_snapshot()
    elif args.action == 'restore':
        if not args.snapshot:
            print("Error: Snapshot name is required for restore action")
            return
        snapshot_manager.restore_snapshot(args.snapshot)
    elif args.action == 'list':
        snapshots = snapshot_manager.list_snapshots()
        if snapshots:
            print("\nAvailable snapshots:")
            for snapshot in snapshots:
                print(f"- {snapshot}")
        else:
            print("\nNo snapshots found")

if __name__ == "__main__":
    main() 