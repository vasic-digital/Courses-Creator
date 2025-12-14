package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestDatabaseIntegration(t *testing.T) {
	const baseURL = "http://localhost:8080/api/v1"

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Health check failed with status: %d", resp.StatusCode)
		}
		t.Log("✓ Health check passed")
	})

	// Test course generation
	t.Run("Course Generation", func(t *testing.T) {
		generateReq := map[string]interface{}{
			"markdown_path": "/tmp/test_course.md",
			"output_dir":    "/tmp/output",
			"options": map[string]interface{}{
				"quality":          "standard",
				"background_music": false,
				"languages":        []string{"en"},
			},
		}

		reqBody, _ := json.Marshal(generateReq)
		resp, err := http.Post(baseURL+"/courses/generate", "application/json", bytes.NewBuffer(reqBody))
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

		// Test job status
		t.Run("Job Status", func(t *testing.T) {
			time.Sleep(1 * time.Second) // Give some time for processing
			resp, err := http.Get(baseURL + "/jobs/" + jobID)
			if err != nil {
				t.Fatalf("Job status check failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Job status check failed with status: %d", resp.StatusCode)
			}
			t.Log("✓ Job status check passed")

			var jobResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&jobResp)
			t.Logf("Job status: %s", jobResp["status"])
		})
	})

	// Test courses list
	t.Run("Courses List", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/courses")
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
	})

	// Test jobs list
	t.Run("Jobs List", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/jobs")
		if err != nil {
			t.Fatalf("Jobs list failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Jobs list failed with status: %d", resp.StatusCode)
		}
		t.Log("✓ Jobs list check passed")

		var jobsResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&jobsResp)
		total := jobsResp["total"].(float64)
		t.Logf("Total jobs: %.0f", total)
	})
}
