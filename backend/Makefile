.PHONY: proto generate run install-deps clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  make install-deps  - Install required dependencies (protoc, protoc-gen-go, protoc-gen-go-grpc)"
	@echo "  make proto         - Generate Go code from .proto files"
	@echo "  make run           - Run the gRPC and HTTP server"
	@echo "  make clean         - Clean generated files"
	@echo "  make help          - Show this help message"

# Install dependencies
install-deps:
	@echo "Installing protoc compiler..."
	@which protoc > /dev/null || (echo "Please install protoc: https://grpc.io/docs/protoc-installation/" && exit 1)
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	@echo "Go bin directory: $(GOBIN)"
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed!"
	@echo ""
	@echo "Note: If you get 'program not found' errors, add this to your ~/.zshrc:"
	@echo "  export PATH=\"$(GOBIN):\$$PATH\""
	@echo "Then run: source ~/.zshrc"

# Get Go bin directory
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
ifeq ($(GOPATH),)
	GOBIN := $(HOME)/go/bin
endif

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@echo "Using Go bin directory: $(GOBIN)"
	@if [ ! -f "$(GOBIN)/protoc-gen-go" ]; then \
		echo "Error: protoc-gen-go not found. Run 'make install-deps' first."; \
		exit 1; \
	fi
	@if [ ! -f "$(GOBIN)/protoc-gen-go-grpc" ]; then \
		echo "Error: protoc-gen-go-grpc not found. Run 'make install-deps' first."; \
		exit 1; \
	fi
	@PATH="$(GOBIN):$$PATH" protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/greeting.proto
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
