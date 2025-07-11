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
	AllFlags map[string]FeatureFlagDetails `json:"all_flags,omitempty"`
	Message  string                        `json:"message"`
}

// 機能フラグの詳細情報を含む構造体
type FeatureFlagDetails struct {
	Enabled bool `json:"enabled"`
	// Description    string                 `json:"description,omitempty"`
	// Value          interface{}            `json:"value,omitempty"`
	// ExpirationDate string                 `json:"expiration_date,omitempty"`
	// IsTemporary    bool                   `json:"is_temporary,omitempty"`
	// CreatedDate    string                 `json:"created_date,omitempty"`
	// ReviewDate     string                 `json:"review_date,omitempty"`
	// FlagType       string                 `json:"flag_type,omitempty"` // "temporary", "permanent", "experiment"
	// Metadata       map[string]interface{} `json:"metadata,omitempty"`
	// Rollout        *RolloutConfig         `json:"rollout,omitempty"`
}

// // ロールアウト設定構造体
// type RolloutConfig struct {
// 	Percentage int      `json:"percentage,omitempty"`
// 	UserGroups []string `json:"user_groups,omitempty"`
// 	Regions    []string `json:"regions,omitempty"`
// }

// AppConfig設定構造体
type FeatureFlags map[string]FeatureFlagDetails

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
func getFeatureFlags(ctx context.Context) (FeatureFlags, error) {
	// // セッショントークンがない場合は新しいセッションを開始
	// if configSession.Token == "" {
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
	// }

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
	return flags, nil
}

// Lambdaハンドラー関数
func handler(ctx context.Context) (Response, error) {
	log.Printf("feature-flags Lambda関数が開始されました")

	// AppConfigからfeature flagsを取得
	flags, err := getFeatureFlags(ctx)
	if err != nil {
		log.Printf("AppConfigからの設定取得に失敗: %v", err)
		return Response{
			AllFlags: nil,
			Message:  fmt.Sprintf("AppConfig取得エラー: %v", err),
		}, nil
	}

	return Response{
		AllFlags: flags,
		Message:  "AppConfigからflag1を正常に取得しました",
	}, nil
}

func main() {
	lambda.Start(handler)
}
