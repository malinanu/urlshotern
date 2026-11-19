#!/bin/sh
# Builds the plugin downloads served at /downloads/ on the site:
#   trunc-chrome-extension.zip  (manifest.json at the root: "Load unpacked" after unzipping)
#   trunc-wordpress-plugin.zip  (a trunc-url-shortener/ folder: Plugins -> Upload Plugin)
# Usage: scripts/build-plugins.sh [output-dir]   (default: frontend/public/downloads)
# Runs in Dockerfile.dev before the frontend build; the zips are not committed.
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OUT=${1:-"$ROOT/frontend/public/downloads"}
mkdir -p "$OUT"
OUT=$(cd "$OUT" && pwd)
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

rm -f "$OUT/trunc-chrome-extension.zip" "$OUT/trunc-wordpress-plugin.zip"

# Chrome extension: everything except the developer README.
(cd "$ROOT/trunc-chrome-extension" && zip -qr -X "$OUT/trunc-chrome-extension.zip" . -x 'README.md' '*.DS_Store')

# WordPress plugin: wrapped in its folder, as WordPress expects. readme.txt stays (WordPress.org listing).
mkdir -p "$TMP/trunc-url-shortener"
cp -R "$ROOT/trunc-url-shortener/." "$TMP/trunc-url-shortener/"
rm -rf "$TMP/trunc-url-shortener/README.md" "$TMP/trunc-url-shortener/node_modules"
(cd "$TMP" && zip -qr -X "$OUT/trunc-wordpress-plugin.zip" trunc-url-shortener -x '*.DS_Store')

echo "Built $OUT/trunc-chrome-extension.zip and $OUT/trunc-wordpress-plugin.zip"
