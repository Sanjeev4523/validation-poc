package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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

	// Start server on port 8080
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("Hello world route available at http://localhost%s/hello", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
