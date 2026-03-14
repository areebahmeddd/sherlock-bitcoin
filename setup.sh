#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# setup.sh — Install project dependencies
#
# Add your install commands below (e.g., npm install, pip install, cargo build).
# This script is run once before grading to set up the environment.
###############################################################################

# Decompress block fixtures if not already present
for gz in fixtures/*.dat.gz; do
  dat="${gz%.gz}"
  if [[ ! -f "$dat" ]]; then
    echo "Decompressing $(basename "$gz")..."
    gunzip -k "$gz"
  fi
done

# Build Go binaries
echo "Building cli binary..."
go build -o bin/cli ./cmd/cli/ || {
  echo "ERROR: Failed to build cli binary" >&2
  exit 1
}

echo "Building web binary..."
go build -o bin/web ./cmd/web/ || {
  echo "ERROR: Failed to build web binary" >&2
  exit 1
}

echo "Setup complete"
