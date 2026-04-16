package bedrock_test

import (
	"context"
	"os"
	"testing"

	"akydd/log-anomale-detector/internal/adapters/out/bedrock"

	"github.com/aws/aws-sdk-go-v2/config"
)

func newClient(t *testing.T) *bedrock.Client {
	t.Helper()

	accountID := os.Getenv("AWS_ACCOUNT_ID")
	if accountID == "" {
		t.Skip("AWS_ACCOUNT_ID not set, skipping integration test")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("loading AWS config: %v", err)
	}

	return bedrock.New(cfg, accountID)
}

func TestClassify_NormalLogs(t *testing.T) {
	c := newClient(t)

	logs := []string{
		`192.168.1.45 - alice [10/Apr/2026:12:00:01 -0700] "GET /index.html HTTP/1.1" 200 1024 "-" "Mozilla/5.0"`,
		`192.168.1.46 - bob [10/Apr/2026:12:00:02 -0700] "GET /api/v1/users HTTP/1.1" 200 2048 "-" "Mozilla/5.0"`,
		`192.168.1.47 - - [10/Apr/2026:12:00:03 -0700] "GET /robots.txt HTTP/1.1" 200 128 "-" "Googlebot/2.1"`,
	}

	results, err := c.Classify(context.Background(), logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for _, r := range results {
		if r.Flag {
			t.Errorf("expected no anomalies in normal logs, got flag=true with anomaly_type=%v", r.AnomalyType)
		}
	}
}

func TestClassify_AuthAttack(t *testing.T) {
	c := newClient(t)

	logs := []string{
		`10.0.0.1 - admin [10/Apr/2026:12:00:01 -0700] "POST /login HTTP/1.1" 401 256 "-" "curl/7.88.1"`,
		`10.0.0.1 - admin [10/Apr/2026:12:00:02 -0700] "POST /login HTTP/1.1" 401 256 "-" "curl/7.88.1"`,
		`10.0.0.1 - admin [10/Apr/2026:12:00:03 -0700] "POST /login HTTP/1.1" 401 256 "-" "curl/7.88.1"`,
		`10.0.0.1 - admin [10/Apr/2026:12:00:04 -0700] "POST /login HTTP/1.1" 401 256 "-" "curl/7.88.1"`,
		`10.0.0.1 - admin [10/Apr/2026:12:00:05 -0700] "POST /login HTTP/1.1" 401 256 "-" "curl/7.88.1"`,
	}

	results, err := c.Classify(context.Background(), logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var flagged bool
	for _, r := range results {
		if r.Flag {
			flagged = true
			if r.AnomalyType == nil {
				t.Error("expected anomaly_type to be set")
			}
			if r.Severity == nil {
				t.Error("expected severity to be set")
			}
			if r.BedrockExplanation == "" {
				t.Error("expected bedrock_explanation to be set")
			}
		}
	}
	if !flagged {
		t.Error("expected anomaly to be detected in auth attack logs")
	}
}

func TestClassify_LatencyOutliers(t *testing.T) {
	c := newClient(t)

	logs := []string{
		`192.168.1.45 - alice [10/Apr/2026:12:00:01 -0700] "GET /api/v1/users HTTP/1.1" 200 1024 "-" "Mozilla/5.0" 15000000`,
		`192.168.1.46 - bob [10/Apr/2026:12:00:02 -0700] "GET /api/v1/orders HTTP/1.1" 200 2048 "-" "Mozilla/5.0" 22000000`,
		`192.168.1.47 - - [10/Apr/2026:12:00:03 -0700] "GET /api/v1/products HTTP/1.1" 200 512 "-" "Mozilla/5.0" 18000000`,
		`192.168.1.48 - alice [10/Apr/2026:12:00:04 -0700] "GET /dashboard HTTP/1.1" 200 4096 "-" "Mozilla/5.0" 35000000`,
		`192.168.1.49 - bob [10/Apr/2026:12:00:05 -0700] "GET /api/v1/reports HTTP/1.1" 200 8192 "-" "Mozilla/5.0" 29000000`,
	}

	results, err := c.Classify(context.Background(), logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var flagged bool
	for _, r := range results {
		if r.Flag {
			flagged = true
			if r.AnomalyType == nil {
				t.Error("expected anomaly_type to be set")
			}
			if r.Severity == nil {
				t.Error("expected severity to be set")
			}
			if r.BedrockExplanation == "" {
				t.Error("expected bedrock_explanation to be set")
			}
		}
	}
	if !flagged {
		t.Error("expected anomaly to be detected in latency outlier logs")
	}
}

func TestClassify_500ErrorSpike(t *testing.T) {
	c := newClient(t)

	logs := []string{
		`192.168.1.45 - - [10/Apr/2026:12:00:01 -0700] "GET /api/v1/users HTTP/1.1" 500 512 "-" "Mozilla/5.0"`,
		`192.168.1.46 - - [10/Apr/2026:12:00:02 -0700] "GET /api/v1/users HTTP/1.1" 500 512 "-" "Mozilla/5.0"`,
		`192.168.1.47 - - [10/Apr/2026:12:00:03 -0700] "GET /api/v1/users HTTP/1.1" 500 512 "-" "Mozilla/5.0"`,
		`192.168.1.48 - - [10/Apr/2026:12:00:04 -0700] "GET /api/v1/users HTTP/1.1" 500 512 "-" "Mozilla/5.0"`,
		`192.168.1.49 - - [10/Apr/2026:12:00:05 -0700] "GET /api/v1/users HTTP/1.1" 500 512 "-" "Mozilla/5.0"`,
	}

	results, err := c.Classify(context.Background(), logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var flagged bool
	for _, r := range results {
		if r.Flag {
			flagged = true
		}
	}
	if !flagged {
		t.Error("expected anomaly to be detected in 500 error spike logs")
	}
}
