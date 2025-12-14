package integration

import (
	"net/http"
	"testing"
)

func TestEndpoints(t *testing.T) {

	endpoints := []string{
		"/api/v1/health",
		"/health",
		"/api/health",
		"/",
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
			if resp.StatusCode == http.StatusOK {
				t.Log("✓ This endpoint works!")
			}
		})
	}
}
