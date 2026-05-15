#!/bin/bash

# Define the paths
GO_APP_DIR="/opt/rideshare"
GO_BINARY="$GO_APP_DIR/rideshare"
CONFIG_FILE="$GO_APP_DIR/conf/configuration.json"
APP_NAME="Sawarilink-Backend"

# Navigate to the Go application directory
cd $GO_APP_DIR

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Go binary built successfully."

    # Stop any previously running instance
    if pm2 list | grep -q "$APP_NAME"; then
        echo "Stopping previously running instance..."
        pm2 stop "$APP_NAME"
        pm2 delete "$APP_NAME"
    fi

    # Start the Go application using PM2
    echo "Starting Go application with PM2..."
    pm2 start $GO_BINARY --name "$APP_NAME" -- -config "$CONFIG_FILE"

    echo "PM2 status:"
    pm2 status
else
    echo "Failed to build Go application. Exiting..."
    exit 1
fi
