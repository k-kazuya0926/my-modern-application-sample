package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBクライアント
var dynamodbClient *dynamodb.Client

// シーケンステーブルの構造体
type SequenceItem struct {
	TableName string `json:"table_name"`
	Seq       int64  `json:"seq"`
}

// ユーザーテーブルの構造体
type UserItem struct {
	ID         int64   `json:"id" dynamodb:"id"`
	UserName   string  `json:"user_name" dynamodb:"user_name"`
	Email      string  `json:"email" dynamodb:"email"`
	AcceptedAt float64 `json:"accepted_at" dynamodb:"accepted_at"`
	Host       string  `json:"host" dynamodb:"host"`
}

// リクエストボディの構造体
type RequestBody struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

// 連番を更新して返す関数
func nextSeq(ctx context.Context, tableName string) (int64, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("my-modern-application-sample-prod-sequences"),
		Key: map[string]types.AttributeValue{
			"table_name": &types.AttributeValueMemberS{
				Value: tableName,
			},
		},
		UpdateExpression: aws.String("SET seq = seq + :val"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":val": &types.AttributeValueMemberN{
				Value: "1",
			},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	}

	result, err := dynamodbClient.UpdateItem(ctx, input)
	if err != nil {
		return 0, err
	}

	seqAttr := result.Attributes["seq"]
	seqValue, err := strconv.ParseInt(seqAttr.(*types.AttributeValueMemberN).Value, 10, 64)
	if err != nil {
		return 0, err
	}

	return seqValue, nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// エラーハンドリング用のdefer
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred: %v", r)
		}
	}()

	// シーケンスデータを得る
	nextSeq, err := nextSeq(ctx, "my-modern-application-sample-prod-users")
	if err != nil {
		log.Printf("Error getting next sequence: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "内部エラーが発生しました"}`,
		}, nil
	}

	// フォームに入力されたデータを得る
	var body string
	if request.IsBase64Encoded {
		decodedBytes, err := base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			log.Printf("Base64 decode error: %v", err)
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"error": "内部エラーが発生しました"}`,
			}, nil
		}
		body = string(decodedBytes)
	} else {
		body = request.Body
	}

	// JSONをパース
	var requestBody RequestBody
	if err := json.Unmarshal([]byte(body), &requestBody); err != nil {
		log.Printf("JSON parse error: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "内部エラーが発生しました"}`,
		}, nil
	}

	// クライアントのIPアドレスを得る
	host := request.RequestContext.HTTP.SourceIP

	// 現在のUNIXタイムスタンプを得る
	now := float64(time.Now().Unix())

	// userテーブルに登録する
	userItem := UserItem{
		ID:         nextSeq,
		UserName:   requestBody.UserName,
		Email:      requestBody.Email,
		AcceptedAt: now,
		Host:       host,
	}

	// DynamoDBアイテムに変換
	item, err := attributevalue.MarshalMap(userItem)
	if err != nil {
		log.Printf("Error marshaling user item: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "内部エラーが発生しました"}`,
		}, nil
	}

	// DynamoDBにアイテムを保存
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String("my-modern-application-sample-prod-users"),
		Item:      item,
	}

	_, err = dynamodbClient.PutItem(ctx, putInput)
	if err != nil {
		log.Printf("Error putting item to DynamoDB: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "内部エラーが発生しました"}`,
		}, nil
	}

	// 結果を返す
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{}`,
	}, nil
}

func main() {
	// DynamoDBクライアントを初期化
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamodbClient = dynamodb.NewFromConfig(cfg)

	lambda.Start(handler)
}
