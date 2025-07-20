package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context) (string, error) {
	var v string
	return "tmp", nil
}

func main() {
	lambda.Start(handler)
}
