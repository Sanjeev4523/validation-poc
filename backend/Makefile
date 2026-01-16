.PHONY: proto generate run install-deps clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  make install-deps  - Check for buf CLI installation"
	@echo "  make proto         - Generate Go code from .proto files using buf"
	@echo "  make run           - Run the gRPC and HTTP server"
	@echo "  make clean         - Clean generated files"
	@echo "  make help          - Show this help message"

# Install dependencies
install-deps:
	@echo "Checking for buf CLI..."
	@which buf > /dev/null || (echo "Error: buf CLI not found. Please install buf: https://buf.build/docs/installation" && exit 1)
	@echo "buf CLI found: $$(buf --version)"
	@echo "Dependencies check complete!"

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files using buf..."
	@which buf > /dev/null || (echo "Error: buf CLI not found. Run 'make install-deps' first." && exit 1)
	@buf generate
	@echo "Proto code generated successfully!"

# Run the server
run:
	@echo "Starting server..."
	@go run main.go

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -f proto/*.pb.go
	@echo "Clean complete!"
