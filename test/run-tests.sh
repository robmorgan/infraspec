#!/bin/bash

# check if infraspec-api is available on port 8000
if nc -z localhost 8000; then
    echo "InfraSpec API is available on port 8000"
else
    echo "InfraSpec API is not available on port 8000"
    exit 1
fi

echo "Running tests..."
AWS_ACCESS_KEY_ID="test" AWS_SECRET_ACCESS_KEY="test" AWS_DEFAULT_REGION="us-east-1" AWS_ENDPOINT_URL="http://localhost:8000" go test -v ./...
