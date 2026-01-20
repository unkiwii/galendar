# Makefile for galendar

# Variables
BINARY_NAME=galendar
MAIN_PACKAGE=./cmd/galendar
INSTALL_DIR=/usr/local/bin
GO=go

# Default target
.DEFAULT_GOAL := build

# Build the project
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BINARY_NAME)"

# Install the executable to a sensible location (requires sudo)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo mkdir -p $(INSTALL_DIR)
	@sudo cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

# Run the program (builds temporary binary, runs it, then deletes it)
# Usage: make run ARGS="--year 2024 --month 5"
.PHONY: run
run:
	@TMP_BINARY=/tmp/galendar-$$$$; \
	echo "Building temporary binary..."; \
	$(GO) build -o $$TMP_BINARY $(MAIN_PACKAGE); \
	echo "Running $(BINARY_NAME)..."; \
	$$TMP_BINARY $(ARGS); \
	EXIT_CODE=$$?; \
	rm -f $$TMP_BINARY; \
	exit $$EXIT_CODE

# Clean temporary files (executable and generated output files)
.PHONY: clean
clean:
	@echo "Cleaning temporary files..."
	@rm -f $(BINARY_NAME)
	@rm -f *.pdf *.svg
	@echo "Clean complete"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make          - Build the project (default)"
	@echo "  make install  - Build and install to $(INSTALL_DIR)"
	@echo "  make run      - Build and run (use ARGS='--year 2024' to pass arguments)"
	@echo "  make clean    - Remove build artifacts and generated files"
	@echo "  make help     - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run ARGS=\"--year 2024\""
	@echo "  make run ARGS=\"--month 5 --year 2024 --output svg\""
