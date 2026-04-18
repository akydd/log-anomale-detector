package e2e_test

import (
	"context"
	"testing"
	"time"

	"akydd/log-anomale-detector/internal/domain"
)

func TestPut_Anomaly(t *testing.T) {
	anomalyType := "Auth attack"
	severity := "HIGH"

	item := domain.ClassifiedLogs{
		Flag:               true,
		Timestamp:          time.Now().UTC().Truncate(time.Millisecond),
		AnomalyType:        &anomalyType,
		Severity:           &severity,
		RawEvidence:        []string{`10.0.0.1 - - "POST /api/v1/login HTTP/1.1" 401 64`},
		BedrockExplanation: "Repeated 401s from a single IP indicate a brute force attack.",
	}

	if err := testClient.Put(context.Background(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPut_NilOptionalFields(t *testing.T) {
	item := domain.ClassifiedLogs{
		Flag:               true,
		Timestamp:          time.Now().UTC().Truncate(time.Millisecond),
		AnomalyType:        nil,
		Severity:           nil,
		RawEvidence:        []string{`192.168.1.45 - alice "GET /index.html HTTP/1.1" 200 1024`},
		BedrockExplanation: "No anomaly detected.",
	}

	if err := testClient.Put(context.Background(), item); err != nil {
		t.Fatalf("unexpected error writing item with nil optional fields: %v", err)
	}
}

func TestGet_ReturnsPutItem(t *testing.T) {
	anomalyType := "500 error spike"
	severity := "MEDIUM"
	ts := time.Now().UTC().Truncate(time.Millisecond)
	explanation := "Sustained 500 errors across multiple paths indicate a backend crash."

	item := domain.ClassifiedLogs{
		Flag:               true,
		Timestamp:          ts,
		AnomalyType:        &anomalyType,
		Severity:           &severity,
		RawEvidence:        []string{`192.168.1.45 - - "GET /api/v1/users HTTP/1.1" 500 512`},
		BedrockExplanation: explanation,
	}

	if err := testClient.Put(context.Background(), item); err != nil {
		t.Fatalf("putting item: %v", err)
	}

	results, err := testClient.Get(context.Background())
	if err != nil {
		t.Fatalf("getting items: %v", err)
	}

	var found bool
	for _, r := range results {
		if !r.Timestamp.Equal(ts) {
			continue
		}
		found = true
		if !r.Flag {
			t.Error("expected Flag to be true")
		}
		if r.AnomalyType == nil || *r.AnomalyType != anomalyType {
			t.Errorf("expected AnomalyType %q, got %v", anomalyType, r.AnomalyType)
		}
		if r.Severity == nil || *r.Severity != severity {
			t.Errorf("expected Severity %q, got %v", severity, r.Severity)
		}
		if r.BedrockExplanation != explanation {
			t.Errorf("expected BedrockExplanation %q, got %q", explanation, r.BedrockExplanation)
		}
		if len(r.RawEvidence) != len(item.RawEvidence) {
			t.Errorf("expected %d raw evidence entries, got %d", len(item.RawEvidence), len(r.RawEvidence))
		}
	}
	if !found {
		t.Error("put item not found in Get results")
	}
}
