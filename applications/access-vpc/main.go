package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
)

var (
	rdsDataClient *rdsdata.Client
	clusterArn    string
	secretArn     string
	databaseName  string
)

func init() {
	ctx := context.Background()

	clusterArn = os.Getenv("CLUSTER_ARN")
	secretArn = os.Getenv("SECRET_ARN")
	databaseName = os.Getenv("DATABASE_NAME")

	if clusterArn == "" || secretArn == "" || databaseName == "" {
		panic("CLUSTER_ARN, SECRET_ARN, and DATABASE_NAME environment variables must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	rdsDataClient = rdsdata.NewFromConfig(cfg)
}

func handler(ctx context.Context) (string, error) {
	result, err := rdsDataClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(clusterArn),
		SecretArn:   aws.String(secretArn),
		Database:    aws.String(databaseName),
		Sql:         aws.String("SELECT 1"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute SELECT 1: %v", err)
	}

	if len(result.Records) > 0 && len(result.Records[0]) > 0 {
		field := result.Records[0][0]
		switch v := field.(type) {
		case *types.FieldMemberLongValue:
			return fmt.Sprintf("access-vpc: Successfully executed SELECT 1, result: %d", v.Value), nil
		case *types.FieldMemberStringValue:
			return fmt.Sprintf("access-vpc: Successfully executed SELECT 1, result: %s", v.Value), nil
		}
	}

	return "access-vpc: Successfully executed SELECT 1", nil
}

func main() {
	lambda.Start(handler)
}
