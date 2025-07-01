#!/bin/bash

# check if localstack is available on port 4566
if nc -z localhost 4566; then
    echo "Localstack is available on port 4566"
else
    echo "Localstack is not available on port 4566"
    exit 1
fi

echo "Running tests..."
AWS_ACCESS_KEY_ID="test" AWS_SECRET_ACCESS_KEY="test" AWS_DEFAULT_REGION="us-east-1" AWS_ENDPOINT_URL="http://localhost:4566" go test -v ./...
