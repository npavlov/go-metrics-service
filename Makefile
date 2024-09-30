# Define Go command, which can be overridden
GO ?= go

build-server:
	$(GO) build -o bin/server ./cmd/agent/*.go

build-agent:
	$(GO) build -o bin/agent ./cmd/agent/*.go

# Run tests
test:
	$(GO) test ./... -coverprofile cover.out

check-coverage:
	$(GO) tool cover -html cover.out

# Clean the binary
clean:
	rm -rf bin/

run-server:
	$(GO) run ./cmd/server/*.go

run-client:
	$(GO) run ./cmd/server/*.go

lint:
	golangci-lint run ./...

# Format the code
fmt:
	$(GO) fmt ./...

# Check for updates on Go dependencies
deps:
	$(GO) get -u ./...

# Default target when just 'make' is run
all: fmt lint build-agent build-server