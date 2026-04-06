package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// RegistrationRequest represents the incoming registration request.
type RegistrationRequest struct {
	Email   string `json:"email"`
	Captcha string `json:"captcha"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Received email: %s", req.Email)
	log.Printf("Received CAPTCHA token: %s", req.Captcha)

	// Validate email domain
	if !validateEmailDomain(req.Email) {
		http.Error(w, "Email domain is not allowed", http.StatusBadRequest)
		return
	}

	// Validate CAPTCHA token
	valid, err := validateRecaptcha(req.Captcha)
	if err != nil {
		log.Printf("CAPTCHA validation failed: %v", err)
		http.Error(w, fmt.Sprintf("CAPTCHA validation failed: %v", err), http.StatusInternalServerError)
		return
	}
	if !valid {
		log.Println("Invalid CAPTCHA token.")
		http.Error(w, "Invalid CAPTCHA", http.StatusUnauthorized)
		return
	}

	// Initialize Overleaf instance
	baseURL := os.Getenv("OL_INSTANCE")
	if baseURL == "" {
		http.Error(w, "OL_INSTANCE is not set", http.StatusInternalServerError)
		return
	}
	overleaf := NewOverleaf(baseURL)

	// Admin login using environment variables
	adminEmail := os.Getenv("OL_ADMIN_EMAIL")
	adminPassword := os.Getenv("OL_ADMIN_PASSWORD")
	if adminEmail == "" || adminPassword == "" {
		http.Error(w, "OL_ADMIN_EMAIL or OL_ADMIN_PASSWORD is not set", http.StatusInternalServerError)
		return
	}

	// Log in as admin
	if err := overleaf.Login(adminEmail, adminPassword); err != nil {
		log.Printf("Admin login failed: %v", err)
		http.Error(w, fmt.Sprintf("Admin login failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Register the user with the provided email
	if err := overleaf.RegisterUser(req.Email); err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully. Please check your email to set your password.",
	})
}

func validateEmailDomain(email string) bool {
	allowedDomains := os.Getenv("ALLOWED_DOMAINS")
	if allowedDomains == "" {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false // Invalid email format
	}

	domain := parts[1]
	allowed := strings.Split(allowedDomains, ",")
	for _, allowedDomain := range allowed {
		if strings.TrimSpace(domain) == strings.TrimSpace(allowedDomain) {
			return true
		}
	}
	return false
}

func validateEnvVars(vars []string) error {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("environment variable %s is not set", v)
		}
	}
	return nil
}
