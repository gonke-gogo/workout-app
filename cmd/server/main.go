package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golv2-learning-app/domain"
	"golv2-learning-app/repository"
	"golv2-learning-app/server"
	"golv2-learning-app/usecase"
)

// getEnv 環境変数を取得（必須）
func getEnv(key string, hideValue bool) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ 環境変数 %s が設定されていません", key)
	}
	if hideValue {
		log.Printf("✅ %s: [HIDDEN]", key)
	} else {
		log.Printf("✅ %s: %s", key, value)
	}
	return value
}

// getEnvWithDefault 環境変数を取得（デフォルト値あり）
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("⚠️  %s not set, using default: %s", key, defaultValue)
		return defaultValue
	}
	log.Printf("✅ %s: %s", key, value)
	return value
}

func main() {
	log.Printf("💪 筋トレアプリを起動中...")
	log.Printf("🔥 環境変数を読み込み中...")

	dbHost := getEnv("DB_HOST", false)
	dbName := getEnv("DB_NAME", false)
	dbUser := getEnv("DB_USER", false)
	dbPass := getEnv("DB_PASSWORD", true)

	var err error
	var dbPort, serverPort int

	dbPortStr := getEnvWithDefault("DB_PORT", "3306")
	dbPort, err = strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("❌ DB_PORT is invalid: %s", dbPortStr)
	}

	serverPortStr := getEnvWithDefault("GRPC_PORT", "50051")
	serverPort, err = strconv.Atoi(serverPortStr)
	if err != nil {
		log.Fatalf("❌ GRPC_PORT is invalid: %s", serverPortStr)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	log.Printf("🏋️ Database connection info: Host=%s, Port=%d, Database=%s, User=%s",
		dbHost, dbPort, dbName, dbUser)

	var repo domain.WorkoutRepository
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("💪 MySQLにGORMで接続中... (試行 %d/%d)", i+1, maxRetries)
		repo, err = repository.NewGORMRepository(dsn)
		if err == nil {
			break
		}
		log.Printf("MySQL接続に失敗: %v", err)
		if i < maxRetries-1 {
			log.Printf("🔄 %v後に再試行...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatalf("%d回の試行後もMySQLリポジトリの作成に失敗: %v", maxRetries, err)
	}
	defer repo.Close()

	log.Printf("✅ MySQLデータベースに接続しました: %s:%d/%s", dbHost, dbPort, dbName)

	// ワークアウトマネージャーを作成（MySQLリポジトリを使用）
	workoutManager := usecase.NewWorkoutManagerWithRepository(repo)

	// gRPCサーバーの作成と起動
	grpcServer := server.NewGRPCServer(workoutManager)

	log.Printf("🚀 ポート %d でgRPCサーバーを起動中...", serverPort)
	log.Printf("🎯 Evansで接続: evans -r repl -p %d", serverPort)
	log.Printf("💪 今日も筋肉を鍛えましょう！")

	if err := grpcServer.Start(serverPort); err != nil {
		log.Fatalf("💥 gRPCサーバーの起動に失敗: %v", err)
	}
}
