package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/rs/cors"
)

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func main() {
	// Load environment variables
	loadEnv()

	// Validate required environment variables
	requiredVars := []string{"ALLOWED_DOMAINS", "TURNSTILE_SECRET_KEY", "OL_INSTANCE", "OL_ADMIN_EMAIL", "OL_ADMIN_PASSWORD"}
	if err := validateEnvVars(requiredVars); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", registerHandler)

	// Apply CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Replace "*" with specific frontend origins, if needed
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap your handler with CORS middleware
	handler := c.Handler(mux)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
