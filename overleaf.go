package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Overleaf struct {
	BaseURL string
	Client  *http.Client
	CSRF    string
}

// NewOverleaf initializes a new Overleaf instance with a cookie jar for session management
func NewOverleaf(baseURL string) *Overleaf {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	return &Overleaf{
		BaseURL: baseURL,
		Client:  client,
	}
}

// get performs a GET request and returns the HTTP response
func (o *Overleaf) get(endpoint string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", o.BaseURL, endpoint)
	resp, err := o.Client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %v", err)
	}
	return resp, nil
}

// post performs a POST request with form data
func (o *Overleaf) post(endpoint string, data map[string]string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", o.BaseURL, endpoint)

	// Encode form data using url.Values
	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("POST request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %v", err)
	}
	return resp, nil
}

// obtainCSRF fetches the CSRF token from the specified endpoint
func (o *Overleaf) obtainCSRF(endpoint string) error {
	resp, err := o.get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch page for CSRF token: %v", err)
	}
	defer resp.Body.Close()

	// Parse the HTML to extract the CSRF token
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}

	csrf, exists := doc.Find(`meta[name="ol-csrfToken"]`).Attr("content")
	if !exists {
		return fmt.Errorf("CSRF token not found")
	}
	o.CSRF = csrf
	log.Printf("Obtained CSRF token: %s", csrf)
	return nil
}

// Login logs in as an admin and saves the session
func (o *Overleaf) Login(email, password string) error {
	// Fetch the login page to get the CSRF token
	if err := o.obtainCSRF("/login"); err != nil {
		return fmt.Errorf("failed to obtain CSRF token: %v", err)
	}

	// Send login request
	resp, err := o.post("/login", map[string]string{
		"email":    email,
		"password": password,
		"_csrf":    o.CSRF,
	})
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	log.Println("Admin login successful.")
	return nil
}

// RegisterUser registers a new user using the admin session
func (o *Overleaf) RegisterUser(email string) error {
	// Fetch the admin registration page to get the CSRF token
	if err := o.obtainCSRF("/admin/register"); err != nil {
		return fmt.Errorf("failed to obtain CSRF token for registration: %v", err)
	}

	// Send user registration request
	resp, err := o.post("/admin/register", map[string]string{
		"email": email,
		"_csrf": o.CSRF,
	})
	if err != nil {
		return fmt.Errorf("registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user registration failed with status code: %d", resp.StatusCode)
	}

	log.Println("User registered successfully.")
	return nil
}
