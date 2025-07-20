// S3にアップロードされたファイルをダウンロードし、パスワード付きZIPファイルに変換して別のS3バケットにアップロードするLambda関数

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/alexmullins/zip"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	// AWS設定を読み込み
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("AWS設定読み込みエラー: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// 出力バケット名を環境変数から取得
	outputBucket := os.Getenv("OUTPUT_BUCKET")
	if outputBucket == "" {
		return fmt.Errorf("OUTPUT_BUCKET環境変数が設定されていません")
	}

	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key

		fmt.Printf("処理中: バケット=%s, ファイル=%s\n", bucketName, objectKey)

		// S3からファイルをダウンロード
		result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    &objectKey,
		})
		if err != nil {
			return fmt.Errorf("ダウンロードエラー: %v", err)
		}
		defer func() {
			if err := result.Body.Close(); err != nil {
				log.Printf("failed to close response body: %v", err)
			}
		}()

		// ファイル内容を読み取り
		fileContent, err := io.ReadAll(result.Body)
		if err != nil {
			return fmt.Errorf("ファイル読み取りエラー: %v", err)
		}

		// パスワード付きZIPファイルを作成
		zipBuffer := new(bytes.Buffer)
		zipWriter := zip.NewWriter(zipBuffer)

		// パスワード付きファイルエントリを作成
		fileWriter, err := zipWriter.Encrypt(filepath.Base(objectKey), "mypassword")
		if err != nil {
			return fmt.Errorf("ZIP暗号化エラー: %v", err)
		}

		_, err = fileWriter.Write(fileContent)
		if err != nil {
			return fmt.Errorf("ZIPファイル書き込みエラー: %v", err)
		}

		err = zipWriter.Close()
		if err != nil {
			return fmt.Errorf("ZIPファイルクローズエラー: %v", err)
		}

		// S3にZIPファイルをアップロード
		zipKey := objectKey + ".zip"
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &outputBucket,
			Key:    &zipKey,
			Body:   bytes.NewReader(zipBuffer.Bytes()),
		})
		if err != nil {
			return fmt.Errorf("アップロードエラー: %v", err)
		}

		fmt.Printf("完了: %s を %s にアップロードしました\n", zipKey, outputBucket)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
