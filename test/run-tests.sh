#!/bin/bash

# check if infraspec-api is available on port 3687
if nc -z localhost 3687; then
    echo "InfraSpec API is available on port 3687"
else
    echo "InfraSpec API is not available on port 3687"
    exit 1
fi

echo "Running tests..."
AWS_ACCESS_KEY_ID="test" AWS_SECRET_ACCESS_KEY="test" AWS_DEFAULT_REGION="us-east-1" AWS_ENDPOINT_URL="http://localhost:3687" go test -v ./...
