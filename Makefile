# Define Go command, which can be overridden
GO ?= go

include Makefile.local

# Default target: formats code, runs the linter, and builds both agent and server binaries
.PHONY: all
all: fmt lint build-agent build-server

# ----------- build Commands -----------
# build the server binary from Go source files in cmd/server directory
.PHONY: build-server

BUILDINFO_PKG_SERVER=github.com/npavlov/go-metrics-service/internal/server/buildinfo
BUILD_FLAGS_SERVER=-X '$(BUILDINFO_PKG_SERVER).Version=1.0.0' \
            -X '$(BUILDINFO_PKG_SERVER).Date=$(shell date -u +%Y-%m-%d)' \
            -X '$(BUILDINFO_PKG_SERVER).Commit=$(shell git rev-parse HEAD)'

build-server:
	$(GO) build -gcflags="all=-N -l" -ldflags="${BUILD_FLAGS_SERVER}" -o bin/server ${CURDIR}/cmd/server/main.go

# build the agent binary from Go source files in cmd/agent directory
.PHONY: build-agent
BUILDINFO_PKG_AGENT=github.com/npavlov/go-metrics-service/internal/agent/buildinfo
BUILD_FLAGS_AGENT=-X '$(BUILDINFO_PKG_AGENT).Version=1.0.0' \
            -X '$(BUILDINFO_PKG_AGENT).Date=$(shell date -u +%Y-%m-%d)' \
            -X '$(BUILDINFO_PKG_AGENT).Commit=$(shell git rev-parse HEAD)'

build-agent:
	$(GO) build -gcflags="all=-N -l" -ldflags="${BUILD_FLAGS_AGENT}" -o bin/agent ${CURDIR}/cmd/agent/main.go

# build multi checker
.PHONY: build-checker
build-checker:
	$(GO) build -gcflags="all=-N -l" -o bin/checker ${CURDIR}/cmd/staticlint/*.go

# ----------- Test Commands -----------
# Run all tests and generate a coverage profile (coverage.out)
.PHONY: test
test:
	$(GO) test ./... -race -coverprofile=coverage.out -covermode=atomic

# View the test coverage report in HTML format
.PHONY: check-coverage
check-coverage:
	$(GO) tool cover -html=coverage.out

# ----------- Clean Command -----------
# Clean the bin directory by removing all generated binaries
.PHONY: clean
clean:
	rm -rf bin/

# ----------- Run Commands -----------
# Run the server directly from Go source files in cmd/server directory
.PHONY: run-server
run-server:
	$(GO) run -ldflags="${BUILD_FLAGS_SERVER}" ${CURDIR}/cmd/server/main.go

# Run the agent directly from Go source files in cmd/agent directory
.PHONY: run-agent
run-agent:
	$(GO) run -ldflags="${BUILD_FLAGS_AGENT}" ${CURDIR}/cmd/agent/main.go
# ----------- Lint and Format Commands -----------
# Run the linter (golangci-lint) on all Go files in the project
.PHONY: lint
lint:
	golangci-lint run ./...

# Run the linter and automatically fix issues
.PHONY: lint-fix
lint-fix:
	golangci-lint run ./... --fix

# Format all Go files in the project using the built-in Go formatting tool
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Format all Go files in the project using gofumpt for strict formatting rules
.PHONY: gofumpt
gofumpt:
	gofumpt -l -w .

# ----------- Dependency Management -----------
# Update all Go module dependencies
.PHONY: deps
deps:
	$(GO) get -u ./...

# ----------- Database Migration Commands -----------
# Create a new migration using Atlas
.PHONY: atlas-migration
atlas-migration:
	atlas migrate diff $(MIGRATION_NAME) --env dev

# Apply migrations using Goose
.PHONY: goose-up
goose-up:
	goose -dir migrations postgres "$(DATABASE_DSN)" up

# Rollback migrations using Goose
.PHONY: goose-down
goose-down:
	goose -dir migrations postgres "$(DATABASE_DSN)" down

.PHONY: collect-profile
collect-profile:
	bash collect_profiles.sh

.PHONY: open-cpu-server-profile
open-cpu-server-profile:
	go tool pprof -http=":9090" ./profiles/cpu_server.pprof

.PHONY: open-heap-server-profile
open-heap-server-profile:
	go tool pprof -http=":9090" ./profiles/heap_server.pprof

.PHONY: compare-cpu-profiles
compare-cpu-profiles:
	go tool pprof -http=":9090" -diff_base=./profiles/cpu_server_old.pprof ./profiles/cpu_server.pprof

.PHONY: compare-head-profiles
compare-head-profiles:
	go tool pprof -http=":9090" -diff_base=./profiles/heap_server_old.pprof ./profiles/heap_server.pprof

# Run only benchmarks and collect memory allocation statistics
.PHONY: bench-mem
bench-mem:
	$(GO) test ./... -bench=. -benchmem

#Generate swagger
.PHONY: generate-swagger
generate-swagger:
	swag init -g ../../cmd/server/main.go -o ./swagger/ -d ./internal/server/ --parseDependency

# Run the Go documentation server
.PHONY: godoc
godoc:
	godoc -http=:6060

# Generate proto contracts

.PHONY: buf-generate

buf-generate:
	buf generate

# Lint proto contracts

.PHONY: buf-lint

buf-lint:
	buf lint