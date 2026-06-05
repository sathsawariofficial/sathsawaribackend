#!/bin/bash

set -euo pipefail

# Directory containing logs
TARGET_DIR="/var/log/myapp"

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

# Find:
#   *.log                 -> current logs
#   YYYY-MM-DDTHH-MM-SS.xxx_Name -> old rotated logs
mapfile -d '' FILES < <(
    find "$TARGET_DIR" -maxdepth 1 -type f \
    \( \
        -name "*.log" \
        -o -regex '.*/[0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}T[0-9-]\+\.[0-9]\+_.*' \
    \) \
    -print0
)

if [[ ${#FILES[@]} -eq 0 ]]; then
    echo "No log files found."
    exit 0
fi

echo "Creating archive: $ARCHIVE"

zip -j "$ARCHIVE" "${FILES[@]}"

echo "Archive created successfully."

echo "Deleting original files..."
rm -f "${FILES[@]}"

echo "Backup complete: $ARCHIVE"