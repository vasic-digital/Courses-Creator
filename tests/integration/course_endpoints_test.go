package integration

import (
	"net/http"
	"testing"
)

func TestCourseEndpoints(t *testing.T) {
	const baseURL = "http://localhost:8080"

	endpoints := []string{
		"/courses/generate",
		"/api/courses/generate",
		"/api/v1/courses/generate",
		"/courses",
		"/api/courses",
		"/api/v1/courses",
	}

	for _, endpoint := range endpoints {
		t.Run("Endpoint_"+endpoint, func(t *testing.T) {
			url := baseURL + endpoint
			resp, err := http.Get(url)
			if err != nil {
				t.Logf("Error testing %s: %v", endpoint, err)
				return
			}
			defer resp.Body.Close()

			t.Logf("Endpoint %s: Status %d", endpoint, resp.StatusCode)
			if resp.StatusCode != 404 {
				t.Logf("✓ Endpoint found (status: %d)", resp.StatusCode)
			}
		})
	}
}
