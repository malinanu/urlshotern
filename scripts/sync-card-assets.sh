#!/usr/bin/env bash
# Digital visiting cards share one set of stylesheets, one template catalog and one set of fonts
# between the React renderer (MicroPages) and the Go renderer (Microsites). The frontend copy is
# the source; this script copies it to the Go side. A Go test fails if the copies drift.
set -euo pipefail
cd "$(dirname "$0")/.."

SRC=frontend/src/components/micropages/cards
DST=internal/services/microsite_web

mkdir -p "$DST/cards" "$DST/fonts"
cp "$SRC/catalog.json" "$DST/cards/catalog.json"
for f in base zenith nova atlas axiom vectra pulse horizon verdant skyway creo bravo imagine; do
  cp "$SRC/css/$f.css" "$DST/cards/$f.css"
done
cp frontend/public/card-fonts/*.woff2 "$DST/fonts/"
cp frontend/public/card-fonts/licenses/LICENSE-*.txt "$DST/fonts/"
echo "card assets synced to $DST"
