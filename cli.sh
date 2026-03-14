#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# cli.sh — Bitcoin chain analysis CLI
#
# Usage:
#   ./cli.sh --block <blk.dat> <rev.dat> <xor.dat>
#
# Block mode:
#   - Reads blk*.dat, rev*.dat, and xor.dat
#   - Parses all blocks and transactions
#   - Applies chain analysis heuristics to every transaction
#   - Writes per-block-file outputs:
#       out/<blk_stem>.json — machine-readable analysis report
#       out/<blk_stem>.md   — human-readable Markdown report
#     where <blk_stem> is the blk filename without extension (e.g., blk04330)
#   - Exits 0 on success, 1 on error
###############################################################################

error_json() {
  local code="$1"
  local message="$2"
  printf '{"ok":false,"error":{"code":"%s","message":"%s"}}\n' "$code" "$message"
}

# --- Block mode ---
if [[ "${1:-}" != "--block" ]]; then
  error_json "INVALID_ARGS" "Usage: cli.sh --block <blk.dat> <rev.dat> <xor.dat>"
  echo "Error: This CLI only supports block mode. Use --block flag." >&2
  exit 1
fi

shift
if [[ $# -lt 3 ]]; then
  error_json "INVALID_ARGS" "Block mode requires: --block <blk.dat> <rev.dat> <xor.dat>"
  echo "Error: Block mode requires 3 file arguments: <blk.dat> <rev.dat> <xor.dat>" >&2
  exit 1
fi

BLK_FILE="$1"
REV_FILE="$2"
XOR_FILE="$3"

for f in "$BLK_FILE" "$REV_FILE" "$XOR_FILE"; do
  if [[ ! -f "$f" ]]; then
    error_json "FILE_NOT_FOUND" "File not found: $f"
    echo "Error: File not found: $f" >&2
    exit 1
  fi
done

# Create output directory
mkdir -p out

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/bin/cli"

# Build the binary if missing (e.g. fresh CI checkout)
if [[ ! -f "$BINARY" ]]; then
  cd "$SCRIPT_DIR" || exit 1
  go build -o "$BINARY" ./cmd/cli/ 2>&1 >&2 || {
    error_json "BUILD_ERROR" "Failed to build cli binary"
    exit 1
  }
fi

exec "$BINARY" --block "$BLK_FILE" "$REV_FILE" "$XOR_FILE"
