#!/bin/bash

# Set the working directory (update this with the path to your Go project)
PROJECT_DIR="/home/ubuntu/go-app/ip2location"
BINARY_NAME="ip2location"
SERVICE_NAME="ip2location.service"

# Step 1: Go to the project directory
echo "Navigating to project directory..."
cd $PROJECT_DIR || { echo "Project directory not found!"; exit 1; }

# Step 2: Pull the latest codebase (if you're using git)
echo "Pulling the latest codebase..."
git pull origin main || { echo "Git pull failed!"; exit 1; }

# Step 3: Stop the service before building
echo "Stopping the $SERVICE_NAME service..."
sudo systemctl stop $SERVICE_NAME || { echo "Failed to stop the service!"; exit 1; }

# Step 4: Build the Go application
echo "Building the Go application..."
GOOS=linux GOARCH=arm64 go build -o /usr/local/bin/$BINARY_NAME || { echo "Build failed!"; exit 1; }

# Step 5: Restart the systemd service
echo "Restarting the $SERVICE_NAME service..."
sudo systemctl start $SERVICE_NAME || { echo "Failed to restart the service!"; exit 1; }

# Step 6: Verify the service status
echo "Verifying the service status..."
sudo systemctl status $SERVICE_NAME --no-pager

# Done
echo "Deployment completed successfully!"
