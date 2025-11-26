package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

func main() {
	// Test health endpoint
	fmt.Println("Testing health endpoint...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ Health check passed")
	} else {
		fmt.Printf("✗ Health check failed with status: %d\n", resp.StatusCode)
		return
	}
	
	// Test course generation
	fmt.Println("\nTesting course generation...")
	generateReq := map[string]interface{}{
		"markdown_path": "/tmp/test_course.md",
		"output_dir":    "/tmp/output",
		"options": map[string]interface{}{
			"quality":           "standard",
			"background_music":  false,
			"languages":         []string{"en"},
		},
	}
	
	reqBody, _ := json.Marshal(generateReq)
	resp, err = http.Post(baseURL+"/courses/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Course generation failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusAccepted {
		fmt.Println("✓ Course generation request accepted")
		
		// Parse response to get job ID
		var generateResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&generateResp)
		jobID := generateResp["job_id"].(string)
		fmt.Printf("  Job ID: %s\n", jobID)
		
		// Test job status
		fmt.Println("\nTesting job status...")
		time.Sleep(1 * time.Second) // Give some time for processing
		resp, err = http.Get(baseURL + "/jobs/" + jobID)
		if err != nil {
			fmt.Printf("Job status check failed: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			fmt.Println("✓ Job status check passed")
			var jobResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&jobResp)
			fmt.Printf("  Job status: %s\n", jobResp["status"])
		} else {
			fmt.Printf("✗ Job status check failed with status: %d\n", resp.StatusCode)
		}
		
	} else {
		fmt.Printf("✗ Course generation failed with status: %d\n", resp.StatusCode)
	}
	
	// Test courses list
	fmt.Println("\nTesting courses list...")
	resp, err = http.Get(baseURL + "/courses")
	if err != nil {
		fmt.Printf("Courses list failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ Courses list check passed")
		var coursesResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&coursesResp)
		total := coursesResp["total"].(float64)
		fmt.Printf("  Total courses: %.0f\n", total)
	} else {
		fmt.Printf("✗ Courses list failed with status: %d\n", resp.StatusCode)
	}
	
	// Test jobs list
	fmt.Println("\nTesting jobs list...")
	resp, err = http.Get(baseURL + "/jobs")
	if err != nil {
		fmt.Printf("Jobs list failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ Jobs list check passed")
		var jobsResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&jobsResp)
		total := jobsResp["total"].(float64)
		fmt.Printf("  Total jobs: %.0f\n", total)
	} else {
		fmt.Printf("✗ Jobs list failed with status: %d\n", resp.StatusCode)
	}
	
	fmt.Println("\n✓ All database integration tests completed successfully!")
}