# Tests

This folder contains automated tests for InfraSpec. All of the tests are written in [Go](https://golang.org/).
Most of these are "integration tests" that deploy real infrastructure using Terraform and verify that infrastructure 
works as expected. This allows InfraSpec to effectively test itself.

We use [LocalStack](https://www.localstack.cloud) so we can run the tests both cheaply and quickly without creating real
AWS infrastructure. However, tests need to be executed in a real AWS account when they are merged into the main branch.
This is currently a manual process.

You can run the tests locally by first starting a LocalStack instance and then using Go:

```sh
docker-compose up
go test -v ./...
```

Or using [Act](https://github.com/nektos/act):

```sh
act -W .github/workflows/test.yml
```
