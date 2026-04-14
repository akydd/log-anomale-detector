package dynamodb

import (
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsdb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Client struct {
	tableName    string
	dynamoClient *awsdb.Client
}

func New(config aws.Config, tableName string) *Client {
	c := awsdb.NewFromConfig(config)
	return &Client{
		tableName:    tableName,
		dynamoClient: c,
	}
}

type logResults struct {
	Timestamp          *time.Time `dynamodbav:"timestamp"`
	AnomalyType        *string    `dynamodbav:"anomaly"`
	Severity           *string    `dynamodbav:"severity"`
	RawEvidence        []string   `dynamodbav:"raw_evidence"`
	BedrockExplanation *string    `dynamodbav:"bedrock_explanation"`
}

func (c *Client) Put(ctx context.Context, logs domain.ClassifiedLogs) error {
	data := logResults{
		Timestamp:          logs.Timestamp,
		AnomalyType:        logs.AnomalyType,
		Severity:           logs.Severity,
		RawEvidence:        logs.RawEvidence,
		BedrockExplanation: logs.BedrockExplanation,
	}

	item, err := attributevalue.MarshalMap(data)
	if err != nil {
		return fmt.Errorf("marshalling map: %w", err)
	}

	_, err = c.dynamoClient.PutItem(ctx, &awsdb.PutItemInput{
		TableName: &c.tableName,
		Item:      item,
	})

	return err
}
