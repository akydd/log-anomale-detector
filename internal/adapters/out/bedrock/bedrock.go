package bedrock

import (
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const (
	modelID     = "anthropic.claude-3-haiku-20240307-v1:0"
	modelPrompt = `
Analyse these webserver logs, classifying activity as one of:
'Normal', '500 error spike', 'Auth attack', or 'Latency degradation'.
Format the response as an array of json objects with the fields:
flag - boolean, equal to true when log entries indicate anomalous activity, false otherwise
timestamp - starting timestamp of the anomalous activity, RFC3339
anomaly_type - one of '500 error spike', 'Auth attack', or 'Latency degradation'
severity - the severity of the anomaly, LOW, MEDIUM, HIGH
raw_evidence - contains the anomalous log entries
bedrock_explanation - explanation of why the logs were flagged as anomalous

Log entries that are 'Normal' should only return a response: [{"flag": false}].

A single batch of log entries may contain multiple anomalies.
The logs entries are:
`
)

type Client struct {
	c bedrockruntime.Client
}

func New(config aws.Config) *Client {
	awsClient := bedrockruntime.NewFromConfig(config)
	return &Client{
		c: *awsClient,
	}
}

func (c *Client) Classify(ctx context.Context, logs []string) ([]domain.ClassifiedLogs, error) {
	output, err := c.c.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelID),
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: modelPrompt + strings.Join(logs, ","),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("invoking bedrock: %w", err)
	}

	msg, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return nil, fmt.Errorf("could not convert response to text")
	}

	messageString := msg.Value.Content[0].(*types.ContentBlockMemberText).Value
	var results []domain.ClassifiedLogs
	err = json.Unmarshal([]byte(messageString), &results)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling model results: %w", err)
	}

	return results, nil
}
