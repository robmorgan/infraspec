# Tests

This folder contains test helpers and utilities for InfraSpec.

## Test Structure

### Unit Tests
Unit tests are located alongside the code they test (e.g., `pkg/awshelpers/region_test.go`). These tests verify individual functions and packages in isolation.

Run unit tests with:
```sh
make test
```

Or directly:
```sh
go test -v $(go list ./... | grep -v '/test$')
```

### Integration Tests
Integration tests are run in CI by executing the `infraspec` CLI against all feature files in the `features/` directory.

On GitHub Actions, the workflow:
1. Builds the `infraspec` binary
2. Runs each feature file using `infraspec <feature-file> --virtual-cloud`
3. Uses the InfraSpec Cloud API at `https://api.infraspec.sh`

## Running Integration Tests Locally

### Using InfraSpec Cloud API
Build the binary and run features with the `--virtual-cloud` flag:

```sh
go build -o ./infraspec ./cmd/infraspec
INFRASPEC_CLOUD_TOKEN=<your-token> ./infraspec features/aws/s3/s3_bucket.feature --virtual-cloud
```

### Using Local InfraSpec API
First, start the local InfraSpec API and httpbin:
```sh
docker-compose up -d
```

Then run features without the virtual cloud flag:
```sh
go build -o ./infraspec ./cmd/infraspec
source ./test/set-env-vars.sh
./infraspec features/aws/s3/s3_bucket.feature
```

## Test Helpers

This directory contains:
- `testhelpers/` - Common test utilities and setup functions
- `httpserver/` - Mock HTTP server for testing HTTP assertions
- `integration/` - Unit tests for HTTP assertion functions
- `docker-compose.yml` - Local test environment setup
- `set-env-vars.sh` - Environment variable setup for local testing
