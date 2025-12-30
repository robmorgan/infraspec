# Tests

This folder contains test helpers and utilities for InfraSpec.

## Test Structure

### Unit Tests

Unit tests are located alongside the code they test (e.g., `pkg/awshelpers/region_test.go`). These tests verify
individual functions and packages in isolation.

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
2. Runs each feature file against the builtin AWS emulator using `infraspec <feature-file>`

## Running Integration Tests Locally

### Using the Builtin InfraSpec AWS Emulator

Build the binary and run features:

```sh
go build -o ./infraspec ./cmd/infraspec
./infraspec features/aws/s3/s3_bucket.feature
```

### Using Real AWS APIs

Build the binary and run features using the `--live` flag:

```sh
go build -o ./infraspec ./cmd/infraspec
./infraspec --live features/aws/s3/s3_bucket.feature
```

**Note:** This creates real running infrastructure. Be sure to cleanup any dangling resources.

### HTTP Tests

The HTTP tests require the `httpbin` emulator to be running locally:

```sh
docker-compose up -d
```

## Test Helpers

This directory contains:

- `testhelpers/` - Common test utilities and setup functions
- `httpserver/` - Mock HTTP server for testing HTTP assertions
- `integration/` - Unit tests for HTTP assertion functions
- `docker-compose.yml` - Local test environment setup
