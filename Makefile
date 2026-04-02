.PHONY: all build test fmt clean install lint help example-simple example-opencode example-codex example-aider example-compare example-compare-thorough-1 example-compare-thorough-3 test-quick

BINARY := cmb
BINDIR := bin
MAIN := ./cmd/cmb

# Default target
all: build

# Build the binary
build:
	@echo "Building cmb..."
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/$(BINARY) $(MAIN)
	@echo "Build complete: ./$(BINDIR)/$(BINARY)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
	@echo "Format complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BINDIR)
	@rm -f coverage.out coverage.html
	@rm -rf .cmb/
	@go clean
	@echo "Clean complete"

# Install the binary
install:
	@echo "Installing cmb..."
	@go install ./cmd/cmb
	@echo "Install complete"

# Lint code
lint:
	@echo "Linting code..."
	@go vet ./...
	@echo "Lint complete"

# Run golangci-lint if available
lint-full:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, running go vet..."; \
		go vet ./...; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies ready"

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "Verification complete"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BINDIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BINDIR)/$(BINARY)-linux-amd64 $(MAIN)
	@GOOS=darwin GOARCH=amd64 go build -o $(BINDIR)/$(BINARY)-darwin-amd64 $(MAIN)
	@GOOS=darwin GOARCH=arm64 go build -o $(BINDIR)/$(BINARY)-darwin-arm64 $(MAIN)
	@GOOS=windows GOARCH=amd64 go build -o $(BINDIR)/$(BINARY)-windows-amd64.exe $(MAIN)
	@echo "Multi-platform build complete"

# Run simple example (quick test with Claude Code)
example-simple:
	@echo "Running simple example with Claude Code..."
	@./bin/cmb run --agent claude-code --task task/test-simple.yaml --no-sandbox

# Run quick test (no sandbox, faster for development)
test-quick:
	@echo "Running quick test (no sandbox)..."
	@./bin/cmb run --agent claude-code --task task/test-simple.yaml --no-sandbox

# Run simple example with OpenCode
example-opencode:
	@echo "Running simple example with OpenCode..."
	@./bin/cmb run --agent opencode --task task/test-simple.yaml

# Run simple example with Codex
example-codex:
	@echo "Running simple example with Codex..."
	@./bin/cmb run --agent codex --task task/test-simple.yaml

# Run simple example with Aider
example-aider:
	@echo "Running simple example with Aider..."
	@./bin/cmb run --agent aider --task task/test-simple.yaml

# Run quick comparison test with all agents (single run each)
example-compare:
	@echo "Running comparison test with all 4 agents (1 run each)..."
	@./bin/cmb run \
		--agent opencode \
		--agent claude-code \
		--agent codex \
		--agent aider \
		--task task/test-simple.yaml

# Run thorough comparison with single run (all agents, detailed)
example-compare-thorough-1:
	@echo "Running thorough comparison (all 4 agents, 1 run each)..."
	@./bin/cmb run \
		--agent opencode \
		--agent claude-code \
		--agent codex \
		--agent aider \
		--task task/refactor-auth.yaml \
		--show-diff

# Run thorough comparison with multiple runs (for variance testing)
example-compare-thorough-3:
	@echo "Running thorough comparison (all 4 agents, 3 runs each)..."
	@./bin/cmb run \
		--agent opencode \
		--agent claude-code \
		--agent codex \
		--agent aider \
		--task task/refactor-auth.yaml \
		--runs 3

# Show help
help:
	@echo "CodematicBench Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build            - Build the cmb binary"
	@echo "  make test             - Run tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make test-race        - Run tests with race detector"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make install          - Install cmb to GOPATH/bin"
	@echo "  make fmt              - Format Go code"
	@echo "  make lint             - Lint code (go vet)"
	@echo "  make lint-full        - Lint code (golangci-lint)"
	@echo "  make deps             - Download and tidy dependencies"
	@echo "  make verify           - Verify dependencies"
	@echo "  make build-all        - Build for multiple platforms"
	@echo ""
	@echo "Examples:"
	@echo "  make test-quick                - Quick test (no sandbox, fastest)"
	@echo "  make example-simple            - Run simple test with Claude Code"
	@echo "  make example-opencode          - Run simple test with OpenCode"
	@echo "  make example-codex             - Run simple test with Codex"
	@echo "  make example-aider             - Run simple test with Aider"
	@echo ""
	@echo "Comparisons:"
	@echo "  make example-compare           - Compare all 4 agents (1 run each)"
	@echo "  make example-compare-thorough-1- Compare all 4 agents (1 run, with diffs)"
	@echo "  make example-compare-thorough-3- Compare all 4 agents (3 runs, variance test)"
	@echo ""
	@echo "  make example                   - Run complex example task"
	@echo ""
	@echo "  make help             - Show this help message"
