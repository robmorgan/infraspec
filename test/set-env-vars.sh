#!/bin/bash

# NOTE: This script is used to set the environment variables for the InfraSpec API container.
# It is used in the integration tests or local testing.
#
# In most cases, you shouldn't execute this script directly, use `make test` instead.
# If you do use this script be sure to source it with: `source ./test/set-env-vars.sh` command to source the
# environment variables.

echo "Setting environment variables for InfraSpec API..."
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test"
export AWS_DEFAULT_REGION="us-east-1"
export AWS_ENDPOINT_URL="http://localhost:3687"
