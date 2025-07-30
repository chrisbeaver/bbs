.PHONY: build setup run clean test

# Build the BBS server
build:
	go build -o bbs main.go

# Setup the database with initial data
setup:
	go run cmd/setup/main.go

# Run the BBS server
run:
	go run main.go

# Run setup then server
start: setup run

# Clean build artifacts
clean:
	rm -f bbs bbs.db

# Run tests (when implemented)
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golint ./...

# Development setup
dev-setup: deps setup
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the BBS server"
	@echo "Connect via: ssh -p 2323 sysop@localhost"
