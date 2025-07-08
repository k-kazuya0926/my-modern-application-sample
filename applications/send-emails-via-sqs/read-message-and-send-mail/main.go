package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var env string

func handler(ctx context.Context, event events.S3Event) error {
	// TODO: kobayashi

	return nil
}

func main() {
	env = os.Getenv("ENV")
	if env == "" {
		log.Fatalf("Environment variable ENV is required")
	}

	lambda.Start(handler)
}
