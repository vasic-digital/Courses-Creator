package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func TestLLMIntegration(t *testing.T) {
	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Health check failed with status: %d", resp.StatusCode)
		}
		t.Log("✓ Health check passed")
	})

	// Test course generation with LLM enhancement
	t.Run("Course Generation", func(t *testing.T) {
		generateReq := map[string]interface{}{
			"markdown_path": "/tmp/test_course.md",
			"output_dir":    "/tmp/output",
			"options": map[string]interface{}{
				"quality":          "high", // Test high quality for LLM features
				"background_music": false,
				"languages":        []string{"en"},
			},
		}

		reqBody, _ := json.Marshal(generateReq)
		resp, err := http.Post(baseURL+"/api/v1/courses/generate", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Course generation failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("Course generation failed with status: %d", resp.StatusCode)
		}
		t.Log("✓ Course generation request accepted")

		// Parse response to get job ID
		var generateResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&generateResp)
		jobID, ok := generateResp["job_id"].(string)
		if !ok {
			t.Fatal("No job_id in response")
		}
		t.Logf("Job ID: %s", jobID)

		// Wait for processing and check job status
		t.Run("Job Status", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				time.Sleep(2 * time.Second)
				resp, err := http.Get(baseURL + "/api/v1/jobs/" + jobID)
				if err != nil {
					t.Logf("Job status check failed: %v", err)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					var jobResp map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&jobResp)
					status := jobResp["status"].(string)
					progress := int(jobResp["progress"].(float64))
					t.Logf("Job status: %s (Progress: %d%%)", status, progress)

					if status == "completed" {
						t.Log("✓ Job completed successfully!")
						return
					} else if status == "failed" {
						if errMsg, ok := jobResp["error"].(string); ok {
							t.Fatalf("Job failed: %s", errMsg)
						}
						t.Fatal("Job failed without error message")
					}
				}
			}
			t.Fatal("Job did not complete within timeout")
		})
	})

	// Test courses list
	t.Run("Courses List", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/courses")
		if err != nil {
			t.Fatalf("Courses list failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Courses list failed with status: %d", resp.StatusCode)
		}
		t.Log("✓ Courses list check passed")

		var coursesResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&coursesResp)
		total := coursesResp["total"].(float64)
		t.Logf("Total courses: %.0f", total)

		// Show course details if any
		if courses, ok := coursesResp["courses"].([]interface{}); ok && len(courses) > 0 {
			t.Log("Courses found:")
			for i, course := range courses {
				if courseMap, ok := course.(map[string]interface{}); ok {
					title := courseMap["title"].(string)
					description := courseMap["description"].(string)
					t.Logf("  %d. %s - %s", i+1, title, truncate(description, 50))
				}
			}
		}
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
