#!/usr/bin/env bash
set -euo pipefail

# Build web resources
echo "Building webresources..."
pushd webresources >/dev/null
npm run build
popd >/dev/null

# Ensure static js output directory exists
mkdir -p cmd/static/js

# Copy vendored runtimes into output directory (ignore missing files)
if [ -f webresources/node_modules/@azure/msal-browser/lib/msal-browser.min.js ]; then
  cp webresources/node_modules/@azure/msal-browser/lib/msal-browser.min.js cmd/static/js/
else
  echo "Warning: msal-browser.min.js not found; run npm install in webresources if needed" >&2
fi

if [ -f webresources/node_modules/htmx.org/dist/htmx.min.js ]; then
  cp webresources/node_modules/htmx.org/dist/htmx.min.js cmd/static/js/
else
  echo "Warning: htmx.min.js not found; run npm install in webresources if needed" >&2
fi

echo "Starting Go app..."
go run cmd/main.go --local
