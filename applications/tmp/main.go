package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
)

// AppConfigDataクライアント（設定取得用）
var appconfigDataClient *appconfigdata.Client

// 環境変数
var (
	env               string
	applicationName   string
	environmentName   string
	configurationName string
)

// セッショントークン（グローバルで保持）
var sessionToken string

// 機能フラグの構造体
type FeatureFlags struct {
	Flag1 struct {
		Enabled bool `json:"enabled"`
	} `json:"flag1"`
}

// AppConfigセッションを開始する関数
func startConfigurationSession(ctx context.Context) error {
	input := &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:                &applicationName,
		EnvironmentIdentifier:                &environmentName,
		ConfigurationProfileIdentifier:       &configurationName,
		RequiredMinimumPollIntervalInSeconds: nil, // デフォルト値を使用
	}

	result, err := appconfigDataClient.StartConfigurationSession(ctx, input)
	if err != nil {
		return fmt.Errorf("AppConfigセッション開始エラー: %v", err)
	}

	sessionToken = *result.InitialConfigurationToken
	log.Printf("AppConfigセッション開始成功: token=%s", sessionToken[:10]+"...")
	return nil
}

// AppConfigから設定を取得する関数
func getFeatureFlags(ctx context.Context) (*FeatureFlags, error) {
	// セッションが開始されていない場合は開始
	if sessionToken == "" {
		if err := startConfigurationSession(ctx); err != nil {
			return nil, err
		}
	}

	// GetLatestConfigurationを使用して設定を取得
	input := &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: &sessionToken,
	}

	result, err := appconfigDataClient.GetLatestConfiguration(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("AppConfig設定取得エラー: %v", err)
	}

	// 次回用のトークンを更新
	if result.NextPollConfigurationToken != nil {
		sessionToken = *result.NextPollConfigurationToken
	}

	// レスポンスのコンテンツをパース
	var flags FeatureFlags
	if len(result.Configuration) > 0 {
		if err := json.Unmarshal(result.Configuration, &flags); err != nil {
			return nil, fmt.Errorf("機能フラグJSONパースエラー: %v", err)
		}
	} else {
		log.Printf("設定データが空です - 既存の設定を使用します")
	}

	return &flags, nil
}

// Lambda ハンドラー関数
func handler(ctx context.Context) (string, error) {
	log.Printf("tmpアプリケーション開始")

	// AppConfigから機能フラグを取得
	flags, err := getFeatureFlags(ctx)
	if err != nil {
		log.Printf("機能フラグ取得エラー: %v", err)
		return "Error getting feature flags", err
	}

	// flag1の状態をログに出力
	log.Printf("flag1の状態: enabled=%t", flags.Flag1.Enabled)

	// flag1が有効かどうかで処理を分岐
	if flags.Flag1.Enabled {
		log.Printf("flag1が有効です - 新機能を実行します")
		return "Hello world with flag1 enabled", nil
	} else {
		log.Printf("flag1が無効です - 従来の処理を実行します")
		return "Hello world with flag1 disabled", nil
	}
}

func main() {
	// 環境変数ENVを取得（必須）
	env = os.Getenv("ENV")
	if env == "" {
		log.Fatalf("Environment variable ENV is required")
	}

	// AppConfig関連の環境変数を設定
	applicationName = "tmp"
	environmentName = env
	configurationName = "tmp"

	// AWS設定をロード
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// AppConfigDataクライアントを初期化（設定取得用）
	appconfigDataClient = appconfigdata.NewFromConfig(cfg)

	log.Printf("tmpアプリケーション初期化完了 - AppConfig設定: app=%s, env=%s, config=%s",
		applicationName, environmentName, configurationName)

	lambda.Start(handler)
}
