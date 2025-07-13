# my-modern-application-sample

## 技術スタック

- **言語**: Go 1.24
- **コンテナ**: Docker
- **アーキテクチャ**: サーバーレス
- **インフラ**: AWS (Lambda, S3, API Gateway, DynamoDB, SES, SQS, SNS, X-Ray, AppConfig, Cognito)
  - https://github.com/k-kazuya0926/my-modern-application-sample-infra
- **CI/CD**: GitHub Actions
- **セキュリティスキャン**: Trivy

## アプリケーション一覧

### 1. hello-world
**概要**: 基本的なLambda関数のサンプル\
**機能**: "Hello world" メッセージを返すシンプルなLambda関数\
**技術スタック**:
- Go
- Lambda

### 2. read-and-write-s3
**概要**: S3ファイル処理とZIP暗号化\
**機能**:
- S3イベントトリガーでファイルアップロードを検知
- アップロードされたファイルをダウンロード
- パスワード付きZIPファイルに変換
- 別のS3バケットにアップロード

**技術スタック**:
- Go
- Lambda
- S3(イベントトリガー・ファイル操作)
- alexmullins/zip(パスワード付きZIP作成)

### 3. register-user
**概要**: ユーザー登録とメール送信機能\
**機能**:
- API Gateway経由でユーザー情報(名前・メールアドレス)を受信
- DynamoDBにユーザー情報を保存(連番ID自動生成)
- S3署名付きURLを生成
- SES経由で登録完了メールを送信

**技術スタック**:
- Go
- Lambda
- API Gateway(HTTP API)
- DynamoDB(ユーザーテーブル・シーケンステーブル)
- S3(署名付きURL生成)
- SES(メール送信)

### 4. send-emails-via-sqs(メール配信システム)

#### 4.1 send-message
**概要**: メール送信キューへの登録\
**機能**:
- S3イベントトリガーで処理開始
- DynamoDBからエラーのないメールアドレスを取得
- SQSキューにメール送信メッセージを登録
- 送信ステータスを未送信に更新

**技術スタック**:
- Go
- Lambda
- S3(イベントトリガー)
- DynamoDB(メールアドレステーブル)
- SQS(メッセージキュー)
- X-Ray

#### 4.2 read-message-and-send-mail
**概要**: SQSメッセージ処理とメール送信\
**機能**:
- SQSメッセージを受信・処理
- S3からメール本文テンプレートを取得
- 重複送信チェック(DynamoDB)
- SES経由でメール送信

**技術スタック**:
- Go
- Lambda
- SQS(メッセージ受信)
- S3(メールテンプレート取得)
- DynamoDB(送信状態管理)
- SES(メール送信)
- X-Ray

#### 4.3 receive-bounce-mail
**概要**: バウンスメール処理\
**機能**:
- SNS経由でSESバウンス通知を受信
- バウンスしたメールアドレスのエラーフラグを更新
- 今後の送信対象から除外

**技術スタック**:
- Go
- Lambda
- SNS(バウンス通知受信)
- DynamoDB(エラーステータス更新)
- X-Ray

### 5. feature-flags
**概要**: AWS AppConfigを使用した機能フラグ管理システム\
**機能**:
- AWS AppConfigから機能フラグ設定を取得
- セッション管理とキャッシュ機能
- 設定データのJSONパース
- 機能フラグの詳細情報(有効/無効状態、属性値)を返却

**技術スタック**:
- Go
- Lambda
- AWS AppConfig(機能フラグ設定管理)

### 6. auth-by-cognito
**概要**: Amazon CognitoによるIDトークンの検証\
**機能**:
- Cognitoユーザープールでの認証処理
- IDトークンの検証

**技術スタック**:
- Go
- Lambda
- Amazon Cognito(ユーザープール・認証管理)
- API Gateway(HTTP API)

### 7. tmp
**概要**: 一時的な実験用Lambda関数

**技術スタック**:
- Go
- Lambda
- 他

## 参考書籍

- GitHub CI/CD実践ガイド――持続可能なソフトウェア開発を支えるGitHub Actionsの設計と運用
- AWS Lambda実践ガイド 第2版
- AWSで実現するモダンアプリケーション入門 〜サーバーレス、コンテナ、マイクロサービスで何ができるのか
