package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var env string

type MailAddress struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	HasError int    `json:"has_error"`
	IsSent   int    `json:"is_sent"`
}

func handler(ctx context.Context, event events.S3Event) error {
	// AWS設定を読み込み
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	// ①DynamoDBのmail-addressesテーブルを操作するオブジェクト
	dynamoClient := dynamodb.NewFromConfig(cfg)
	tableName := fmt.Sprintf("my-modern-application-sample-%s-mail-addresses", env)

	// ②SQSのキューを操作するオブジェクト
	sqsClient := sqs.NewFromConfig(cfg)
	queueName := fmt.Sprintf("my-modern-application-sample-%s-send-mail", env)

	// キューのURLを取得
	queueResult, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return fmt.Errorf("failed to get queue URL: %v", err)
	}
	queueURL := queueResult.QueueUrl

	for _, record := range event.Records {
		// ③S3に置かれたファイルパスを取得
		bucketName := record.S3.Bucket.Name
		fileName := record.S3.Object.Key

		// ④has_errorが0のものをmail-addressesテーブルから取得
		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(tableName),
			IndexName:              aws.String("has_error-index"),
			KeyConditionExpression: aws.String("has_error = :has_error"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":has_error": &types.AttributeValueMemberN{
					Value: "0",
				},
			},
		}

		result, err := dynamoClient.Query(ctx, queryInput)
		if err != nil {
			return fmt.Errorf("failed to query DynamoDB: %v", err)
		}

		// ⑤上記の1件1件についてループ処理
		for _, item := range result.Items {
			email := ""
			userName := ""

			// emailの値を取得
			if emailAttr, ok := item["email"]; ok {
				if emailVal, ok := emailAttr.(*types.AttributeValueMemberS); ok {
					email = emailVal.Value
				}
			}

			// user_nameの値を取得
			if userNameAttr, ok := item["user_name"]; ok {
				if userNameVal, ok := userNameAttr.(*types.AttributeValueMemberS); ok {
					userName = userNameVal.Value
				}
			}

			// ⑥送信済みを示すis_sentを0にする
			updateInput := &dynamodb.UpdateItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"email": &types.AttributeValueMemberS{
						Value: email,
					},
				},
				UpdateExpression: aws.String("set is_sent = :val"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":val": &types.AttributeValueMemberN{
						Value: "0",
					},
				},
			}

			_, err := dynamoClient.UpdateItem(ctx, updateInput)
			if err != nil {
				return fmt.Errorf("failed to update DynamoDB item: %v", err)
			}

			// ⑦SQSにメッセージとして登録する
			sendMessageInput := &sqs.SendMessageInput{
				QueueUrl:    queueURL,
				MessageBody: aws.String(email),
				MessageAttributes: map[string]sqsTypes.MessageAttributeValue{
					"user_name": {
						DataType:    aws.String("String"),
						StringValue: aws.String(userName),
					},
					"bucket_name": {
						DataType:    aws.String("String"),
						StringValue: aws.String(bucketName),
					},
					"file_name": {
						DataType:    aws.String("String"),
						StringValue: aws.String(fileName),
					},
				},
			}

			sqsResponse, err := sqsClient.SendMessage(ctx, sendMessageInput)
			if err != nil {
				return fmt.Errorf("failed to send SQS message: %v", err)
			}

			// 結果をログに出力しておく
			responseJSON, _ := json.Marshal(sqsResponse)
			log.Printf("SQS Response: %s", string(responseJSON))
		}
	}

	return nil
}

func main() {
	env = os.Getenv("ENV")
	if env == "" {
		log.Fatalf("Environment variable ENV is required")
	}

	lambda.Start(handler)
}
