.PHONY: build proto server client test clean

# デフォルトターゲット
all: proto build

# プロトコルバッファの生成
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/workout.proto


# アプリケーションのビルド
build: proto
	go build -o bin/taskmanager cmd/server/main.go

# Docker Composeで起動
compose-up:
	docker-compose up --build

# Docker Composeでバックグラウンド起動
compose-up-d:
	docker-compose up -d --build

# Docker Composeで停止
compose-down:
	docker-compose down

# Docker Composeでログ確認
docker-logs:
	docker-compose logs -f

# データベースのリセット
docker-reset:
	docker-compose down -v
	docker-compose up --build

# サーバーの実行
server: build
	./bin/taskmanager

# サーバーの実行（ポート指定）
server-port: build
	./bin/taskmanager -port=50052

# テストの実行
test:
	go test ./...

# ベンチマークの実行
bench:
	go test -bench=. ./...

# カバレッジの確認
coverage:
	go test -cover ./...

# 依存関係の整理
deps:
	go mod tidy

# クリーンアップ
clean:
	rm -rf bin/
	rm -f proto/*.pb.go

# Evansのインストール（macOS）
install-evans:
	brew install evans

# Evansのインストール（Linux）
install-evans-linux:
	curl -L https://github.com/ktr0731/evans/releases/latest/download/evans_linux_amd64.tar.gz | tar xz
	sudo mv evans /usr/local/bin/

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  make proto        - プロトコルバッファファイルを生成"
	@echo "  make build        - アプリケーションをビルド"
	@echo "  make server       - サーバーを起動（ポート50051）"
	@echo "  make server-port  - サーバーを起動（ポート50052）"
	@echo "  make test         - テストを実行"
	@echo "  make bench        - ベンチマークを実行"
	@echo "  make coverage     - カバレッジを確認"
	@echo "  make deps         - 依存関係を整理"
	@echo "  make clean        - ビルドファイルを削除"
	@echo "  make install-evans - Evansをインストール（macOS）"
	@echo "  make help         - このヘルプを表示"
