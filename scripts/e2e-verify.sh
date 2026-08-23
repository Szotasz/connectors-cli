#!/bin/bash
# Post-fix end-to-end verify for the object-flag parse (PR #4): a real
# `falai generate` through the gateway with an object --params. Proves the
# params travel as a JSON object (the v0.3.0 bug sent them as a raw string
# and every known model answered "Missing required parameter").
#
# PAID STEP: submits ONE fal-ai/flux/schnell image job. The fleet convention
# requires the manifest next to this script -- without it, stop loudly.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
if [ ! -f "$HERE/paid-steps.yaml" ]; then
  echo "FIZETOS LEPES MANIFESZT NELKUL -- a hivasai szamolatlanok. Hianyzik: $HERE/paid-steps.yaml" >&2
  exit 3
fi
: "${CONNECTORS_HU_TOKEN:?CONNECTORS_HU_TOKEN kell (cnk_ API kulcs)}"
BIN="${CONNECTORS_BIN:-$HERE/../connectors}"
[ -x "$BIN" ] || { echo "nincs build-elt binaris: $BIN (go build -o connectors .)" >&2; exit 2; }
"$BIN" falai generate --model_id "fal-ai/flux/schnell" --params '{"prompt":"a tiny red fox sketch, minimal line art"}'
