package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

// DynamoDBクライアント
var dynamodbClient *dynamodb.Client

// S3クライアント
var s3Client *s3.Client

// SESクライアント
var sesClient *ses.Client

// 環境変数
var (
	env            string
	contentsBucket string
	fileName       string
	mailFrom       string
)

// シーケンステーブルの構造体
type SequenceItem struct {
	TableName string `json:"table_name"`
	Seq       int64  `json:"seq"`
}

// リクエストボディの構造体
type RequestBody struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

// 連番を更新して返す関数
func nextSeq(ctx context.Context, tableName string) (int64, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(fmt.Sprintf("my-modern-application-sample-%s-sequences", env)),
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

// メール送信関数
func sendmail(ctx context.Context, to, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: aws.String(mailFrom),
		ReplyToAddresses: []string{
			mailFrom,
		},
		Destination: &sestypes.Destination{
			ToAddresses: []string{
				to,
			},
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
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred: %v", r)
		}
	}()

	// シーケンスデータを得る
	nextSeq, err := nextSeq(ctx, fmt.Sprintf("my-modern-application-sample-%s-users", env))
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

	// 署名付きURLを生成（GET用）
	presignClient := s3.NewPresignClient(s3Client)
	presignRequest, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(contentsBucket),
		Key:    aws.String(fileName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(8) * time.Hour
	})
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "内部エラーが発生しました"}`,
		}, nil
	}

	// DynamoDBアイテムを手動で作成
	// 注意: attributevalue.MarshalMap()はdynamodbタグを正しく認識しないため、手動で作成
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberN{
			Value: strconv.FormatInt(nextSeq, 10),
		},
		"user_name": &types.AttributeValueMemberS{
			Value: requestBody.UserName,
		},
		"email": &types.AttributeValueMemberS{
			Value: requestBody.Email,
		},
		"accepted_at": &types.AttributeValueMemberN{
			Value: strconv.FormatFloat(now, 'f', -1, 64),
		},
		"host": &types.AttributeValueMemberS{
			Value: host,
		},
		"url": &types.AttributeValueMemberS{
			Value: presignRequest.URL,
		},
	}

	// DynamoDBにアイテムを保存
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(fmt.Sprintf("my-modern-application-sample-%s-users", env)),
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

	// 登録完了メールを送信
	mailbody := fmt.Sprintf(`%s様
ご登録ありがとうございました。
下記のURLからダウンロードできます。
%s`, requestBody.UserName, presignRequest.URL)

	err = sendmail(ctx, requestBody.Email, "登録ありがとうございました", mailbody)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		// メール送信エラーでもAPIは成功として返す（ユーザー登録は完了しているため）
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
	// 環境変数ENVを取得（必須）
	env = os.Getenv("ENV")
	if env == "" {
		log.Fatalf("Environment variable ENV is required")
	}

	// 環境変数を取得
	contentsBucket = os.Getenv("CONTENTS_BUCKET")
	if contentsBucket == "" {
		log.Fatalf("Environment variable CONTENTS_BUCKET is required")
	}

	fileName = os.Getenv("FILE_NAME")
	if fileName == "" {
		log.Fatalf("Environment variable FILE_NAME is required")
	}

	mailFrom = os.Getenv("MAIL_FROM")
	if mailFrom == "" {
		log.Fatalf("Environment variable MAIL_FROM is required")
	}

	// AWS設定をロード
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// DynamoDBクライアントを初期化
	dynamodbClient = dynamodb.NewFromConfig(cfg)

	// S3クライアントを初期化
	s3Client = s3.NewFromConfig(cfg)

	// SESクライアントを初期化
	sesClient = ses.NewFromConfig(cfg)

	lambda.Start(handler)
}
