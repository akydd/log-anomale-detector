package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
)

type Client struct {
	c        *awssns.Client
	topicARN string
}

func New(config aws.Config, topicID string) *Client {
	c := awssns.NewFromConfig(config)
	return &Client{
		c:        c,
		topicARN: topicID,
	}
}

func (c *Client) Notify(ctx context.Context, msg string) error {
	_, err := c.c.Publish(ctx, &awssns.PublishInput{
		TopicArn: aws.String(c.topicARN),
	})
	if err != nil {
		return fmt.Errorf("publishing to sns: %w", err)
	}

	return nil
}
