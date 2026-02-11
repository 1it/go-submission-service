#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables - use env vars if set, otherwise use defaults
DB_PATH="${DB_PATH:-/app/data/subscribers.db}"
BACKUP_DIR="${BACKUP_DIR:-/app/backups}"
BACKUP_FILE="subscribers_$(date +%Y%m%d_%H%M%S).db"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Check if the database file exists
if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file not found at $DB_PATH"
    exit 1
fi

# Create backup
echo "Creating backup of $DB_PATH to $BACKUP_DIR/$BACKUP_FILE"
cp "$DB_PATH" "$BACKUP_DIR/$BACKUP_FILE"

# Verify backup
if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
    echo "Backup created successfully: $BACKUP_DIR/$BACKUP_FILE"
    ls -lh "$BACKUP_DIR/$BACKUP_FILE"
else
    echo "Error: Backup failed"
    exit 1
fi

# Cleanup old backups (keep last 10)
echo "Cleaning up old backups (keeping last 10)..."
ls -t "$BACKUP_DIR"/subscribers_*.db | tail -n +11 | xargs -r rm
echo "Done." 