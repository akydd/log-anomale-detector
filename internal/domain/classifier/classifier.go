package classifier

import (
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"fmt"
)

type db interface {
	Put(context.Context, domain.ClassifiedLogs) error
}

type notifier interface {
	Notify(context.Context, string) error
}

type ai interface {
	Classify(context.Context, []string) ([]domain.ClassifiedLogs, error)
}

type Client struct {
	db       db
	ai       ai
	notifier notifier
}

func New(d db, n notifier, a ai) *Client {
	return &Client{
		db:       d,
		ai:       a,
		notifier: n,
	}
}

func (c *Client) Classify(ctx context.Context, logs []string) ([]domain.ClassifiedLogs, error) {
	results, err := c.ai.Classify(ctx, logs)
	if err != nil {
		return nil, fmt.Errorf("ai.Classify: %w", err)
	}

	for _, r := range results {
		if !r.Flag {
			continue
		}

		err = c.db.Put(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("writing to db: %w", err)
		}

		if r.Severity == "HIGH" {
			err = c.notifier.Notify(ctx, r.BedrockExplanation)
			if err != nil {
				return nil, fmt.Errorf("notifying: %w", err)
			}
		}
	}

	return results, nil
}
