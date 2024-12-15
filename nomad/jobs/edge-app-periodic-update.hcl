job "edge-app-periodic" {

  type = "sysbatch"
  periodic {
    // https://github.com/hashicorp/cronexpr#implementation
    crons = [
      "*/25 * * * * * *",
    ]
    prohibit_overlap = false
    time_zone        = "Asia/Jakarta"
  }

  region = "global"
  datacenters = ["edge"]

  constraint {
    attribute = "${meta.platform}"
    value     = "aws"
  }

  group "edge-app" {

    # disconnect {
    #   lost_after = "1h"
    #   replace    = false
    #   reconcile  = "best_score"
    # }

    task "edge-app" {

      driver = "raw_exec"

      config {
        command = "/bin/bash"
        args = [
          "-c",
          "chmod +x $NOMAD_TASK_DIR/update.sh && $NOMAD_TASK_DIR/update.sh"
        ]
      }

      template {
        destination = "local/update.sh"
        change_mode = "noop"
        data = <<EOF
#!/bin/bash

# Configuration
SERVICE_NAME="edge-app"            # Replace with your systemd service name
CURRENT_BINARY="/opt/edge-app/edge-app"  # Path to the current binary
NEW_BINARY_URL="https://github.com/rahadiangg/demo-nomad-edge/releases/download/v0.0.2/edge-app_v0.0.2_linux_amd64" # Base URL to download new binary
NEW_VERSION="0.0.2"             # Define the desired version here

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
echo "Target version: v$NEW_VERSION"

# Compare versions using sort -V
if printf "%s\n%s\n" "$CURRENT_VERSION" "$NEW_VERSION" | sort -V | tail -n1 | grep -q "^$NEW_VERSION$"; then
    if [[ "$CURRENT_VERSION" != "$NEW_VERSION" ]]; then
        echo "Newer version detected. Upgrading..."
        
        # Download the new binary
        NEW_BINARY_TEMP="/opt/edge-app/edge_app_$NEW_VERSION"
        echo "Downloading new version from $NEW_BINARY_URL"
        curl -L -o "$NEW_BINARY_TEMP" "$NEW_BINARY_URL" || {
            echo "Error: Failed to download new binary"
            exit 1
        }

        # Validate the new binary
        if [ ! -x "$NEW_BINARY_TEMP" ]; then
            chmod +x "$NEW_BINARY_TEMP"
        fi

        NEW_BINARY_VERSION=$("$NEW_BINARY_TEMP" --version | sed 's/^v//')
        if [[ "$NEW_BINARY_VERSION" != "$NEW_VERSION" ]]; then
            echo "Error: Version mismatch in downloaded binary ($NEW_BINARY_VERSION != $NEW_VERSION)"
            rm -f "$NEW_BINARY_TEMP"
            exit 1
        fi

        # Stop the service
        echo "Stopping the service: $SERVICE_NAME"
        sudo systemctl stop "$SERVICE_NAME"

        # Replace the binary
        echo "Replacing the binary..."
        sudo mv "$NEW_BINARY_TEMP" "$CURRENT_BINARY"
        sudo chmod +x "$CURRENT_BINARY"

        # Start the service
        echo "Starting the service: $SERVICE_NAME"
        sudo systemctl start "$SERVICE_NAME"

        echo "Upgrade complete. Service restarted."
    else
        echo "Current version is up-to-date. No upgrade needed."
    fi
else
    echo "Current version is newer than or equal to target version. No upgrade needed."
fi
EOF
      }
    }
  }
}