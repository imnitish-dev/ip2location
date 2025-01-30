#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Error: .env file not found"
    exit 1
fi

# Verify required environment variables are set
required_vars=(
    "MAXMIND_ACCOUNT"
    "MAXMIND_LICENSE_KEY"
    "IP2LOCATION_TOKEN"
    "IP2LOCATION_CODE"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: Required environment variable $var is not set"
        exit 1
    fi
done

# Directory setup
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DOWNLOAD_DIR="$SCRIPT_DIR/downloads"
DB_DIR="$SCRIPT_DIR"

# Create download directory if it doesn't exist
mkdir -p "$DOWNLOAD_DIR"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Update MaxMind Database
update_maxmind() {
    log "Updating MaxMind database..."
    
    # Create a temporary extraction directory
    mkdir -p "$DOWNLOAD_DIR/maxmind_temp"
    
    # Download the latest database using basic auth
    log "Downloading MaxMind database..."
    
    local output_file="$DOWNLOAD_DIR/GeoLite2-City.tar.gz"
    
    curl -u "$MAXMIND_ACCOUNT:$MAXMIND_LICENSE_KEY" \
        --fail \
        --silent \
        --show-error \
        --location \
        --output "$output_file" \
        'https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz'
    
    local curl_status=$?
    
    if [ $curl_status -ne 0 ]; then
        log "Failed to download MaxMind database. Curl exit code: $curl_status"
        return 1
    fi
    
    if [ ! -s "$output_file" ]; then
        log "Downloaded MaxMind file is empty"
        return 1
    fi
    
    log "Extracting MaxMind database..."
    tar -xzf "$output_file" -C "$DOWNLOAD_DIR/maxmind_temp"
    
    if [ $? -ne 0 ]; then
        log "Failed to extract MaxMind database"
        return 1
    fi
    
    # Move the new database file
    local mmdb_file=$(find "$DOWNLOAD_DIR/maxmind_temp" -name "*.mmdb" -type f)
    
    if [ -z "$mmdb_file" ]; then
        log "MaxMind database file not found after extraction"
        return 1
    fi
    
    mv "$mmdb_file" "$DB_DIR/MaxMind.mmdb"
    log "MaxMind database updated successfully"
    return 0
}

# Update IP2Location Database
update_ip2location() {
    log "Updating IP2Location database..."
    
    local output_file="$DOWNLOAD_DIR/IP2LOCATION-LITE-DB11.IPV6.BIN.ZIP"
    
    # Download the latest database with modified URL and proper handling
    curl --fail --silent --show-error --location \
        --output "$output_file" \
        --retry 3 \
        --retry-delay 2 \
        "https://www.ip2location.com/download/?token=$IP2LOCATION_TOKEN&file=$IP2LOCATION_CODE"
    
    local curl_status=$?
    
    if [ $curl_status -ne 0 ]; then
        log "Failed to download IP2Location database. Curl exit code: $curl_status"
        return 1
    fi
    
    if [ ! -s "$output_file" ]; then
        log "Downloaded IP2Location file is empty"
        return 1
    fi
    
    # Verify file is a valid ZIP
    if ! unzip -t "$output_file" >/dev/null 2>&1; then
        log "Downloaded file is not a valid ZIP archive"
        return 1
    fi
    
    log "Extracting IP2Location database..."
    unzip -o "$output_file" -d "$DOWNLOAD_DIR"
    
    if [ $? -ne 0 ]; then
        log "Failed to extract IP2Location database"
        return 1
    fi
    
    # Move the new database file (updated for IPv6)
    local bin_file="$DOWNLOAD_DIR/IP2LOCATION-LITE-DB11.IPV6.BIN"
    if [ -f "$bin_file" ]; then
        mv "$bin_file" "$DB_DIR/IP2LOCATION.BIN"
        log "IP2Location database updated successfully"
        return 0
    else
        log "IP2Location BIN file not found after extraction"
        return 1
    fi
}

# Clean up old files
cleanup() {
    log "Cleaning up temporary files..."
    rm -rf "$DOWNLOAD_DIR/maxmind_temp"
    rm -f "$DOWNLOAD_DIR"/*.tar.gz
    rm -f "$DOWNLOAD_DIR"/*.zip
    rm -f "$DOWNLOAD_DIR"/*
}

# Main execution
log "Starting database update process..."

update_maxmind
MAXMIND_STATUS=$?

update_ip2location
IP2LOCATION_STATUS=$?

cleanup

# Check if both updates were successful
if [ $MAXMIND_STATUS -eq 0 ] && [ $IP2LOCATION_STATUS -eq 0 ]; then
    log "All databases updated successfully"
    exit 0
else
    log "One or more database updates failed"
    exit 1
fi