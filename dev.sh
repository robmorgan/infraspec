#!/bin/bash
# dev.sh - Run infraspec against a local infraspec-api instance
#
# Usage:
#   ./dev.sh features/s3.feature          # Run specific feature
#   ./dev.sh features/                    # Run all features
#   ./dev.sh --help                       # Show infraspec help
#
# Prerequisites:
#   Start infraspec-api locally first:
#   cd ../infraspec-api && go run ./cmd/emulator --port 3687 --auth-enabled --api-keys "infraspec-api:securetoken"
#
# Environment variables:
#   INFRASPEC_LOCAL_API_URL  - Local API URL (default: http://localhost:3687)

set -euo pipefail

LOCAL_API_URL="${INFRASPEC_LOCAL_API_URL:-http://localhost:3687}"

# Use --virtual-cloud mode with local endpoint override
# This exercises the same code paths as production
export AWS_ENDPOINT_URL="$LOCAL_API_URL"

# Token for local infraspec-api (matches: --api-keys "infraspec-api:securetoken")
export INFRASPEC_CLOUD_TOKEN="securetoken"

# Set AWS region
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"

# Optional: Enable debug logging
# export INFRASPEC_DEBUG=1

# Check if infraspec-api is running
if ! curl -s --connect-timeout 2 "$LOCAL_API_URL/_health" > /dev/null 2>&1; then
    echo "⚠️  infraspec-api not responding at $LOCAL_API_URL"
    echo ""
    echo "Start it with:"
    echo "  cd ../infraspec-api && go run ./cmd/emulator --port 3687 --auth-enabled --api-keys \"infraspec-api:securetoken\""
    echo ""
    echo "Or set INFRASPEC_LOCAL_API_URL to a different address."
    exit 1
fi

echo "✓ Using local infraspec-api at $LOCAL_API_URL"
echo ""

# Run infraspec with --virtual-cloud flag and pass through all arguments
exec go run ./cmd/infraspec --virtual-cloud "$@"
