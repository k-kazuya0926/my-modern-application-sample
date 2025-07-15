package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context) (string, error) {
	return "create-purchase-history", nil
}

func main() {
	lambda.Start(handler)
}
