#!/bin/bash

# Configuration
APP_DIR="/home/ubuntu/go-app/ip2location"
APP_NAME="ip2location"
LOG_FILE="/var/log/ip2location/update.log"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Create log directory
mkdir -p "$(dirname $LOG_FILE)"

log "Starting update process..."

# Step 1: Update databases
cd $APP_DIR
./update-db.sh
DB_STATUS=$?

if [ $DB_STATUS -eq 0 ]; then
    log "Database update successful"
else
    log "Database update failed"
fi

# Step 2: Restart application using systemctl
log "Restarting application..."
sudo systemctl restart ip2location

# Step 3: Verify service is running
sleep 2
if systemctl is-active --quiet ip2location; then
    log "Application restarted successfully"
else
    log "Failed to restart application"
    systemctl status ip2location >> "$LOG_FILE"
fi

log "Update process completed" 