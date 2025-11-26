package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	// Test health endpoint
	fmt.Println("Testing health endpoint...")
	resp, err := http.Get(baseURL + "/api/v1/health")
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
	
	// Test course generation with LLM enhancement
	fmt.Println("\nTesting course generation with LLM integration...")
	generateReq := map[string]interface{}{
		"markdown_path": "/tmp/test_course.md",
		"output_dir":    "/tmp/output",
		"options": map[string]interface{}{
			"quality":           "high", // Test high quality for LLM features
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
		
		// Wait for processing and check job status
		fmt.Println("\nWaiting for job completion...")
		for i := 0; i < 10; i++ {
			time.Sleep(2 * time.Second)
			resp, err := http.Get(baseURL + "/jobs/" + jobID)
			if err != nil {
				fmt.Printf("Job status check failed: %v\n", err)
				continue
			}
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				var jobResp map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&jobResp)
				status := jobResp["status"].(string)
				progress := int(jobResp["progress"].(float64))
				fmt.Printf("  Job status: %s (Progress: %d%%)\n", status, progress)
				
				if status == "completed" {
					fmt.Println("✓ Job completed successfully!")
					break
				} else if status == "failed" {
					if errMsg, ok := jobResp["error"].(string); ok {
						fmt.Printf("✗ Job failed: %s\n", errMsg)
					}
					break
				}
			}
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
		
		// Show course details if any
		if courses, ok := coursesResp["courses"].([]interface{}); ok && len(courses) > 0 {
			fmt.Println("  Courses:")
			for i, course := range courses {
				if courseMap, ok := course.(map[string]interface{}); ok {
					title := courseMap["title"].(string)
					description := courseMap["description"].(string)
					fmt.Printf("    %d. %s - %s\n", i+1, title, description[:min(50, len(description))])
				}
			}
		}
	} else {
		fmt.Printf("✗ Courses list failed with status: %d\n", resp.StatusCode)
	}
	
	fmt.Println("\n✓ All LLM integration tests completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}