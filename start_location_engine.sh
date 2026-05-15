#!/bin/bash

# 1. Kill and remove the container if it already exists
echo "Cleaning up old Meilisearch containers..."
docker rm -f meili_rideshare 2>/dev/null || true

# 2. Start Meilisearch
echo "Starting Meilisearch..."
docker run -d --name meili_rideshare \
    -p 7700:7700 \
    -v $(pwd)/meili_data:/meili_data \
    getmeili/meilisearch:v1.7 \
    meilisearch --master-key="masterKey"

# 3. Wait for Meilisearch to be ready (Authenticated Health Check)
echo "Waiting for Meilisearch to be ready..."
until curl -s -H "Authorization: Bearer masterKey" http://localhost:7700/health | grep -q "status\":\"available\""; do
    printf '.'
    sleep 1
done

# 4. Upload the JSON
echo -e "\nUploading data..."
# IMPORTANT: Ensure the path to your JSON is correct
JSON_PATH="/home/raotalha/Code/PersonalCode/rideshare/backend/location/data/unified_pakistan_locations.json"

curl \
  -X POST 'http://localhost:7700/indexes/locations/documents' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer masterKey' \
  --data-binary @"$JSON_PATH"

echo -e "\n--- Success! Data is indexing in the background. ---"
echo "You can check progress at: http://localhost:7700"