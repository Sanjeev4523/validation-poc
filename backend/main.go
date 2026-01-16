package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buf.build/go/protovalidate"
	"validation-service/backend/proto"
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
	// Create validator instance
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v", err)
		}

		s := grpc.NewServer()
		proto.RegisterGreetingServiceServer(s, &greetingServer{
			validator: validator,
		})

		log.Printf("gRPC server starting on port :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Register the hello world route
	http.HandleFunc("/hello", helloHandler)

	// Also register root route for convenience
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		helloHandler(w, r)
	})

	// Start HTTP server on port 8080
	port := ":8080"
	log.Printf("HTTP server starting on port %s", port)
	log.Printf("Hello world route available at http://localhost%s/hello", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
