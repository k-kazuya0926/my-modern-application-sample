package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// リクエストの内容をログ出力
	log.Printf("HTTP Method: %s", request.RequestContext.HTTP.Method)
	log.Printf("Path: %s", request.RequestContext.HTTP.Path)
	log.Printf("Headers: %+v", request.Headers)
	log.Printf("Query String Parameters: %+v", request.QueryStringParameters)
	log.Printf("Path Parameters: %+v", request.PathParameters)
	log.Printf("Stage Variables: %+v", request.StageVariables)
	log.Printf("Request Context: %+v", request.RequestContext)
	log.Printf("Is Base64 Encoded: %t", request.IsBase64Encoded)
	log.Printf("Raw Path: %s", request.RawPath)
	log.Printf("Raw Query String: %s", request.RawQueryString)
	log.Printf("Route Key: %s", request.RouteKey)

	// Bodyの処理
	var body string
	if request.IsBase64Encoded {
		// Base64デコード
		decodedBytes, err := base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			log.Printf("Base64 decode error: %v", err)
			body = request.Body // デコードに失敗した場合は元のBodyを使用
		} else {
			body = string(decodedBytes)
			log.Printf("Decoded Body: %s", body)
		}
	} else {
		body = request.Body
		log.Printf("Body: %s", body)
	}

	// JSONとしてパースしてみる（オプション）
	if body != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
			log.Printf("Parsed JSON Body: %+v", jsonData)
		} else {
			log.Printf("Body is not valid JSON: %v", err)
		}
	}

	// レスポンスを返す
	response := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"message": "OK", "received": true}`,
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
