COLOR 	:= "\e[1;36m%s\e[0m\n"
RED 	:= "\e[1;31m%s\e[0m\n"

.PHONY: deps
deps: ## install dependencies
	go mod download
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.2.2

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: lint
lint: ## lint
	@printf $(COLOR) "Linting..."
	@[ ! -e .golangci.yml ] || golangci-lint run
	@[ ! -e "$(REPO_ROOT)/.golangci.yml" ] || { printf $(COLOR) "Using root .golangci.yml" ; golangci-lint run -c "$(REPO_ROOT)/.golangci.yml"; }

.PHONY: feature-lint
feature-lint: ## feature-lint
	npm run lint:gherkin

.PHONY: fmt
fmt: tidy ## tidy, format and imports
	gofumpt -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	goimports -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	gci write --skip-generated -s standard -s default -s "prefix(github.com/robmorgan/infraspec)" .

.PHONY: build
build: ## build infraspec binary
	@printf $(COLOR) "Building infraspec..."
	go build -o bin/infraspec ./cmd/infraspec

.PHONY: cloudmirror
cloudmirror: ## build cloudmirror tool (maintainer-only)
	@printf $(COLOR) "Building cloudmirror..."
	go build -o bin/cloudmirror ./tools/cloudmirror/cmd/cloudmirror

.PHONY: test
test: ## run all unit tests
	@printf $(COLOR) "Running tests..."
	@go test -v $$(go list ./... | grep -v '/test$$')

.PHONY: go-test-cover
go-test-cover: ## run test & generate coverage
	@printf $(COLOR) "Running test with coverage..."
	@go test -race -coverprofile=cover.out -coverpkg=./... ./...
	@go tool cover -html=cover.out -o cover.html

.PHONY: help
.DEFAULT_GOAL := help
help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'