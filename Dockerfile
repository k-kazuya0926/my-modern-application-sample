# ビルド用ステージ
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# セキュリティ: CGOを無効化し、静的リンクでビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o bootstrap main.go

# 実行用ステージ（Lambda公式イメージ）
FROM public.ecr.aws/lambda/provided:al2023

# ビルドしたバイナリをコピー
COPY --from=build /app/bootstrap ${LAMBDA_RUNTIME_DIR}

# 実行権限を確認
RUN chmod +x ${LAMBDA_RUNTIME_DIR}/bootstrap

# Lambdaエントリポイントはデフォルトで "bootstrap"
CMD ["bootstrap"]
