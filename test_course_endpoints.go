package main

import (
	"fmt"
	"net/http"
)

const baseURL = "http://localhost:8080"

func main() {
	// Test course endpoints
	endpoints := []string{
		"/courses/generate",
		"/api/courses/generate", 
		"/api/v1/courses/generate",
		"/courses",
		"/api/courses",
		"/api/v1/courses",
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
		if resp.StatusCode != 404 {
			fmt.Printf("  ✓ Endpoint found (status: %d)\n", resp.StatusCode)
		}
	}
}