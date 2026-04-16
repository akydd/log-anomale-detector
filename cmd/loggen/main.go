package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	ips = []string{
		"192.168.1.45", "10.0.0.112", "203.0.113.77",
		"198.51.100.22", "172.16.0.55", "10.0.0.201",
	}
	users = []string{"-", "-", "-", "alice", "bob"}
	paths = []string{
		"/index.html", "/api/v1/users", "/static/css/main.css",
		"/blog/2026/03/new-features", "/images/logo.png",
		"/api/v1/users/42/profile", "/robots.txt",
	}
	notFoundPaths = []string{"/favicon.ico", "/old-page", "/wp-admin"}
	referers      = []string{
		"-", "https://example.com/dashboard",
		"https://example.com/index.html", "https://google.com/",
	}
	agents = []string{
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36`,
		`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36`,
		`Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15`,
		`Googlebot/2.1 (+http://www.google.com/bot.html)`,
		`curl/7.88.1`,
	}
)

func generateLog(t time.Time) string {
	ip := ips[rand.Intn(len(ips))]
	user := users[rand.Intn(len(users))]
	agent := agents[rand.Intn(len(agents))]
	referer := referers[rand.Intn(len(referers))]
	ts := t.Format("02/Jan/2006:15:04:05 -0700")

	var status int
	var path string
	var size int

	if rand.Intn(5) == 0 {
		status = 404
		path = notFoundPaths[rand.Intn(len(notFoundPaths))]
		size = rand.Intn(512) + 128
	} else {
		status = 200
		path = paths[rand.Intn(len(paths))]
		size = rand.Intn(30000) + 512
	}

	responseTimeMicros := rand.Intn(495_000) + 5_000 // 5ms–500ms
	return fmt.Sprintf(`%s - %s [%s] "GET %s HTTP/1.1" %d %d "%s" "%s" %d`,
		ip, user, ts, path, status, size, referer, agent, responseTimeMicros)
}

// handler outputs 10 realistic HTTP access log entries per invocation,
// with a ~20% chance of a 404 and ~80% chance of a 200.
func handler() {
	for range 10 {
		fmt.Println(generateLog(time.Now()))
	}
}

func main() {
	lambda.Start(handler)
}
