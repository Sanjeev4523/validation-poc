# Validation Service Backend

A Go backend service with gRPC and HTTP endpoints.

## Prerequisites

- Go 1.25.5 or later
- Protocol Buffers compiler (`protoc`)
- Go plugins for protoc:
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`

## Setup

### 1. Install Protocol Buffers Compiler

**macOS:**
```bash
brew install protobuf
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Or download from: https://github.com/protocolbuffers/protobuf/releases
```

**Windows:**
Download from: https://github.com/protocolbuffers/protobuf/releases

### 2. Install Go Dependencies

```bash
# Install Go modules
go mod download

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Or use the Makefile:
```bash
make install-deps
```

## Development Workflow

### Generate Proto Code

After modifying `.proto` files, regenerate the Go code:

```bash
make proto
```

Or manually:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/greeting.proto
```

### Run the Server

```bash
make run
```

Or:
```bash
go run main.go
```

The server will start:
- **gRPC server** on port `50051`
- **HTTP server** on port `8080`

### Test the Services

**HTTP endpoint:**
```bash
curl http://localhost:8080/hello
```

**gRPC endpoint:**
You can use tools like `grpcurl` or `evans` to test gRPC:
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Test the gRPC service
grpcurl -plaintext localhost:50051 proto.GreetingService/SayHello -d '{"name": "World"}'
```

## Project Structure

```
backend/
├── main.go              # Main server code
├── proto/
│   ├── greeting.proto   # Protocol buffer definition
│   ├── greeting.pb.go   # Generated Go code (do not edit)
│   └── greeting_grpc.pb.go  # Generated gRPC code (do not edit)
├── go.mod               # Go module dependencies
└── Makefile             # Build automation
```

## Common Tasks

- `make help` - Show all available commands
- `make install-deps` - Install required dependencies
- `make proto` - Generate code from proto files
- `make run` - Run the server
- `make clean` - Remove generated files

## Notes

- Never manually edit `*.pb.go` files - they are auto-generated
- Always run `make proto` after modifying `.proto` files
- The server runs both gRPC and HTTP servers concurrently
