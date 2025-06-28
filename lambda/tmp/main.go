package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

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
		Body: `{"message": "OK"}`,
	}

	return response, nil
}

// HTTPリクエストをAPIGatewayProxyRequestに変換
func httpToAPIGatewayEvent(r *http.Request) events.APIGatewayProxyRequest {
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// リクエストボディを読み取り
	var body string
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			body = string(bodyBytes)
		}
		r.Body.Close()
	}

	return events.APIGatewayProxyRequest{
		HTTPMethod:            r.Method,
		Path:                  r.URL.Path,
		QueryStringParameters: queryParams,
		Headers:               headers,
		Body:                  body,
	}
}

// HTTPハンドラー関数（Webサーバー用）
func httpHandler(w http.ResponseWriter, r *http.Request) {
	// HTTPリクエストをAPIGatewayProxyRequestに変換
	event := httpToAPIGatewayEvent(r)

	// Lambda handlerを実行
	response, err := handler(context.Background(), event)
	if err != nil {
		log.Printf("Handler error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// レスポンスヘッダーを設定
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// ステータスコードを設定
	w.WriteHeader(response.StatusCode)

	// レスポンスボディを書き込み
	w.Write([]byte(response.Body))
}

func main() {
	// 環境変数ENVをチェック
	if os.Getenv("ENV") == "LOCAL" {
		// Webサーバーとして起動
		http.HandleFunc("/", httpHandler)
		log.Fatal(http.ListenAndServe(":8080", nil))
	} else {
		lambda.Start(handler)
	}
}
