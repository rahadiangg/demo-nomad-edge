#!/bin/bash

# Configuration
SERVICE_NAME="edge-app"            # Replace with your systemd service name
CURRENT_BINARY="/opt/edge-app/edge-app"  # Path to the current binary
BINARY_URL="https://github.com/rahadiangg/demo-nomad-edge/releases/download/v0.0.1/edge-app_v0.0.1_linux_amd64" # Base URL to download binary versions
ROLLBACK_VERSION="0.0.1"       # Define the rollback version here

# Get the current version
if [ ! -x "$CURRENT_BINARY" ]; then
    echo "Error: Current binary not found or not executable at $CURRENT_BINARY"
    exit 1
fi

CURRENT_VERSION=$("$CURRENT_BINARY" --version | sed 's/^v//') # Remove 'v' prefix for comparison
if [ -z "$CURRENT_VERSION" ]; then
    echo "Error: Unable to get current version from $CURRENT_BINARY"
    exit 1
fi

echo "Current version: v$CURRENT_VERSION"
echo "Rollback to version: v$ROLLBACK_VERSION"

# Check if rollback is necessary
if [[ "$CURRENT_VERSION" == "$ROLLBACK_VERSION" ]]; then
    echo "Current version is already v$ROLLBACK_VERSION. No rollback needed."
    exit 0
fi

# Download the rollback version
ROLLBACK_BINARY_TEMP="/opt/edge-app/edge_app_$ROLLBACK_VERSION"
echo "Downloading rollback version from $BINARY_URL"
curl -L -o "$ROLLBACK_BINARY_TEMP" "$BINARY_URL" || {
    echo "Error: Failed to download rollback binary"
    exit 1
}

# Validate the rollback binary
if [ ! -x "$ROLLBACK_BINARY_TEMP" ]; then
    chmod +x "$ROLLBACK_BINARY_TEMP"
fi

ROLLBACK_BINARY_VERSION=$("$ROLLBACK_BINARY_TEMP" --version | sed 's/^v//')
if [[ "$ROLLBACK_BINARY_VERSION" != "$ROLLBACK_VERSION" ]]; then
    echo "Error: Version mismatch in downloaded binary ($ROLLBACK_BINARY_VERSION != $ROLLBACK_VERSION)"
    rm -f "$ROLLBACK_BINARY_TEMP"
    exit 1
fi

# Stop the service
echo "Stopping the service: $SERVICE_NAME"
sudo systemctl stop "$SERVICE_NAME"

# Replace the binary
echo "Replacing the binary with rollback version..."
sudo mv "$ROLLBACK_BINARY_TEMP" "$CURRENT_BINARY"
sudo chmod +x "$CURRENT_BINARY"

# Start the service
echo "Starting the service: $SERVICE_NAME"
sudo systemctl start "$SERVICE_NAME"

echo "Rollback complete. Service restarted."
