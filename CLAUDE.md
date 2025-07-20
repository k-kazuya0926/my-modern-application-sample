# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

**重要**: コードベースに変更を加える際は、このCLAUDE.mdファイルの内容も適宜更新してください。新しいアプリケーションの追加、アーキテクチャパターンの変更、開発プロセスの変更などがあった場合は、このガイドラインを最新の状態に保つことで、今後の開発作業の効率性と一貫性を確保できます。

## 開発コマンド

### テストとローカル開発
- 個別のGoアプリケーションのテスト: `cd applications/<app-name> && go test ./...`
- Go mod tidyの実行: `cd applications/<app-name> && go mod tidy`
- ローカルビルド: `cd applications/<app-name> && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go`

### Dockerビルドコマンド
- Lambdaコンテナのビルド: `docker build -f applications/shared/lambda/Dockerfile --build-arg FUNCTION_NAME=<app-name> -t <image-name> applications/`
- コンテナのローカルテスト: `docker run --rm -p 9000:8080 <image-name>`

### CI/CDワークフロー
- 静的解析: `docker run --rm -v "$(pwd):$(pwd)" -w "$(pwd)" rhysd/actionlint:latest`
- セキュリティスキャン: TrivyスキャナーがmainブランチへのPR/pushで自動実行
- Lambdaビルド: 各アプリケーションはファイル変更時にトリガーされる専用ビルドワークフロー

## アーキテクチャ概要

### コア構造
これは**Go 1.24とAWS Lambdaで構築されたサーバーレスマイクロサービスアーキテクチャ**です。`applications/`ディレクトリ内の各アプリケーションは、独自の目的と依存関係を持つ独立したLambda関数です。

### 主要パターン

**Lambda関数構造:**
- 各アプリケーションは同じパターンに従います: `main.go`, `go.mod`, `go.sum`
- すべてLambdaランタイムに`github.com/aws/aws-lambda-go/lambda`を使用
- AWS SDK v2をアプリケーション全体で一貫して使用 (`github.com/aws/aws-sdk-go-v2/`)
- 接続再利用のための`main()`関数でのグローバルクライアント初期化
- 設定用の環境変数 (ENV、バケット名など)

**Dockerビルドパターン:**
- `applications/shared/lambda/Dockerfile`の共有Dockerfile
- マルチステージビルド: ビルド用golang:1.24-alpine、ランタイム用AWS Lambdaベースイメージ
- セキュリティフラグ付き静的リンク: `CGO_ENABLED=0`, `-ldflags='-w -s -extldflags "-static"'`
- ビルドするアプリケーションを指定する`FUNCTION_NAME`ビルド引数を使用

**CI/CDパターン:**
- 各Lambda関数に専用のGitHub Actionsワークフロー
- パスベーストリガー（関連ファイル変更時のみビルド）
- コンテナイメージストレージ用ECR
- セキュアな認証情報管理用AWS OIDC
- 再利用可能なビルド/デプロイロジック用コンポジットアクション

### アプリケーションカテゴリ

**シンプル関数:**
- `hello-world`: 基本的なLambdaハンドラーパターン
- `tmp`: 実験的関数

**AWSサービス統合:**
- `read-and-write-s3`: S3イベント処理、ZIP暗号化
- `register-user`: API Gateway + DynamoDB + S3 + SES統合
- `feature-flags`: AWS AppConfig統合
- `auth-by-cognito`: Cognito JWTトークン検証

**メッセージ処理:**
- `send-emails-via-sqs/`: SQSキューイング、SES送信、バウンス処理を含む完全なメールシステム
- `fan-out/`: 複数コンシューマーでのSNS/SQSファンアウトパターン

**オーケストレーション:**
- `saga-orchestration/`: 補償アクション付きStep Functions基盤の分散トランザクションパターン

### 主要実装パターン

**エラーハンドリング:**
- Lambdaハンドラーでのログ付きパニックリカバリ
- グレースフルデグラデーション（例：メール送信失敗がユーザー登録を中断しない）
- API Gateway用の構造化エラーレスポンス

**AWS SDK使用法:**
- コンテキスト対応操作付きV2 SDK
- グローバルに初期化されたサービス固有クライアント
- DynamoDB用の適切な属性値ハンドリング

**環境設定:**
- 環境固有リソース命名: `my-modern-application-sample-{env}-{resource}`
- 起動時に検証される必須環境変数

**セキュリティ:**
- 攻撃面を減らす静的バイナリコンパイル
- ハードコードされた認証情報なし
- セキュアなS3アクセス用の署名付きURL

## インフラストラクチャ統合

このコードベースは以下で管理されるインフラストラクチャと連携するよう設計されています: https://github.com/k-kazuya0926/my-modern-application-sample-infra

使用するAWSサービス:
- **API Gateway**: Webエンドポイント用HTTP API
- **Lambda**: コンテナイメージ付きサーバーレスコンピューティング
- **DynamoDB**: ID生成用シーケンステーブル付きNoSQLデータベース
- **S3**: イベントトリガーと署名付きURL付きオブジェクトストレージ
- **SES**: バウンス処理付きメール送信
- **SQS/SNS**: メッセージキューイングとパブ/サブ
- **Step Functions**: Sagaパターン用ワークフローオーケストレーション
- **AppConfig**: フィーチャーフラグ管理
- **Cognito**: ユーザー認証
- **ECR**: コンテナイメージレジストリ
- **X-Ray**: 分散トレーシング

## 開発ノート

### 新しいLambda関数の追加
1. `applications/`下に新しいディレクトリを作成
2. 既存パターンに従って`go.mod`, `go.sum`, `main.go`を追加
3. 命名規則に従った対応するGitHub Actionsワークフローを作成
4. 適切な`FUNCTION_NAME`ビルド引数で共有Dockerfileを使用

### 共通依存関係
- AWS Lambda Goランタイム: `github.com/aws/aws-lambda-go`
- AWS SDK v2: `github.com/aws/aws-sdk-go-v2/`
- 必要に応じてインポートされる特定サービス（dynamodb, s3, sesなど）

### 環境変数
すべてのLambda関数に必要:
- `ENV`: 環境名（prod, devなど）
- サービス固有変数（バケット名、テーブル名など）