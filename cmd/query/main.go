package main

import (
	"akydd/log-anomale-detector/internal/adapters/out/dynamodb"
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

type db interface {
	Get(context.Context) ([]domain.ClassifiedLogs, error)
}

var tableName string
var d db

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		slog.Error("TABLE_NAME must be set")
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("loading AWS config", "error", err)
		os.Exit(1)
	}
	d = dynamodb.New(cfg, tableName)
}

func handler(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	results, err := d.Get(ctx)
	if err != nil {
		slog.Error("failed to Get dynamodb", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "internal server error",
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	body, err := json.Marshal(results)
	if err != nil {
		slog.Error("failed marshal data", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "internal server error",
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
