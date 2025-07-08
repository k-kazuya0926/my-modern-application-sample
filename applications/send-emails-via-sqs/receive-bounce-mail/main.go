package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBクライアント
var dynamoClient *dynamodb.Client

// 環境変数
var mailTable string

// SNSメッセージの構造体
type SNSMessage struct {
	NotificationType string `json:"notificationType"`
	Bounce           Bounce `json:"bounce"`
}

type Bounce struct {
	BouncedRecipients []BouncedRecipient `json:"bouncedRecipients"`
}

type BouncedRecipient struct {
	EmailAddress string `json:"emailAddress"`
}

// 初期化処理
func init() {
	// 環境変数の読み込み
	mailTable = os.Getenv("MAIL_TABLE")
	if mailTable == "" {
		log.Fatal("MAIL_TABLE environment variable is required")
	}

	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("AWS設定の読み込みに失敗しました:", err)
	}

	// DynamoDBクライアントの初期化
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

// Lambda ハンドラー関数
func handler(ctx context.Context, event events.SNSEvent) error {
	// SNSイベントの各レコードを処理
	for _, record := range event.Records {
		// SNSメッセージの取得
		message := record.SNS.Message

		// JSONメッセージをパース
		var snsMessage SNSMessage
		if err := json.Unmarshal([]byte(message), &snsMessage); err != nil {
			log.Printf("SNSメッセージのパースに失敗しました: %v", err)
			continue
		}

		// バウンス通知の場合のみ処理
		if snsMessage.NotificationType == "Bounce" {
			// バウンスした受信者の処理
			for _, bouncedRecipient := range snsMessage.Bounce.BouncedRecipients {
				email := bouncedRecipient.EmailAddress

				// DynamoDBのhas_errorフィールドを1に更新
				if err := updateErrorStatus(ctx, email); err != nil {
					log.Printf("メールアドレス %s のエラーステータス更新に失敗しました: %v", email, err)
				} else {
					log.Printf("メールアドレス %s のエラーステータスを更新しました", email)
				}
			}
		}
	}

	return nil
}

// DynamoDBのhas_errorフィールドを更新する関数
func updateErrorStatus(ctx context.Context, email string) error {
	// UpdateItemの実行
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(mailTable),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		UpdateExpression: aws.String("set has_error = :val"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":val": &types.AttributeValueMemberN{Value: "1"},
		},
	})

	return err
}

func main() {
	lambda.Start(handler)
}
