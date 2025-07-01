COLOR 	:= "\e[1;36m%s\e[0m\n"
RED 	:= "\e[1;31m%s\e[0m\n"

.PHONY: deps
deps: ## install dependencies
	go mod download
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.2.1

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: lint
lint: ## lint
	@printf $(COLOR) "Linting..."
	@[ ! -e .golangci.yml ] || golangci-lint run
	@[ ! -e "$(REPO_ROOT)/.golangci.yml" ] || { printf $(COLOR) "Using root .golangci.yml" ; golangci-lint run -c "$(REPO_ROOT)/.golangci.yml"; }

.PHONY: fmt
fmt: tidy ## tidy, format and imports
	[ ! -e buf.gen.yaml ] || buf format -w
	gofumpt -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	goimports -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	gci write --skip-generated -s standard -s default -s "prefix(github.com/robmorgan/infraspec)" .

.PHONY: go-test-cover
go-test-cover: ## run test & generate coverage
	@printf $(COLOR) "Running test with coverage..."
	@go test -race -coverprofile=cover.out -coverpkg=./... ./...
	@go tool cover -html=cover.out -o cover.html

.PHONY: help
.DEFAULT_GOAL := help
help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'