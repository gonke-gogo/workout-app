# ビルドステージ
FROM golang:1.21-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache protobuf-dev

# 作業ディレクトリを設定
WORKDIR /app

# go.modとgo.sumをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# Go gRPCプラグインをインストール（Go 1.21互換バージョン）
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# ソースコードをコピー
COPY . .

# プロトコルバッファを生成
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/workout.proto

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 実行ステージ
FROM alpine:latest

# ca-certificatesをインストール（HTTPS通信用）
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドしたアプリケーションをコピー
COPY --from=builder /app/main .

# ポートを公開
EXPOSE 50051

# アプリケーションを実行
CMD ["./main"]
