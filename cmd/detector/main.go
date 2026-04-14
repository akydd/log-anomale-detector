package main

import (
	"akydd/log-anomale-detector/internal/adapters/out/bedrock"
	"akydd/log-anomale-detector/internal/adapters/out/dynamodb"
	"akydd/log-anomale-detector/internal/adapters/out/sns"
	"akydd/log-anomale-detector/internal/domain/classifier"
	"context"

	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/config"
)

var snsClient *sns.Client
var dynamoClient *dynamodb.Client
var bedrockClient *bedrock.Client
var snsTopicARN string
var tableName string
var c *classifier.Client

// Clients are initialized here so that warm Lambda invocations can reuse it.
func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("loading AWS config: %v", err)
	}
	dynamoClient = dynamodb.New(cfg, tableName)
	bedrockClient = bedrock.New(cfg)
	snsClient = sns.New(cfg, snsTopicARN)

	c = classifier.New(dynamoClient, snsClient, bedrockClient)
}

func handler(ctx context.Context, event events.CloudwatchLogsEvent) error {
	data, err := event.AWSLogs.Parse()
	if err != nil {
		return fmt.Errorf("parsing log event: %w", err)
	}

	var logs []string
	for _, e := range data.LogEvents {
		logs = append(logs, e.Message)
	}

	results, err := c.Classify(ctx, logs)
	if err != nil {
		return fmt.Errorf("classifyLogEvents: %w", err)
	}

	fmt.Println(results)
	return nil
}

func main() {
	lambda.Start(handler)
}
