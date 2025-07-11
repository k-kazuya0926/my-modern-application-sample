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
	Flag1Enabled bool   `json:"flag1_enabled"`
	Message      string `json:"message"`
}

// AppConfig設定構造体
type FeatureFlags struct {
	Flags map[string]struct {
		Name           string                 `json:"name"`
		Enabled        bool                   `json:"enabled"`
		Variants       map[string]interface{} `json:"variants"`
		DefaultVariant string                 `json:"defaultVariant"`
	} `json:"flags"`
	Values map[string]struct {
		Enabled bool `json:"enabled"`
	} `json:"values"`
	Version string `json:"version"`
}

// セッション情報を保持する構造体
type ConfigSession struct {
	Token string
}

var configSession *ConfigSession

// 初期化処理
func init() {
	// 環境変数の読み込み
	applicationID = os.Getenv("APPCONFIG_APPLICATION_ID")
	environmentID = os.Getenv("APPCONFIG_ENVIRONMENT_ID")
	configurationProfileID = os.Getenv("APPCONFIG_CONFIGURATION_PROFILE_ID")

	if applicationID == "" || environmentID == "" || configurationProfileID == "" {
		log.Printf("警告: AppConfig環境変数が設定されていません")
	}

	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("AWS設定の読み込みに失敗しました: %v", err)
	}

	// AppConfigDataクライアントの初期化
	appConfigDataClient = appconfigdata.NewFromConfig(cfg)

	// セッション情報の初期化
	configSession = &ConfigSession{}
}

// AppConfigからfeature flagsを取得
func getFeatureFlags(ctx context.Context) (*FeatureFlags, error) {
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

	// 設定データが空でない場合のみパース
	if len(configResp.Configuration) == 0 {
		log.Printf("設定データが更新されていません（既に最新の設定を取得済み）")
		return nil, fmt.Errorf("設定データが空です")
	}

	// JSONをパース
	var flags FeatureFlags
	if err := json.Unmarshal(configResp.Configuration, &flags); err != nil {
		return nil, fmt.Errorf("AppConfig設定のパースに失敗: %w", err)
	}

	log.Printf("AppConfigから設定を取得しました: %s", string(configResp.Configuration))
	return &flags, nil
}

// Lambdaハンドラー関数
func handler(ctx context.Context) (Response, error) {
	log.Printf("feature-flags Lambda関数が開始されました")

	// 環境変数が設定されていない場合のフォールバック
	if applicationID == "" || environmentID == "" || configurationProfileID == "" {
		log.Printf("AppConfig環境変数が未設定のため、デフォルト値を使用します")
		return Response{
			Flag1Enabled: true, // デフォルトでtrue
			Message:      "AppConfig未設定のため、デフォルト値を使用",
		}, nil
	}

	// AppConfigからfeature flagsを取得
	flags, err := getFeatureFlags(ctx)
	if err != nil {
		log.Printf("AppConfigからの設定取得に失敗: %v", err)
		// エラー時はデフォルト値を返す
		return Response{
			Flag1Enabled: false,
			Message:      fmt.Sprintf("AppConfig取得エラー: %v", err),
		}, nil
	}

	// flag1の状態を確認
	flag1Enabled := false
	if flag1Value, exists := flags.Values["flag1"]; exists {
		flag1Enabled = flag1Value.Enabled
	}

	log.Printf("flag1の状態: %v", flag1Enabled)

	return Response{
		Flag1Enabled: flag1Enabled,
		Message:      "AppConfigからflag1を正常に取得しました",
	}, nil
}

func main() {
	lambda.Start(handler)
}
