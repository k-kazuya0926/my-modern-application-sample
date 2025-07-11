package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
)

// AppConfigDataクライアント（グローバル変数として保持）
var appConfigDataClient *appconfigdata.Client

// 環境変数
var (
	applicationID          string
	environmentID          string
	configurationProfileID string
)

// レスポンス構造体
type Response struct {
	StatusCode int          `json:"statusCode"`
	Body       ResponseBody `json:"body"`
}

type ResponseBody struct {
	AllFlags map[string]FeatureFlagDetails `json:"all_flags,omitempty"`
	Message  string                        `json:"message"`
	Error    string                        `json:"error,omitempty"`
}

// 機能フラグの詳細情報を含む構造体
type FeatureFlagDetails struct {
	Enabled    bool   `json:"enabled"`
	Attribute1 string `json:"attribute1,omitempty"`
}

// AppConfig設定構造体
type FeatureFlags map[string]FeatureFlagDetails

// セッション情報を保持する構造体
type ConfigSession struct {
	Token       string
	CachedFlags FeatureFlags // キャッシュされた設定値
}

var configSession *ConfigSession

// 初期化処理
func init() {
	// パニックリカバリーの実装
	defer func() {
		if r := recover(); r != nil {
			log.Printf("初期化中にパニックが発生しました: %v", r)
			panic(r) // 初期化時のパニックは再度パニックさせる
		}
	}()

	// 環境変数の読み込みと厳密なバリデーション
	applicationID = os.Getenv("APPCONFIG_APPLICATION_ID")
	environmentID = os.Getenv("APPCONFIG_ENVIRONMENT_ID")
	configurationProfileID = os.Getenv("APPCONFIG_CONFIGURATION_PROFILE_ID")

	// 必須環境変数のバリデーション
	if applicationID == "" {
		log.Fatalf("必須環境変数が未設定です: APPCONFIG_APPLICATION_ID")
	}
	if environmentID == "" {
		log.Fatalf("必須環境変数が未設定です: APPCONFIG_ENVIRONMENT_ID")
	}
	if configurationProfileID == "" {
		log.Fatalf("必須環境変数が未設定です: APPCONFIG_CONFIGURATION_PROFILE_ID")
	}

	log.Printf("環境変数を読み込みました")

	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("AWS設定の読み込みに失敗しました: %v", err)
	}

	// AppConfigDataクライアントの初期化
	appConfigDataClient = appconfigdata.NewFromConfig(cfg)

	// セッション情報の初期化
	configSession = &ConfigSession{}

	log.Printf("feature-flags Lambda関数の初期化が完了しました")
}

// AppConfigからfeature flagsを取得
func getFeatureFlags(ctx context.Context) (FeatureFlags, error) {
	// コンテキストタイムアウトの設定（30秒）
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// セッショントークンがない場合は新しいセッションを開始
	if configSession.Token == "" {
		startSessionInput := &appconfigdata.StartConfigurationSessionInput{
			ApplicationIdentifier:          &applicationID,
			EnvironmentIdentifier:          &environmentID,
			ConfigurationProfileIdentifier: &configurationProfileID,
		}

		sessionResp, err := appConfigDataClient.StartConfigurationSession(ctx, startSessionInput)
		if err != nil {
			return nil, fmt.Errorf("AppConfigセッション開始に失敗: %w", err)
		}

		configSession.Token = *sessionResp.InitialConfigurationToken
		log.Printf("新しいAppConfigセッションを開始しました")
	}

	// 最新の設定データを取得
	getConfigInput := &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: &configSession.Token,
	}

	configResp, err := appConfigDataClient.GetLatestConfiguration(ctx, getConfigInput)
	if err != nil {
		return nil, fmt.Errorf("AppConfig設定取得に失敗: %w", err)
	}

	// 次回用のトークンを保存
	configSession.Token = *configResp.NextPollConfigurationToken

	// 設定データが空の場合（設定に変更がない場合）
	if len(configResp.Configuration) == 0 {
		log.Printf("設定データが更新されていません（既に最新の設定を取得済み）")

		// キャッシュされた設定がある場合はそれを返す
		if configSession.CachedFlags != nil {
			log.Printf("キャッシュされた設定を返します")
			return configSession.CachedFlags, nil
		}

		// 初回取得でキャッシュもない場合はエラー
		return nil, fmt.Errorf("初回設定取得に失敗: 設定データが空です")
	}

	// JSONをパース
	var flags FeatureFlags
	if err := json.Unmarshal(configResp.Configuration, &flags); err != nil {
		return nil, fmt.Errorf("AppConfig設定のパースに失敗: %w", err)
	}

	// 正常に取得できた設定をキャッシュに保存
	configSession.CachedFlags = flags
	log.Printf("AppConfigから設定を取得してキャッシュに保存しました, データ: %s", string(configResp.Configuration))

	return flags, nil
}

// Lambdaハンドラー関数
func handler(ctx context.Context) (Response, error) {
	// パニックリカバリーの実装
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ハンドラー実行中にパニックが発生しました: %v", r)
			// パニック時は500エラーを返す
		}
	}()

	log.Printf("feature-flags Lambda関数が開始されました")

	// AppConfigからfeature flagsを取得
	flags, err := getFeatureFlags(ctx)
	if err != nil {
		log.Printf("AppConfigからの設定取得に失敗: %v", err)
		return Response{
			StatusCode: 500,
			Body: ResponseBody{
				AllFlags: nil,
				Message:  "内部サーバーエラーが発生しました",
				Error:    "AppConfig取得エラー", // 内部エラーの詳細は外部に露出しない
			},
		}, nil
	}

	log.Printf("AppConfigから設定を正常に取得しました, フラグ数: %d", len(flags))

	return Response{
		StatusCode: 200,
		Body: ResponseBody{
			AllFlags: flags,
			Message:  "AppConfigから設定を正常に取得しました",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
