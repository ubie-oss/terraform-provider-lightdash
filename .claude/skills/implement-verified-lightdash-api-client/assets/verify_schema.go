package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// This script fetches a live API response to verify the schema.
// Usage: LIGHTDASH_API_KEY=your_key LIGHTDASH_URL=https://app.lightdash.cloud go run verify_schema.go /api/v1/your/endpoint
func main() {
	apiKey := os.Getenv("LIGHTDASH_API_KEY")
	baseURL := os.Getenv("LIGHTDASH_URL")
	if baseURL == "" {
		baseURL = "https://app.lightdash.cloud"
	}

	if apiKey == "" {
		fmt.Println("Error: LIGHTDASH_API_KEY environment variable is not set.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_schema.go <endpoint_path>")
		os.Exit(1)
	}
	endpoint := os.Args[1]
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	url := baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", "ApiKey "+apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Println("Response Body:")
	fmt.Println(string(body))
}
