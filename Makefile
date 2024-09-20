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
	@for target in $(TARGETS); do \
		GOOS=$$(echo $$target | cut -d '/' -f 1); \
		GOARCH=$$(echo $$target | cut -d '/' -f 2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		mkdir -p ./build/$(PROJECT_NAME)_$$GOOS_$$GOARCH; \
		if [ "$$GOOS" = "android" ]; then \
			export CGO_ENABLED=1; \
			export ANDROID_NDK_HOME=/android/ndk; \
		elif [ "$$GOOS" = "ios" ]; then \
			export CGO_ENABLED=1; \
			export CC=clang; \
		elif [ "$$GOOS" = "illumos" ]; then \
			export CGO_ENABLED=1; \
			export CC=gcc; \
			export CFLAGS="-D_XOPEN_SOURCE=600"; \
		else \
			export CGO_ENABLED=0; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH with CGO_ENABLED=$$CGO_ENABLED"; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o ./build/$(PROJECT_NAME)_$$GOOS_$$GOARCH/$(PROJECT_NAME) $(PACKAGE_LIST) || { echo "Skipping $$GOOS/$$GOARCH due to build errors."; continue; }; \
		file ./build/$(PROJECT_NAME)_$$GOOS_$$GOARCH/$(PROJECT_NAME); \
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
