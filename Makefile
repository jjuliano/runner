# Variables
PROJECT_NAME = kdeps
TEST_REPORT = test-report.out
COVERAGE_REPORT = coverage.out
PACKAGE_LIST = ./...

# Default target
all: test

# Run tests and generate a report
test:
	@echo "Running tests..."
	@go test -v $(PACKAGE_LIST) | tee $(TEST_REPORT)
	@go test -coverprofile=$(COVERAGE_REPORT) $(PACKAGE_LIST)

# Build the project
build:
	@echo "Building project..."
	@go build -o $(PROJECT_NAME)

# Clean up generated files
clean:
	@echo "Cleaning up..."
	@rm -f $(PROJECT_NAME) $(TEST_REPORT) $(COVERAGE_REPORT)

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

# Additional targets for cross-compilation
linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(PROJECT_NAME)_linux_amd64 $(PACKAGE_LIST)

darwin:
	@echo "Building for macOS (Darwin)..."
	@GOOS=darwin GOARCH=amd64 go build -o $(PROJECT_NAME)_darwin_amd64 $(PACKAGE_LIST)

windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(PROJECT_NAME)_windows_amd64.exe $(PACKAGE_LIST)

arm:
	@echo "Building for ARM..."
	@GOOS=linux GOARCH=arm go build -o $(PROJECT_NAME)_linux_arm $(PACKAGE_LIST)

.PHONY: all test build clean lint fmt vet coverage linux darwin windows arm
# Variables
PROJECT_NAME = kdeps
TEST_REPORT = test-report.out
COVERAGE_REPORT = coverage.out
PACKAGE_LIST = ./...

# Default target
all: test

# Run tests and generate a report
test:
	@echo "Running tests..."
	@go test -v $(PACKAGE_LIST) | tee $(TEST_REPORT)
	@go test -coverprofile=$(COVERAGE_REPORT) $(PACKAGE_LIST)

# Build the project
build:
	@echo "Building project..."
	@go build -o $(PROJECT_NAME)

# Clean up generated files
clean:
	@echo "Cleaning up..."
	@rm -f $(PROJECT_NAME) $(TEST_REPORT) $(COVERAGE_REPORT)

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

# Additional targets for cross-compilation
linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(PROJECT_NAME)_linux_amd64 $(PACKAGE_LIST)

darwin:
	@echo "Building for macOS (Darwin)..."
	@GOOS=darwin GOARCH=amd64 go build -o $(PROJECT_NAME)_darwin_amd64 $(PACKAGE_LIST)

windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(PROJECT_NAME)_windows_amd64.exe $(PACKAGE_LIST)

arm:
	@echo "Building for ARM..."
	@GOOS=linux GOARCH=arm go build -o $(PROJECT_NAME)_linux_arm $(PACKAGE_LIST)

.PHONY: all test build clean lint fmt vet coverage linux darwin windows arm
