package e2e_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"akydd/log-anomale-detector/internal/adapters/out/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsdb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	testTableName string
	testClient    *dynamodb.Client
)

func TestMain(m *testing.M) {
	if os.Getenv("AWS_ACCOUNT_ID") == "" {
		fmt.Println("AWS_ACCOUNT_ID not set, skipping e2e tests")
		os.Exit(0)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Printf("loading AWS config: %v\n", err)
		os.Exit(1)
	}

	testTableName = fmt.Sprintf("log-anomaly-detector-test-%d", time.Now().UnixMilli())
	raw := awsdb.NewFromConfig(cfg)

	_, err = raw.CreateTable(context.Background(), &awsdb.CreateTableInput{
		TableName:   aws.String(testTableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("timestamp"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("timestamp"), KeyType: types.KeyTypeHash},
		},
	})
	if err != nil {
		fmt.Printf("creating test table: %v\n", err)
		os.Exit(1)
	}

	waiter := awsdb.NewTableExistsWaiter(raw)
	if err = waiter.Wait(context.Background(), &awsdb.DescribeTableInput{
		TableName: aws.String(testTableName),
	}, 60*time.Second); err != nil {
		fmt.Printf("waiting for test table: %v\n", err)
		os.Exit(1)
	}

	testClient = dynamodb.New(cfg, testTableName)

	code := m.Run()

	raw.DeleteTable(context.Background(), &awsdb.DeleteTableInput{
		TableName: aws.String(testTableName),
	})

	os.Exit(code)
}
