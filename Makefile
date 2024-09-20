# Variables
PROJECT_NAME := runner
TEST_REPORT := test-report.out
COVERAGE_REPORT := coverage.out
# Update PACKAGE_LIST to point to the correct package directory
PACKAGE_LIST := ./

# List of GOOS/GOARCH pairs for macOS, Windows, and Linux
TARGETS := $(shell go tool dist list)

# Default target
all: test

# Run tests and generate a report
test:
	@echo "Running tests..."
	@go test -v $(PACKAGE_LIST) | tee $(TEST_REPORT)
	@go test -coverprofile=$(COVERAGE_REPORT) $(PACKAGE_LIST)

# Build the project for all targets
build: $(TARGETS)

# Build for each target
$(TARGETS):
	@echo "Building for $@..."
	@GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
	mkdir -p ./build/$(PROJECT_NAME)_$(word 1,$(subst /, ,$@))_$(word 2,$(subst /, ,$@)) && \
	go build -o ./build/$(PROJECT_NAME)_$(word 1,$(subst /, ,$@))_$(word 2,$(subst /, ,$@))/$(PROJECT_NAME) $(PACKAGE_LIST)

# Clean up generated files
clean:
	@echo "Cleaning up..."
	@rm -rf ./build $(TEST_REPORT) $(COVERAGE_REPORT)

# Run linting using golangci-lint (you need to have golangci-lint installed)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt $(PACKAGE_LIST)

# Run vet
vet:
	@echo "Running vet..."
	@go vet $(PACKAGE_LIST)

# Display coverage in browser (you need to have go tool cover installed)
coverage: test
	@go tool cover -html=$(COVERAGE_REPORT)

.PHONY: all test build clean lint fmt vet coverage $(TARGETS)
