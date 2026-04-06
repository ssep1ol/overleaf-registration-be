package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func validateRecaptcha(token string) (bool, error) {
	secret := os.Getenv("TURNSTILE_SECRET_KEY")
	if secret == "" {
		return false, fmt.Errorf("TURNSTILE_SECRET_KEY is not set")
	}

	log.Printf("Validating CAPTCHA token: %s", token)

	url := "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	data := map[string][]string{
		"secret":   {secret},
		"response": {token},
	}

	resp, err := http.PostForm(url, data)
	if err != nil {
		return false, fmt.Errorf("failed to verify CAPTCHA: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode CAPTCHA response: %v", err)
	}

	log.Printf("CAPTCHA validation result: %v", result.Success)
	return result.Success, nil
}
