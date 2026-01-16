package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"validation-service/backend/config"
	"validation-service/backend/handler"
	"validation-service/backend/logger"
	"validation-service/backend/proto"
	"validation-service/backend/service"

	"buf.build/go/protovalidate"
)

// greetingServer implements the GreetingServiceServer interface
type greetingServer struct {
	proto.UnimplementedGreetingServiceServer
	validator protovalidate.Validator
}

// SayHello implements the SayHello RPC method
func (s *greetingServer) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	// Validate the request
	if err := s.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	return &proto.HelloResponse{
		Message: "hello " + req.GetName(),
	}, nil
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create response
	response := map[string]string{
		"message": "Hello, World!",
	}

	// Encode and send JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Load environment variables from .env file
	// This must be called before any code that reads environment variables
	if err := config.LoadEnv(); err != nil {
		// Non-fatal: if .env doesn't exist, we'll use system environment variables
	}

	// Initialize logger with level from environment variable
	logger.Init()

	logger.Info("Starting validation service...")

	// Create validator instance
	logger.Debug("Initializing protovalidate validator...")
	validator, err := protovalidate.New()
	if err != nil {
		logger.Fatal("Failed to create validator: %v", err)
	}
	logger.Info("Protovalidate validator initialized successfully")

	// Get base path (directory where main.go is located)
	logger.Debug("Resolving base path...")
	basePath, err := filepath.Abs(".")
	if err != nil {
		logger.Fatal("Failed to get base path: %v", err)
	}
	logger.Debug("Base path resolved to: %s", basePath)

	// Parse BSR configuration from buf.yaml
	logger.Debug("Parsing BSR configuration from buf.yaml...")
	bsrOrg, bsrModule := service.GetBSRConfig(basePath)
	logger.Info("BSR configuration: org=%s, module=%s", bsrOrg, bsrModule)

	// Get schema source mode from environment variable for schema service
	schemaSourceMode := config.GetSchemaSourceMode("schema")
	logger.Info("Schema source mode: %d", schemaSourceMode)

	// Initialize schema service
	logger.Debug("Initializing schema service...")
	schemaService := service.NewSchemaService(bsrOrg, bsrModule, basePath, schemaSourceMode)
	logger.Info("Schema service initialized successfully with mode=%d", schemaSourceMode)

	// Initialize schema handler
	logger.Debug("Initializing schema handler...")
	schemaHandler := handler.NewSchemaHandler(schemaService)
	logger.Info("Schema handler initialized successfully")

	// Get BSR token for validation service
	bsrToken := config.GetEnv("BUF_TOKEN", "")
	if bsrToken == "" {
		logger.Warn("BUF_TOKEN is not set. BSR requests may fail for private repositories.")
	} else {
		logger.Debug("BUF_TOKEN is set (length: %d)", len(bsrToken))
	}

	// Get validation source mode from environment variable for validation service
	validationSourceMode := config.GetSchemaSourceMode("validation")
	logger.Info("Validation source mode: %d", validationSourceMode)

	// Initialize validation service
	logger.Debug("Initializing validation service...")
	validationService := service.NewValidationService(validator, validationSourceMode, bsrOrg, bsrModule, bsrToken)
	logger.Info("Validation service initialized successfully with mode=%d", validationSourceMode)

	// Initialize validation handler
	logger.Debug("Initializing validation handler...")
	validationHandler := handler.NewValidationHandler(validationService)
	logger.Info("Validation handler initialized successfully")

	// Initialize commits service
	logger.Debug("Initializing commits service...")
	commitsService := service.NewCommitsService(bsrOrg, bsrModule, bsrToken)
	logger.Info("Commits service initialized successfully")

	// Initialize commits handler
	logger.Debug("Initializing commits handler...")
	commitsHandler := handler.NewCommitsHandler(commitsService)
	logger.Info("Commits handler initialized successfully")

	// Start gRPC server in a goroutine
	go func() {
		logger.Debug("Starting gRPC server on port :50051...")
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			logger.Fatal("Failed to listen on port 50051: %v", err)
		}

		s := grpc.NewServer()
		proto.RegisterGreetingServiceServer(s, &greetingServer{
			validator: validator,
		})

		logger.Info("gRPC server starting on port :50051")
		if err := s.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC server: %v", err)
		}
	}()

	// Register the hello world route with CORS
	logger.Debug("Registering HTTP routes...")
	http.HandleFunc("/hello", corsMiddleware(helloHandler))
	logger.Debug("Registered route: GET /hello")

	// Register schema API route with CORS
	http.HandleFunc("/api/v1/schema/", corsMiddleware(schemaHandler.GetSchema))
	logger.Debug("Registered route: GET /api/v1/schema/{messageName}")

	// Register proto files list API route with CORS
	http.HandleFunc("/api/v1/proto-files", corsMiddleware(schemaHandler.ListProtoFiles))
	logger.Debug("Registered route: GET /api/v1/proto-files")

	// Register validation API route with CORS
	http.HandleFunc("/api/v1/validate-proto", corsMiddleware(validationHandler.ValidateProto))
	logger.Debug("Registered route: POST /api/v1/validate-proto")

	// Register commits API route with CORS
	http.HandleFunc("/api/v1/commits", corsMiddleware(commitsHandler.GetCommits))
	logger.Debug("Registered route: GET /api/v1/commits")

	// Also register root route for convenience with CORS
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		helloHandler(w, r)
	}))
	logger.Debug("Registered route: GET /")

	// Start HTTP server on port 8080
	port := ":8080"
	logger.Info("HTTP server starting on port %s", port)
	logger.Info("Hello world route available at http://localhost%s/hello", port)
	logger.Info("Schema API route available at http://localhost%s/api/v1/schema/{messageName}", port)
	logger.Info("Proto files API route available at http://localhost%s/api/v1/proto-files", port)
	logger.Info("Validation API route available at http://localhost%s/api/v1/validate-proto", port)
	logger.Info("Commits API route available at http://localhost%s/api/v1/commits", port)
	logger.Info("Validation service started successfully")

	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Fatal("Server failed to start: %v", err)
	}
}
