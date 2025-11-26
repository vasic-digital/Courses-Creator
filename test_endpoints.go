package main

import (
	"fmt"
	"net/http"
)

const baseURL = "http://localhost:8080"

func main() {
	// Test various endpoints
	endpoints := []string{
		"/api/v1/health",
		"/health",
		"/api/health",
		"/",
	}
	
	for _, endpoint := range endpoints {
		url := baseURL + endpoint
		fmt.Printf("Testing %s...\n", url)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		fmt.Printf("  Status: %d\n", resp.StatusCode)
		if resp.StatusCode == http.StatusOK {
			fmt.Println("  ✓ This endpoint works!")
		}
	}
}