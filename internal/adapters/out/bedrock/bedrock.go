package bedrock

import (
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brdocument "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const (
	inferenceProfile = "global.anthropic.claude-opus-4-5-20251101-v1:0"
	toolName         = "classify_logs"
	modelPrompt      = `
Analyse these webserver logs, classifying activity as one of:
'Normal', '500 error spike', 'Auth attack', or 'Latency degradation'.
Use the classify_logs tool to return the classification results.

A batch of log entries that are 'Normal' should only record the flag,
timestamp, raw_evidence, and bedrock_explanation. The flag should be false in this case.

A single batch of log entries may contain multiple anomalies.

The logs are in Combined Log Format with an additional response time field (microseconds):
host ident authuser [day/month/year:HH:MM:SS timezone] "method path protocol" status bytes "referer" "user-agent" response_time_microseconds

Normal response times are under 500ms (500,000 microseconds). Response times above 1 second (1,000,000 microseconds) are elevated. Response times above 10 seconds (10,000,000 microseconds) are latency outliers and should be classified as 'Latency degradation'.

The log entries are:
`
)

var toolConfig = &types.ToolConfiguration{
	Tools: []types.Tool{
		&types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				Name:        aws.String(toolName),
				Description: aws.String("Returns classification results for the provided log entries"),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: brdocument.NewLazyDocument(map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"results": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"flag":                map[string]interface{}{"type": "boolean"},
										"timestamp":           map[string]interface{}{"type": "string", "format": "date-time"},
										"anomaly_type":        map[string]interface{}{"type": "string", "enum": []string{"500 error spike", "Auth attack", "Latency degradation"}},
										"severity":            map[string]interface{}{"type": "string", "enum": []string{"LOW", "MEDIUM", "HIGH"}},
										"raw_evidence":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
										"bedrock_explanation": map[string]interface{}{"type": "string"},
									},
									"required": []string{"flag", "bedrock_explanation", "timestamp"},
								},
							},
						},
						"required": []string{"results"},
					}),
				},
			},
		},
	},
	ToolChoice: &types.ToolChoiceMemberTool{
		Value: types.SpecificToolChoice{
			Name: aws.String(toolName),
		},
	},
}

type Client struct {
	c         bedrockruntime.Client
	accountID string
}

func New(config aws.Config, accountID string) *Client {
	awsClient := bedrockruntime.NewFromConfig(config)
	return &Client{
		c:         *awsClient,
		accountID: accountID,
	}
}

func (c *Client) Classify(ctx context.Context, logs []string) ([]domain.ClassifiedLogs, error) {
	modelARN := fmt.Sprintf("arn:aws:bedrock:ca-west-1:%s:inference-profile/%s", c.accountID, inferenceProfile)

	output, err := c.c.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelARN),
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: modelPrompt + strings.Join(logs, "\n"),
					},
				},
			},
		},
		ToolConfig: toolConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("invoking bedrock: %w", err)
	}

	msg, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return nil, fmt.Errorf("unexpected output type from bedrock")
	}

	var toolUse *types.ToolUseBlock
	for _, block := range msg.Value.Content {
		if tu, ok := block.(*types.ContentBlockMemberToolUse); ok {
			toolUse = &tu.Value
			break
		}
	}
	if toolUse == nil {
		return nil, fmt.Errorf("no tool use block in bedrock response")
	}

	var raw interface{}
	if err := toolUse.Input.UnmarshalSmithyDocument(&raw); err != nil {
		return nil, fmt.Errorf("unmarshaling tool input: %w", err)
	}

	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling tool input: %w", err)
	}

	var wrapper struct {
		Results []domain.ClassifiedLogs `json:"results"`
	}
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshaling classified logs: %w", err)
	}

	return wrapper.Results, nil
}
