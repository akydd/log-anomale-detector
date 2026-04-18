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

type logResult struct {
	Timestamp          time.Time `dynamodbav:"timestamp"`
	AnomalyType        *string   `dynamodbav:"anomaly"`
	Severity           *string   `dynamodbav:"severity"`
	RawEvidence        []string  `dynamodbav:"raw_evidence"`
	BedrockExplanation string    `dynamodbav:"bedrock_explanation"`
}

func (c *Client) Get(ctx context.Context) ([]domain.ClassifiedLogs, error) {
	var items []logResult
	var response *awsdb.ScanOutput
	var err error

	scanPaginator := awsdb.NewScanPaginator(c.dynamoClient, &awsdb.ScanInput{
		TableName: aws.String(c.tableName),
	})

	for scanPaginator.HasMorePages() {
		response, err = scanPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scanning table: %w", err)
		}

		var page []logResult
		err = attributevalue.UnmarshalListOfMaps(response.Items, &page)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling data: %w", err)
		}

		items = append(items, page...)
	}

	// Convert to domain model
	var results []domain.ClassifiedLogs
	for _, item := range items {
		j := domain.ClassifiedLogs{
			Flag:               true,
			Timestamp:          item.Timestamp,
			AnomalyType:        item.AnomalyType,
			Severity:           item.Severity,
			RawEvidence:        item.RawEvidence,
			BedrockExplanation: item.BedrockExplanation,
		}
		results = append(results, j)
	}
	return results, nil
}

func (c *Client) Put(ctx context.Context, logs domain.ClassifiedLogs) error {
	data := logResult{
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
