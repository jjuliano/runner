# Variables
PROJECT_NAME := runner
TEST_REPORT := test-report.out
COVERAGE_REPORT := coverage.out
PACKAGE_LIST := ./

TARGETS := $(shell go tool dist list)

# Default target
all: test

# Run tests and generate a report
test:
	@echo "Running tests..."
	@go test -v $(PACKAGE_LIST) | tee $(TEST_REPORT)
	@go test -coverprofile=$(COVERAGE_REPORT) $(PACKAGE_LIST)

build:
	@mkdir -p ./build
	@for target in $(TARGETS); do \
		X_OS=$$(echo $$target | cut -d '/' -f 1); \
		X_ARCH=$$(echo $$target | cut -d '/' -f 2); \
		if [ -z "$$X_OS" ] || [ -z "$$X_ARCH" ]; then \
			echo "Invalid target: $$target resulted in X_OS=$$X_OS and X_ARCH=$$X_ARCH"; \
			continue; \
		fi; \
		echo "Building for $$X_OS/$$X_ARCH..."; \
		mkdir -p ./build/$(PROJECT_NAME)_$$X_OS_$$X_ARCH || { \
			echo "Failed to create directory ./build/$(PROJECT_NAME)_$$X_OS_$$X_ARCH"; \
			continue; \
		}; \
		GOOS=$$X_OS GOARCH=$$X_ARCH go build -o ./build/$$X_OS/$$X_ARCH/$(PROJECT_NAME) $(PACKAGE_LIST) || { \
			echo "Build failed for $$X_OS/$$X_ARCH"; \
			continue; \
		}; \
		echo "Build succeeded for $$X_OS/$$X_ARCH"; \
	done


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
