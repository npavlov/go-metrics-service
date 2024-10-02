.PHONY: build-server build-agent test clean run-server run-agent lint fmt deps

# Define Go command, which can be overridden
GO ?= go

# Build the server binary from the Go source files in the cmd/server directory
build-server:
	$(GO) build -o bin/server ./cmd/server/*.go

# Build the agent binary from the Go source files in the cmd/agent directory
build-agent:
	$(GO) build -o bin/agent ./cmd/agent/*.go

# Run all tests and generate a coverage profile (cover.out)
test:
	$(GO) test ./... -race -coverprofile=coverage.out -covermode=atomic

# View the test coverage report in HTML format
check-coverage:
	$(GO) tool cover -html coverage.out

# Clean the bin directory by removing all generated binaries
clean:
	rm -rf bin/

# Run the server directly from the Go source files in the cmd/server directory
run-server:
	$(GO) run ${CURDIR}/cmd/server/*.go

# Run the client directly from the Go source files in the cmd/client directory
# Why CURDIR - https://stackoverflow.com/questions/52437728/bash-what-is-the-difference-between-pwd-and-curdir
run-agent:
	$(GO) run ${CURDIR}/cmd/agent/*.go

# Run the linter (golangci-lint) on all Go files in the project to check for coding issues
lint:
	golangci-lint run ./...

# Format all Go files in the project using the built-in Go formatting tool
fmt:
	$(GO) fmt ./...

# Check for updates on Go module dependencies and update them if necessary
deps:
	$(GO) get -u ./...

# Default target when 'make' is run, it formats code, runs the linter, and builds both the agent and server binaries
all: fmt lint build-agent build-server