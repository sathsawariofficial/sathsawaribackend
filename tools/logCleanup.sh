#!/bin/bash

set -euo pipefail

# Directory containing logs
TARGET_DIR="/home/raotalha/Code/PersonalCode/sathsawaribackend"

# Backup directory
BACKUP_DIR="$TARGET_DIR/backups"

mkdir -p "$BACKUP_DIR"

# Safety check
if [[ ! -d "$TARGET_DIR" ]]; then
    echo "Error: directory does not exist -> $TARGET_DIR"
    exit 1
fi

TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
ARCHIVE="$BACKUP_DIR/logs_$TIMESTAMP.zip"

# Only match rotated timestamped log files like:
# 2026-06-05T09-11-08.237_SathSawari.log
mapfile -d '' FILES < <(
    find "$TARGET_DIR" -maxdepth 1 -type f \
    -name "20??-??-??T*_*\.log" \
    -print0
)

if [[ ${#FILES[@]} -eq 0 ]]; then
    echo "No rotated log files found."
    exit 0
fi

echo "Files to archive:"
printf '%s\n' "${FILES[@]}"

echo "Creating archive: $ARCHIVE"

zip -j "$ARCHIVE" "${FILES[@]}"

echo "Archive created successfully."

echo "Deleting archived files..."
rm -f "${FILES[@]}"

echo "Backup complete: $ARCHIVE"