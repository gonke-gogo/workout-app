package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golv2-learning-app/repository"
	"golv2-learning-app/server"
)

func main() {
	var (
		port   = flag.Int("port", 50051, "gRPC server port")
		dbHost = flag.String("db-host", "localhost", "Database host")
		dbPort = flag.Int("db-port", 3306, "Database port")
		dbName = flag.String("db-name", "workoutdb", "Database name")
		dbUser = flag.String("db-user", "workoutuser", "Database user")
		dbPass = flag.String("db-pass", "workoutpass", "Database password")
	)
	flag.Parse()

	// 環境変数から設定を読み込み
	log.Printf("💪 筋トレアプリを起動中...")
	log.Printf("🔥 環境変数を読み込み中...")

	if envHost := os.Getenv("DB_HOST"); envHost != "" {
		*dbHost = envHost
		log.Printf("DB_HOST from env: %s", envHost)
	}
	if envPort := os.Getenv("DB_PORT"); envPort != "" {
		log.Printf("DB_PORT from env: %s", envPort)
		if port, err := strconv.Atoi(envPort); err != nil {
			log.Printf("Invalid DB_PORT: %s, using default: %d", envPort, *dbPort)
		} else {
			*dbPort = port
			log.Printf("DB_PORT parsed: %d", *dbPort)
		}
	}
	if envName := os.Getenv("DB_NAME"); envName != "" {
		*dbName = envName
		log.Printf("DB_NAME from env: %s", envName)
	}
	if envUser := os.Getenv("DB_USER"); envUser != "" {
		*dbUser = envUser
		log.Printf("DB_USER from env: %s", envUser)
	}
	if envPass := os.Getenv("DB_PASSWORD"); envPass != "" {
		*dbPass = envPass
		log.Printf("DB_PASSWORD from env: [HIDDEN]")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		*dbUser, *dbPass, *dbHost, *dbPort, *dbName)

	log.Printf("🏋️ Database connection info: Host=%s, Port=%d, Database=%s, User=%s",
		*dbHost, *dbPort, *dbName, *dbUser)

	var repo repository.WorkoutRepository
	var err error
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

	log.Printf("✅ MySQLデータベースに接続しました: %s:%d/%s", *dbHost, *dbPort, *dbName)

	// ワークアウトマネージャーを作成（MySQLリポジトリを使用）
	workoutManager := server.NewWorkoutManagerWithRepository(repo)

	// gRPCサーバーの作成と起動
	grpcServer := server.NewGRPCServer(workoutManager)

	log.Printf("🚀 ポート %d でgRPCサーバーを起動中...", *port)
	log.Printf("🎯 Evansで接続: evans -r repl -p %d", *port)
	log.Printf("💪 今日も筋肉を鍛えましょう！")

	if err := grpcServer.Start(*port); err != nil {
		log.Fatalf("💥 gRPCサーバーの起動に失敗: %v", err)
	}
}
