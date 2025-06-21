package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// イベントの内容をJSONとしてログに出力
	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
	} else {
		log.Printf("Received event: %s", string(eventJSON))
	}

	// リクエストの基本情報をログ出力
	log.Printf("HTTP Method: %s", event.HTTPMethod)
	log.Printf("Path: %s", event.Path)
	log.Printf("Query Parameters: %v", event.QueryStringParameters)
	log.Printf("Headers: %v", event.Headers)

	// レスポンスを作成
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"message": "Hello tmp!", "status": "success"}`,
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
