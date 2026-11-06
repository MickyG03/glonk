.DEFAULT_GOAL := help
.PHONY: help fmt test vet run build check

help:
	@echo "Available targets:"
	@echo "  fmt     - Format Go code"
	@echo "  test    - Run tests"
	@echo "  vet     - Check for suspicious code with go vet"
	@echo "  run     - Run an example (default bounce2d; EXAMPLE=./examples/cube3d)"
	@echo "  build   - Compile every package"
	@echo "  check   - Run tests and vet"

fmt:
	@echo "Formatting..."
	go fmt ./...

test:
	@echo "Testing..."
	go test ./...

vet:
	@echo "Vetting..."
	go vet ./...

EXAMPLE ?= ./examples/bounce2d
run:
	@echo "Running..."
	go run $(EXAMPLE)

build:
	@echo "Building..."
	go build ./...

check: test vet
	@echo "All checks passed!"
