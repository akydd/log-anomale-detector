package e2e_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"akydd/log-anomale-detector/internal/domain"
)

func skipIfNotDeployed(t *testing.T) (endpoint, apiKey string) {
	t.Helper()
	endpoint = os.Getenv("QUERY_ENDPOINT")
	apiKey = os.Getenv("API_KEY")
	if endpoint == "" || apiKey == "" {
		t.Skip("QUERY_ENDPOINT and API_KEY not set, skipping query e2e tests")
	}
	return endpoint, apiKey
}

func TestQuery_ReturnsOK(t *testing.T) {
	endpoint, apiKey := skipIfNotDeployed(t)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("calling query endpoint: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestQuery_ReturnsValidJSON(t *testing.T) {
	endpoint, apiKey := skipIfNotDeployed(t)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("calling query endpoint: %v", err)
	}
	defer res.Body.Close()

	var results []domain.ClassifiedLogs
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	for i, r := range results {
		if r.Timestamp.IsZero() {
			t.Errorf("result[%d]: timestamp is zero", i)
		}
	}
}

func TestQuery_RejectsRequestWithoutAPIKey(t *testing.T) {
	endpoint, _ := skipIfNotDeployed(t)

	res, err := http.Get(endpoint)
	if err != nil {
		t.Fatalf("calling query endpoint: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", res.StatusCode)
	}
}
