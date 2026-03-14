#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# web.sh — Web visualizer
#
# Starts the web visualizer server.
#
# Behavior:
#   - Reads PORT env var (default: 3000)
#   - Prints the URL (e.g., http://127.0.0.1:3000) to stdout
#   - Keeps running until terminated (CTRL+C / SIGTERM)
#   - Must serve GET /api/health -> 200 { "ok": true }
#
# TODO: Replace the stub below with your web server start command.
###############################################################################

PORT="${PORT:-3000}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/bin/web"

# Build the web binary if missing
if [[ ! -f "$BINARY" ]]; then
  cd "$SCRIPT_DIR" || exit 1
  go build -o "$BINARY" ./cmd/web/ 2>&1 >&2 || {
    echo '{"ok":false,"error":{"code":"BUILD_ERROR","message":"Failed to build web binary"}}' >&2
    exit 1
  }
fi

export PORT
exec "$BINARY"
