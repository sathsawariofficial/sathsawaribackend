#!/bin/bash

set -euo pipefail

###############################################
# PostgreSQL Configuration
###############################################

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="rideshare"
DB_USER="postgres"

# Export password or use ~/.pgpass
export PGPASSWORD="postgres"

###############################################
# Backup Configuration
###############################################

BACKUP_DIR="/home/raotalha/Code/PersonalCode/sathsawaribackend/db_backups"

mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.dump"

echo "Creating PostgreSQL backup..."

pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -Fc \
    --verbose \
    -f "$BACKUP_FILE"

echo
echo "Backup completed successfully."
echo "File:"
echo "$BACKUP_FILE"