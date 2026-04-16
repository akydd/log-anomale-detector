package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type payload struct {
	Type     string `json:"type"`
	Duration int    `json:"duration"`
}

var (
	ips = []string{
		"192.168.1.45", "10.0.0.112", "203.0.113.77",
		"198.51.100.22", "172.16.0.55", "10.0.0.201", "192.168.1.88",
	}
	users = []string{"-", "-", "-", "alice", "bob", "carol"}
	paths = []string{
		"/api/v1/orders", "/api/v1/products", "/api/v1/users",
		"/api/v1/payments", "/api/v1/reports", "/dashboard",
	}
	referers = []string{
		"-", "https://example.com/dashboard",
		"https://example.com/checkout", "https://example.com/",
	}
	agents = []string{
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36`,
		`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36`,
		`Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15`,
		`Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36`,
		`curl/7.88.1`,
	}
)

func pick(s []string) string { return s[rand.Intn(len(s))] }

func ts(t time.Time) string { return t.Format("02/Jan/2006:15:04:05 -0700") }

// generate500Spike writes 500 errors across different IPs and paths with
// 100-500ms gaps, simulating a backend crash or bad deploy.
func generate500Spike(dur time.Duration) int {
	deadline := time.Now().Add(dur)
	count := 0
	for time.Now().Before(deadline) {
		responseTimeMicros := rand.Intn(4_000_000) + 1_000_000 // 1s–5s
		fmt.Printf("%s - %s [%s] \"POST %s HTTP/1.1\" 500 %d \"%s\" \"%s\" %d\n",
			pick(ips), pick(users), ts(time.Now()), pick(paths),
			rand.Intn(200)+400, pick(referers), pick(agents), responseTimeMicros)
		count++
		time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)
	}
	return count
}

// generateBruteForce writes 401 responses from a single external IP hitting
// the login endpoint once per second, simulating credential stuffing.
func generateBruteForce(dur time.Duration) int {
	attackerIP := "45.33.32.156"
	deadline := time.Now().Add(dur)
	count := 0
	for time.Now().Before(deadline) {
		responseTimeMicros := rand.Intn(45_000) + 5_000 // 5ms–50ms
		fmt.Printf("%s - - [%s] \"POST /api/v1/login HTTP/1.1\" 401 64 \"-\" \"python-requests/2.31.0\" %d\n",
			attackerIP, ts(time.Now()), responseTimeMicros)
		count++
		time.Sleep(time.Second)
	}
	return count
}

// generateLatencyOutliers writes 200 responses with abnormally high response
// times (13-43s in the final microsecond field) every 500ms, simulating DB
// timeouts or resource exhaustion.
func generateLatencyOutliers(dur time.Duration) int {
	deadline := time.Now().Add(dur)
	count := 0
	for time.Now().Before(deadline) {
		responseTimeMicros := rand.Intn(30_000_000) + 13_000_000 // 13s–43s
		fmt.Printf("%s - %s [%s] \"GET %s HTTP/1.1\" 200 %d \"%s\" \"%s\" %d\n",
			pick(ips), pick(users), ts(time.Now()), pick(paths),
			rand.Intn(14000)+512, pick(referers), pick(agents), responseTimeMicros)
		count++
		time.Sleep(500 * time.Millisecond)
	}
	return count
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var p payload
	if err := json.Unmarshal([]byte(req.Body), &p); err != nil || p.Duration <= 0 || p.Type == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error":"body must be JSON with a positive \"duration\" (seconds) and a \"type\" of 500-spike, brute-force, or latency"}`,
		}, nil
	}
	if p.Duration > 120 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error":"duration cannot exceed 120 seconds"}`,
		}, nil
	}

	dur := time.Duration(p.Duration) * time.Second

	var fn func(time.Duration) int
	switch p.Type {
	case "500-spike":
		fn = generate500Spike
	case "brute-force":
		fn = generateBruteForce
	case "latency":
		fn = generateLatencyOutliers
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"error":"unknown type: %s"}`, p.Type),
		}, nil
	}

	count := fn(dur)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       fmt.Sprintf(`{"written":%d}`, count),
	}, nil
}

func main() {
	lambda.Start(handler)
}
