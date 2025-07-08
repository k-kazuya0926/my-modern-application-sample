package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	env          string
	mailFrom     string
	tableName    string
	s3Client     *s3.Client
	dynamoClient *dynamodb.Client
	sesClient    *ses.Client
)

// handler SQSイベントを処理してメール送信を行う
func handler(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		err := processMessage(ctx, record)
		if err != nil {
			log.Printf("メッセージ処理中にエラーが発生しました: %v", err)
			return err
		}
	}
	return nil
}

// processMessage 個別のSQSメッセージを処理する
func processMessage(ctx context.Context, record events.SQSMessage) error {
	email := record.Body

	// メッセージ属性から必要な情報を取得
	bucketName := record.MessageAttributes["bucket_name"].StringValue
	fileName := record.MessageAttributes["file_name"].StringValue
	userName := record.MessageAttributes["user_name"].StringValue

	if bucketName == nil || fileName == nil || userName == nil {
		return fmt.Errorf("必要なメッセージ属性が不足しています")
	}

	log.Printf("メール処理開始: email=%s, bucket=%s, file=%s, user=%s",
		email, *bucketName, *fileName, *userName)

	// S3バケットからメール本文を取得
	mailData, err := getMailDataFromS3(ctx, *bucketName, *fileName)
	if err != nil {
		return fmt.Errorf("S3からメールデータ取得エラー: %w", err)
	}

	// メールデータを解析
	subject, body, err := parseMailData(mailData)
	if err != nil {
		return fmt.Errorf("メールデータ解析エラー: %w", err)
	}

	// DynamoDBで送信状態を確認・更新
	wasSent, err := updateSendStatus(ctx, email)
	if err != nil {
		return fmt.Errorf("DynamoDB更新エラー: %w", err)
	}

	// 未送信の場合のみメール送信
	if !wasSent {
		err = sendEmail(ctx, email, subject, body)
		if err != nil {
			return fmt.Errorf("メール送信エラー: %w", err)
		}
		log.Printf("メール送信完了: %s", email)
	} else {
		log.Printf("再送信スキップ: %s", email)
	}

	return nil
}

// getMailDataFromS3 S3からメールデータを取得する
func getMailDataFromS3(ctx context.Context, bucketName, fileName string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	}

	result, err := s3Client.GetObject(ctx, input)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	// レスポンスボディを安全に読み取り
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// parseMailData メールデータを解析してsubjectとbodyを取得する
func parseMailData(mailData string) (string, string, error) {
	lines := strings.Split(mailData, "\n")
	if len(lines) < 3 {
		return "", "", fmt.Errorf("メールデータの形式が不正です")
	}

	subject := lines[0]
	body := strings.Join(lines[2:], "\n")

	return subject, body, nil
}

// updateSendStatus DynamoDBで送信状態を確認・更新する
func updateSendStatus(ctx context.Context, email string) (bool, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		UpdateExpression: aws.String("set is_sent = :val"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":val": &types.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: types.ReturnValueUpdatedOld,
	}

	result, err := dynamoClient.UpdateItem(ctx, input)
	if err != nil {
		return false, err
	}

	// 前回の値を確認
	if oldValue, exists := result.Attributes["is_sent"]; exists {
		if numValue, ok := oldValue.(*types.AttributeValueMemberN); ok {
			oldIsSent, err := strconv.Atoi(numValue.Value)
			if err != nil {
				return false, err
			}
			return oldIsSent == 1, nil
		}
	}

	// 属性が存在しない場合は未送信とみなす
	return false, nil
}

// sendEmail SESを使用してメールを送信する
func sendEmail(ctx context.Context, toEmail, subject, body string) error {
	input := &ses.SendEmailInput{
		Source:           aws.String(mailFrom),
		ReplyToAddresses: []string{mailFrom},
		Destination: &sestypes.Destination{
			ToAddresses: []string{toEmail},
		},
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &sestypes.Body{
				Text: &sestypes.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	_, err := sesClient.SendEmail(ctx, input)
	return err
}

func main() {
	// 環境変数の読み込み
	env = os.Getenv("ENV")
	if env == "" {
		log.Fatalf("Environment variable ENV is required")
	}

	mailFrom = os.Getenv("MAIL_FROM")
	if mailFrom == "" {
		log.Fatalf("Environment variable MAIL_FROM is required")
	}

	tableName = fmt.Sprintf("my-modern-application-sample-%s-mail-addresses", env)

	// AWS設定の初期化
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("AWS設定の読み込みに失敗しました: %v", err)
	}

	// AWSクライアントの初期化
	s3Client = s3.NewFromConfig(cfg)
	dynamoClient = dynamodb.NewFromConfig(cfg)
	sesClient = ses.NewFromConfig(cfg)

	lambda.Start(handler)
}
