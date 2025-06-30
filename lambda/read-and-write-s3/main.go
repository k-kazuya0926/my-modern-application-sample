package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		// S3イベントの詳細を取得
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		eventName := record.EventName

		// ログ出力
		log.Printf("S3イベント受信: %s", eventName)
		log.Printf("バケット名: %s", bucketName)
		log.Printf("ファイル名: %s", objectKey)

		// 追加情報もログ出力
		fmt.Printf("リージョン: %s\n", record.AWSRegion)
		fmt.Printf("イベント時刻: %s\n", record.EventTime)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
