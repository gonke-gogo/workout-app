- ListAPI
  - repository/workout_repository.go
    - `workouts := make([]*Workout, 0, 100)`
      - capacityを事前指定
    - defer func() {
        duration := time.Since(start)
        fmt.Printf("🔍 ListWorkouts実行時間: %v\n", duration)
      }()
      - 関数終了時に処理を予約
      - 正常終了でもエラー終了でも必ず実行
  - server/grpc_server_repository.go
    - `convertedWorkouts := make([]*proto.Workout, 0, len(workouts))`
      - 取得できる全件を基準に容量を確保

- パッケージの切り分けについて
  - golv2-learning-app/
    ├── cmd/           # エントリーポイント層
    ├── server/        # プレゼンテーション層（gRPC）
    ├── repository/    # データアクセス層
    ├── proto/         # API定義層
    └── utils/         # 共通ユーティリティ層
