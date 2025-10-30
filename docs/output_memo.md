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

- パッケージの切り分けについて（クリーンアーキテクチャ）
  - ディレクトリ構成
    ```
    golv2-learning-app/
    ├── cmd/           # エントリーポイント層
    │   └── server/    # アプリケーション起動
    ├── server/        # プレゼンテーション層（gRPC API）
    ├── usecase/       # ビジネスロジック層（ユースケース）
    ├── domain/        # ドメイン層（エンティティ・ビジネスルール）
    │   ├── enums.go       # Enum定義
    │   ├── workout.go     # Workoutエンティティ
    │   └── repository.go  # リポジトリインターフェース
    ├── repository/    # データアクセス層（永続化の実装）
    ├── errors/        # エラー定義（独立パッケージ）
    └── proto/         # API定義層
    ```
  
  - 依存関係の方向（依存性の逆転）
    ```
    server → usecase → domain ← repository
                        ↓
                     errors
    ```
    - 全ての層がdomainに依存（内側に向かう依存）
    - domainは他のどの層にも依存しない（独立性）
    - repository層がdomainのインターフェースを実装
  
  - 各層の責務
    - domain層: ビジネスルール、エンティティ、インターフェース定義
    - usecase層: ユースケース（ビジネスロジック）の実装
    - repository層: データベースアクセスの具体的な実装（GORM使用）
    - server層: gRPCなどの外部インターフェース（プレゼンテーション）
    - errors層: アプリケーション全体で使うエラー型定義
  
  - ビジネスロジックが実装詳細に汚染されない設計
    - usecase層はインターフェース（domain.WorkoutRepository）のみに依存
    - usecase層はGORM、MySQL等の技術的詳細を知らない
    - 例：usecaseでは `wm.repo.CreateWorkout(workout)` と書くだけ
    - 実際にGORMを使うコードはrepository層に隔離
    - メリット:
      - DB変更（MySQL→PostgreSQL）してもusecaseは無変更
      - ORM変更（GORM→sqlx）してもusecaseは無変更
      - テスト時にモックに差し替え可能→DBサーバー不要
      - ビジネスルールが技術的詳細と混ざらず、可読性が高い

- 型定義
  - type WorkoutID int64
    type WorkoutStatus int
    type Difficulty int
    type MuscleGroup int
  - 違う種類の値を誤って渡した場合でもコンパイルエラーになるので型の安全性が守られる
  - 他の型との混同を防げる
  - 意図しない値を使用することを防止

- リクエスト構造体によるパラメータの一元管理
  - usecase/workout_usecase.go
  - CreateWorkoutRequest構造体で全パラメータを管理
    ```go
    type CreateWorkoutRequest struct {
        ExerciseType domain.ExerciseType // 必須: ワークアウトの種目
        Description  string               // オプション
        Difficulty   domain.Difficulty    // オプション
        MuscleGroup  domain.MuscleGroup   // オプション
        Sets         int32                // オプション
        Reps         int32                // オプション
        Weight       float64              // オプション
        Notes        string               // オプション
    }
    
    func (wm *WorkoutManager) CreateWorkout(req CreateWorkoutRequest) (*domain.Workout, error)
    ```
  
  - 使い方の例
    ```go
    // パターン1: 必須パラメータのみ（他はデフォルト値）
    workout, _ := manager.CreateWorkout(CreateWorkoutRequest{
        ExerciseType: domain.BenchPress,
    })
    
    // パターン2: 一部のオプションパラメータを指定
    workout, _ := manager.CreateWorkout(CreateWorkoutRequest{
        ExerciseType: domain.BenchPress,
        Difficulty:   domain.DifficultyAdvanced,
        Weight:       100.0,
    })
    
    // パターン3: 全てのパラメータを指定
    workout, _ := manager.CreateWorkout(CreateWorkoutRequest{
        ExerciseType: domain.Squat,
        Description:  "重量級トレーニング",
        Difficulty:   domain.DifficultyBeast,
        MuscleGroup:  domain.Legs,
        Sets:         5,
        Reps:         5,
        Weight:       150.0,
        Notes:        "フォーム要確認",
    })
    ```

  
  - メリット
    - 🎯 層の責務が明確: Server層は型変換のみ、バリデーションはUseCase層
    - 📝 可読性向上: 構造体のフィールドが一目瞭然
    - 🔧 保守性向上: フィールド追加が容易（1箇所追加するだけ）
    - ✅ テストが簡単: 1つの構造体で全パターンをテスト可能
    - 🔒 型安全性: IDEの補完が効く、typoを防げる
  
  - 従来の方法との比較
    - 悪い例（引数が多すぎる）:
      ```go
      func CreateWorkout(exerciseType, description, difficulty, muscleGroup, sets, reps, weight, notes) 
      // ↑ 引数に渡す順番を間違えたりするとバグる
      ```
    - 良い例（リクエスト構造体）:
      ```go
      func CreateWorkout(req CreateWorkoutRequest)
      // ↑ 1つの構造体だとフィールド値が増えても対応できるなど拡張しやすい、型安全
      ```
- iota
  - 定数の追加や削除があると、番号がズレるので注意が必要。意図せずそれを行うとバグの原因になりかねない。元々あった構造に対して、新たに値を割り込ませることはしないほうが良い。 

- errorに情報付加
  - server/workout_manager_repository.go<br>
    ```
    if err := wm.validateWorkoutData(workout); err != nil {
		workoutErr := &WorkoutError{
			Op:           "CreateWorkout",
			ExerciseType: exerciseType,
			Message:      "workout data validation failed",
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}
    ```
  - この部分でエラー用に定義した構造体に情報を付加して、ログに出力する

- domain/repository.go
  - 疎結合の実現（依存性の逆転原則）
    - インターフェースの配置: repositoryパッケージではなく、domainパッケージに配置
    - 依存の方向: usecase → domain ← repository（両方がdomainに依存）
    - 疎結合の効果:
      - usecaseはWorkoutRepositoryインターフェースだけを知り、具体的な実装（GORMRepository）を知らない
      - 実装の差し替えが容易（MySQL → PostgreSQL、本番 → テスト用モックなど）
      - テスト時にモックリポジトリに簡単に差し替え可能（DBサーバー不要）
    - 密結合との比較:
      - 悪い例: `repo *repository.GORMRepository` （具体的な型に依存）
      - 良い例: `repo domain.WorkoutRepository` （インターフェースに依存）
    - 結果: パッケージ間が疎結合になり、変更に強く、テストしやすい設計を実現

- インターフェースによる共通化
  - usecase/mock_repository.go & usecase/workout_usecase_test.go
  - 同じインターフェース、異なる実装
    ```
    domain.WorkoutRepository インターフェース
        ↓ これを実装
    ├─ repository.GORMRepository     (本番用: MySQL + GORM)
    └─ usecase.MockWorkoutRepository (テスト用: メモリ上)
    
    ↓ どちらも同じように使える！
    
    manager := usecase.NewWorkoutManagerWithRepository(repo)
    ```
  
  - 本番とテストで同じコードを使用（共通化）
    ```go
    // 本番環境
    repo, _ := repository.NewGORMRepository(dsn)
    manager := usecase.NewWorkoutManagerWithRepository(repo)
    
    // テスト環境
    mockRepo := usecase.NewMockWorkoutRepository()
    manager := usecase.NewWorkoutManagerWithRepository(mockRepo)
    //        
    ```
  
  - 実装したテストケース
    - TestCreateWorkout_WithMock: 基本的な作成テスト
    - TestGetWorkout_WithMock: 取得テスト
    - TestCreateWorkout_WithOptions: オプション付き作成テスト
    - TestListWorkouts_WithMock: 一覧取得テスト
    - TestDeleteWorkout_WithMock: 削除テスト
    - TestValidation_NegativeSets: バリデーションテスト
    - 全テスト成功、DBサーバー不要で実行可能 ✅
  
  - メリット
    - DBサーバー不要でテスト可能（超高速: 0.689s で全テスト完了）
    - 本番とテストでビジネスロジックのコードが共通化
    - テストが簡単（セットアップ・接続設定不要）
    - CI/CDパイプラインで即座にテスト実行可能
    - モックの動作を自由にコントロール可能（エラーケースのテストも容易）
  
  - 共通化の意義
    - インターフェースがあることで、実装の違いを吸収
    - 本番コードとテストコードで同じAPIを使用
    - 実装を追加しても既存コード（usecase層）は無変更
    - 将来的にPostgreSQLやRedis実装を追加しても、usecase層は変更不要
- 値にレシーバーメソッドを定義
  - domain/enums.go
  - 型定義に対してメソッドを定義できる
    ```go
    // 型定義
    type MuscleGroup int
    type ExerciseType int
    type Difficulty int
    
    // レシーバーメソッド定義
    func (mg MuscleGroup) Japanese() string {
        switch mg {
        case Chest:
            return "胸"
        case Back:
            return "背中"
        // ...
        }
    }
    ```
  
  - 使い方
    ```go
    muscle := domain.Chest
    fmt.Printf("%s", muscle.Japanese())   // "胸" (Japanese()メソッドで日本語表示)
    fmt.Printf("%d", muscle)              // "1" (基底型のint値として出力)
    ```
  
  - メリット
    - Enum値（数値）を人間が読める文字列に変換できる
    - 型に振る舞いを持たせられる（オブジェクト指向的）
    - 同じ型に複数のメソッドを定義できる（Japanese, Validate等）
    - 必要に応じて新しいメソッドを追加できる（柔軟性が高い）
  
  - レシーバーの種類
    - 値レシーバー: `func (mg MuscleGroup)` - 読み取り専用、コピーが渡される
    - ポインタレシーバー: `func (mg *MuscleGroup)` - 変更可能、元の値を変更できる
    - 今回は値レシーバーを使用（読み取り専用で十分）

- 文字列結合パターンと性能の違い
  - server/workout_server.go - buildWorkoutSummary
  - 3つの主な文字列結合パターン
  
  - パターン1: + 演算子（最も遅い）
    ```go
    result := ""
    for i := 0; i < 1000; i++ {
        result = result + "text"  // 毎回新しい文字列を作成
    }
    // 毎回メモリを割り当てなければならない
    ```
  
  - パターン2: strings.Builder（最速）✅
    ```go
    var builder strings.Builder
    builder.Grow(1000 * 4)  // 事前に容量確保
    for i := 0; i < 1000; i++ {
        builder.WriteString("text")  // 効率的に追加
    }
    result := builder.String()
    // 利点: メモリ再割り当てを最小化 → O(n)の計算量
    ```
  
  - 実装例（buildWorkoutSummary）
    ```go
    func (s *GRPCServer) buildWorkoutSummary(workouts []*domain.Workout) string {
        var builder strings.Builder
        // 事前容量確保（再割り当てを減らす）
        estimatedSize := len(workouts)*50 + 100
        builder.Grow(estimatedSize)
        
        builder.WriteString("📋 ワークアウト一覧 (")
        builder.WriteString(fmt.Sprintf("%d件", len(workouts)))
        builder.WriteString("):\n")
        
        for i, workout := range workouts {
            builder.WriteString(fmt.Sprintf("  %d. %s (%s)", 
                i+1, workout.ExerciseType.Japanese(), workout.MuscleGroup.Japanese()))
            // ...
        }
        
        return builder.String()
    }
    ```
  
  - パフォーマンス比較（概算）
    - + 演算子: 1000回の連結で 約100ms
    - strings.Builder: 1000回の連結で 約1ms（事前容量確保あり）
  
  - 使い分け
    - ループ内での文字列結合は必ず`strings.Builder`を使う
    - 可能なら`builder.Grow()`で事前容量確保
    - 単純な2-3個の文字列連結なら`+`でもOK

- リポジトリ層のテストをスタブ・モックで実装
  - repository/workout_repository_test.go
  - go-sqlmockを使用したモックテスト（実DBなしでテスト）
  
  - モック vs スタブ
    - スタブ: 特定の値を返すだけ（例: SQLiteインメモリDB）
    - モック: 呼び出しの検証も行う（例: go-sqlmock）
    - このコードはモック（呼び出しの順番・引数・回数を厳密にチェック）
  
  - SQL文の変数定義（DRY原則）
    ```go
    // ファイル上部で一度だけ定義（再利用可能）
    var (
        insertWorkoutQuery = regexp.QuoteMeta("INSERT INTO `workouts`")
        selectWorkoutQuery = regexp.QuoteMeta("SELECT * FROM `workouts` WHERE `workouts`.`id` = ? ORDER BY `workouts`.`id` LIMIT 1")
        updateWorkoutQuery = regexp.QuoteMeta("UPDATE `workouts` SET")
        deleteWorkoutQuery = regexp.QuoteMeta("DELETE FROM `workouts` WHERE `workouts`.`id` = ?")
    )
    ```
    - メリット: SQL文を変更する時は1箇所だけ修正すればOK
    - メリット: テストケース内で同じSQL文を何度も書かなくて済む
  
  - テーブル駆動テストの構造
    ```go
    tests := []struct {
        name         string        // テストケース名
        workout      *domain.Workout
        mockAffected int64         // ① DB操作で影響を受けた行数
        mockError    error         // ② シミュレートするエラー
        wantErr      bool          // ③ エラーを期待するか
        description  string        // ④ テストの説明（オプション）
    }{
        {
            name:         "正常系: ワークアウト更新",
            mockAffected: 1,      // 1行が更新される
            mockError:    nil,    // エラーなし
            wantErr:      false,  // エラーを期待しない
        },
        {
            name:         "異常系: DB接続エラー",
            mockAffected: 0,
            mockError:    sql.ErrConnDone,  // 接続エラーをシミュレート
            wantErr:      true,             // エラーを期待する
        },
    }
    ```
  
  - フィールドの役割
    - `mockAffected`: UPDATE/DELETE文で「何行が影響を受けたか」をシミュレート
    - `mockError`: DB操作時のエラーを意図的に発生させる（異常系テストに必須）
    - `wantErr`: テストの合否判定に使用（エラーの有無をチェック）
    - `description`: テストケースの説明（ドキュメント用、オプション）
  
  - モックの期待値設定の流れ
    ```go
    // ① モックDBのセットアップ
    repo, mock, db := setupMockDB(t)
    defer db.Close()
    
    // ② トランザクション開始の期待
    mock.ExpectBegin()
    
    // ③ エラーケースと正常系で分岐
    if tt.mockError != nil {
        // エラーケース: エラーを返す
        mock.ExpectExec(insertWorkoutQuery).
            WillReturnError(tt.mockError)
        mock.ExpectRollback()  // ロールバックされる
    } else {
        // 正常系: 成功結果を返す
        mock.ExpectExec(insertWorkoutQuery).
            WithArgs(
                sqlmock.AnyArg(),        // exercise_type（どんな値でもOK）
                tt.workout.Description,  // 具体的な値をチェック
                sqlmock.AnyArg(),        // status
                // ... 他の引数
            ).
            WillReturnResult(sqlmock.NewResult(tt.mockResultID, tt.mockAffected))
        mock.ExpectCommit()    // コミットされる
    }
    
    // ④ 実際のテスト実行
    err := repo.CreateWorkout(tt.workout)
    
    // ⑤ エラーチェック
    if (err != nil) != tt.wantErr {
        t.Errorf("CreateWorkout() error = %v, wantErr %v", err, tt.wantErr)
    }
    
    // ⑥ モックの期待値が全て満たされたか確認
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled mock expectations: %v", err)
    }
    ```
  
  - WithArgs()の詳細解説
    
    WithArgs()とは？
    - SQL実行時に渡される引数（プレースホルダー`?`の値）を検証するメソッド
    - GORMが生成するINSERT文の引数の順番と値をチェックする
    
    実際のSQL文との対応
    ```sql
    -- GORMが生成するSQL
    INSERT INTO `workouts` 
        (`exercise_type`, `description`, `status`, `difficulty`, 
         `muscle_group`, `sets`, `reps`, `weight`, `notes`, 
         `created_at`, `updated_at`, `completed_at`) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    --      ↑1 ↑2 ↑3 ↑4 ↑5 ↑6 ↑7 ↑8 ↑9 ↑10↑11↑12
    ```
    
    引数の検証戦略
    ```go
    mock.ExpectExec(insertWorkoutQuery).
        WithArgs(
            sqlmock.AnyArg(),        // ① exercise_type (Enum値)
            tt.workout.Description,  // ② description (ユーザー入力値)
            sqlmock.AnyArg(),        // ③ status (Enum値)
            sqlmock.AnyArg(),        // ④ difficulty (Enum値)
            sqlmock.AnyArg(),        // ⑤ muscle_group (Enum値)
            tt.workout.Sets,         // ⑥ sets (ユーザー入力値)
            tt.workout.Reps,         // ⑦ reps (ユーザー入力値)
            tt.workout.Weight,       // ⑧ weight (ユーザー入力値)
            tt.workout.Notes,        // ⑨ notes (ユーザー入力値)
            sqlmock.AnyArg(),        // ⑩ created_at (自動生成)
            sqlmock.AnyArg(),        // ⑪ updated_at (自動生成)
            sqlmock.AnyArg(),        // ⑫ completed_at (NULL可)
        )
    ```
    
    各引数の検証方針と理由
    
    | 引数 | 検証方法 | 理由 |
    |-----|---------|------|
    | ① exercise_type | `AnyArg()` | テストケースごとに異なるEnum値。型が正しければOK |
    | ② description | 具体的な値 | ユーザー入力。正しく渡されているか検証したい |
    | ③ status | `AnyArg()` | デフォルト値（Planned）。値は気にしない |
    | ④ difficulty | `AnyArg()` | テストケースで変わる。型が正しければOK |
    | ⑤ muscle_group | `AnyArg()` | テストケースで変わる。型が正しければOK |
    | ⑥ sets | 具体的な値 | ユーザー入力。3が渡されたら3であるべき |
    | ⑦ reps | 具体的な値 | ユーザー入力。10が渡されたら10であるべき |
    | ⑧ weight | 具体的な値 | ユーザー入力。60.0が渡されたら60.0であるべき |
    | ⑨ notes | 具体的な値 | ユーザー入力。メモが正しく保存されるか検証 |
    | ⑩ created_at | `AnyArg()` | システムが自動生成。時刻の正確性は関係ない |
    | ⑪ updated_at | `AnyArg()` | システムが自動生成。時刻の正確性は関係ない |
    | ⑫ completed_at | `AnyArg()` | 初期はNULL。値は関係ない |
    
    なぜ引数の順番が重要か？
    - GORMは内部でINSERT文を生成する際、構造体のフィールド順でSQL引数を並べる
    - モックは「1番目の引数、2番目の引数...」と順番でチェックする
    - 順番が1つでもズレると「期待と違う！」とエラーになる
    
    sqlmock.AnyArg()を使う3つの理由
    
    1. 動的に変わる値（Enum）
       ```go
       // テストケース1
       workout.ExerciseType = domain.BenchPress  // = 1
       
       // テストケース2
       workout.ExerciseType = domain.Squat       // = 2
       
       // 毎回値が変わるので、AnyArg()で「どんな値でもOK」とする
       ```
    
    2. システムが自動生成する値（タイムスタンプ）
       ```go
       workout.CreatedAt = time.Now()
       // → 2025-10-29 22:30:15.123456789
       
       // 実行のたびに異なる時刻になるので、厳密にチェックできない
       // AnyArg()で「time.Timeが渡されていればOK」とする
       ```
    
    3. テストの本質と関係ない値
       ```go
       workout.CompletedAt = nil  // 初期はNULL
       
       // このテストでは「完了日時」は重要ではない
       // AnyArg()で「とりあえず何か渡されていればOK」
       ```
    
    具体的な値をチェックする理由
    
    ```go
    // テストケース
    workout := &domain.Workout{
        Description: "テスト用のベンチプレス",
        Sets:        3,
        Reps:        10,
        Weight:      60.0,
    }
    
    // WithArgsで検証
    WithArgs(
        sqlmock.AnyArg(),
        tt.workout.Description,  // "テスト用のベンチプレス"
        sqlmock.AnyArg(),
        sqlmock.AnyArg(),
        sqlmock.AnyArg(),
        tt.workout.Sets,         // 3
        tt.workout.Reps,         // 10
        tt.workout.Weight,       // 60.0
        // ...
    )
    
    // もしGORMが間違ったフィールドをINSERTしようとしたら？
    // 例: SetsとRepsが逆になっていたら？
    // → モックが検知してテスト失敗！
    ```
    
    検証のバランス
    - ❌ 全部AnyArg(): テストが甘すぎる（バグを見逃す可能性）
    - ❌ 全部具体的な値: テストが厳密すぎる（メンテナンスが大変）
    - ✅ 重要な値だけ具体的にチェック: バランスが良い
    
    実際のテストケースでの動き
    ```go
    // テストケース: ベンチプレス作成
    {
        workout: &domain.Workout{
            ExerciseType: domain.BenchPress,  // = 1
            Description:  "テスト",
            Status:       domain.WorkoutStatusPlanned,  // = 0
            Sets:         3,
            Reps:         10,
            Weight:       60.0,
            Notes:        "メモ",
        },
    }
    
    // モックの期待値
    mock.ExpectExec(insertWorkoutQuery).
        WithArgs(
            sqlmock.AnyArg(),   // 1 (BenchPress) → OK、数値が来た
            "テスト",            // "テスト" → OK、一致した
            sqlmock.AnyArg(),   // 0 (Planned) → OK、数値が来た
            sqlmock.AnyArg(),   // 任意の難易度 → OK
            sqlmock.AnyArg(),   // 任意の部位 → OK
            3,                  // 3 → OK、一致した
            10,                 // 10 → OK、一致した
            60.0,               // 60.0 → OK、一致した
            "メモ",              // "メモ" → OK、一致した
            sqlmock.AnyArg(),   // time.Now() → OK、時刻が来た
            sqlmock.AnyArg(),   // time.Now() → OK、時刻が来た
            sqlmock.AnyArg(),   // nil → OK、何か来た
        )
    
    // 全ての引数がマッチ → テスト成功 ✅
    ```
    
    もし引数が間違っていたら？
    ```go
    // バグのあるコード（SetsとRepsが逆）
    INSERT INTO workouts (...) VALUES (?, ?, ..., 10, 3, ...)
    //                                            ↑   ↑
    //                                          Reps Sets（逆！）
    
    // モックの期待値
    WithArgs(..., 3, 10, ...)  // Sets=3, Reps=10を期待
    //             ↑   ↑
    
    // 実際に来た値
    VALUES (..., 10, 3, ...)   // Reps=10, Sets=3（逆！）
    //           ↑   ↑
    
    // → 不一致！テスト失敗 ❌
    // エラーメッセージ: "argument 5: expected 3, got 10"
    ```
  
  - モックの動作フロー
    ```
    GORM: "BEGIN実行します"
    Mock: "OK、期待通り！（ExpectBegin()）"
    
    GORM: "INSERT実行します、引数は..."
    Mock: "OK、期待通り！（ExpectExec + WithArgs）"
    Mock: "結果を返すよ！（WillReturnResult）"
    
    GORM: "COMMIT実行します"
    Mock: "OK、期待通り！（ExpectCommit）"
    ```
  
  - エラーチェックのロジック
    ```go
    if (err != nil) != tt.wantErr {
        // (実際にエラー発生) != (エラーを期待)
    }
    
    // パターン1: 正常系が成功 → PASS
    // (nil != nil) != false → false != false → ✅
    
    // パターン2: 異常系が正しくエラー → PASS
    // (エラー != nil) != true → true != true → ✅
    
    // パターン3: エラーが出るべきなのに出ない → FAIL
    // (nil != nil) != true → false != true → ❌
    
    // パターン4: エラーが出るべきでないのに出た → FAIL
    // (エラー != nil) != false → true != false → ❌
    ```
  
  - メリット
    - ✅ 実DBなしでテスト可能（高速・安定・CI/CDで即実行）
    - ✅ SQL実行の順序・引数・結果を厳密に検証
    - ✅ トランザクションの挙動（BEGIN/COMMIT/ROLLBACK）も検証
    - ✅ 正常系と異常系の両方をテスト可能
    - ✅ 様々なエラーシナリオを簡単にシミュレート
    - ✅ テーブル駆動テストで複数パターンを効率的にテスト
