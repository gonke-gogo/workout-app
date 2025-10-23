package repository

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMRepository GORMを使用したリポジトリ実装
type GORMRepository struct {
	db *gorm.DB
}

// NewGORMRepository 新しいGORMリポジトリを作成
func NewGORMRepository(dsn string) (*GORMRepository, error) {
	// パフォーマンス最適化設定
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// PrepareStmt: true, // プリペアドステートメントでパフォーマンス向上
	}

	// MySQL設定でUTF-8を明示的に指定
	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,   // string型のデフォルトサイズ
		DisableDatetimePrecision:  true,  // datetime精度を無効化
		DontSupportRenameIndex:    true,  // インデックスリネームをサポートしない
		DontSupportRenameColumn:   true,  // カラムリネームをサポートしない
		SkipInitializeWithVersion: false, // バージョン初期化をスキップしない
	}

	db, err := gorm.Open(mysql.New(mysqlConfig), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// UTF-8設定を明示的に実行
	db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci")
	db.Exec("SET CHARACTER SET utf8mb4")

	// 接続プール設定（パフォーマンス最適化）
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 接続プールの最適化
	sqlDB.SetMaxOpenConns(25)                 // 最大接続数
	sqlDB.SetMaxIdleConns(25)                 // アイドル接続数
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // 接続の最大生存時間

	// マイグレーション実行
	if err := db.AutoMigrate(&Workout{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &GORMRepository{db: db}, nil
}

// CreateWorkout ワークアウトを作成
func (r *GORMRepository) CreateWorkout(workout *Workout) error {
	if err := r.db.Create(workout).Error; err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}
	return nil
}

// GetWorkoutByID ワークアウトをIDで取得
func (r *GORMRepository) GetWorkoutByID(id WorkoutID) (*Workout, error) {
	var workout Workout
	if err := r.db.First(&workout, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workout not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	return &workout, nil
}

// UpdateWorkout ワークアウトを更新
func (r *GORMRepository) UpdateWorkout(workout *Workout) error {
	workout.UpdatedAt = time.Now()
	if err := r.db.Save(workout).Error; err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}
	return nil
}

// DeleteWorkout ワークアウトを削除
func (r *GORMRepository) DeleteWorkout(id WorkoutID) error {
	if err := r.db.Delete(&Workout{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}
	return nil
}

// ListWorkouts ワークアウト一覧を取得（Go基礎技術による最適化版）
func (r *GORMRepository) ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*Workout, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("🔍 ListWorkouts実行時間: %v\n", duration)
	}()

	// ここを変えて性能評価
	workouts := make([]*Workout, 0, 100)

	query := r.db.Model(&Workout{})

	// インデックスを効率的に使用するクエリ構築
	// 複合インデックスの左端カラムを優先
	if statusFilter != nil && difficultyFilter != nil {
		// idx_workouts_status_difficulty を使用
		query = query.Where("status = ? AND difficulty = ?", *statusFilter, *difficultyFilter)
	} else if statusFilter != nil && muscleGroupFilter != nil {
		// idx_workouts_status_muscle を使用
		query = query.Where("status = ? AND muscle_group = ?", *statusFilter, *muscleGroupFilter)
	} else {
		// 個別の条件
		if statusFilter != nil {
			query = query.Where("status = ?", *statusFilter)
		}
		if difficultyFilter != nil {
			query = query.Where("difficulty = ?", *difficultyFilter)
		}
		if muscleGroupFilter != nil {
			query = query.Where("muscle_group = ?", *muscleGroupFilter)
		}
	}

	// ORDER BYにインデックスを使用
	if err := query.Order("created_at DESC").Find(&workouts).Error; err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}

	fmt.Printf("🎯 取得件数: %d件\n", len(workouts))
	return workouts, nil
}

// BuildWorkoutSummary Go基礎技術による効率的な文字列構築
func (r *GORMRepository) BuildWorkoutSummary(workouts []*Workout) string {
	if len(workouts) == 0 {
		return "ワークアウトなし"
	}

	// Go基礎技術2: strings.Builder + 事前容量確保
	var builder strings.Builder
	// 概算容量を計算（各ワークアウト名 + フォーマット文字列）
	estimatedSize := len(workouts) * 30 // 平均30文字と仮定
	builder.Grow(estimatedSize)

	builder.WriteString("📋 ワークアウト一覧:\n")

	for i, workout := range workouts {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("  %d. %s (%s) - %dセット×%d回",
			i+1, workout.Name, workout.MuscleGroup, workout.Sets, workout.Reps))

		if workout.Weight > 0 {
			builder.WriteString(fmt.Sprintf(" @ %.1fkg", workout.Weight))
		}
	}

	return builder.String()
}

// FilterWorkoutsByStatus Go基礎技術による効率的なフィルタリング
func (r *GORMRepository) FilterWorkoutsByStatus(workouts []*Workout, targetStatus WorkoutStatus) []*Workout {
	// Go基礎技術3: append最適化 - 事前容量確保
	// 結果サイズを推定（全体の約1/3がマッチすると仮定）
	estimatedSize := len(workouts) / 3
	if estimatedSize < 10 {
		estimatedSize = 10 // 最小容量
	}

	filtered := make([]*Workout, 0, estimatedSize)

	for _, workout := range workouts {
		if workout.Status == targetStatus {
			filtered = append(filtered, workout)
		}
	}

	return filtered
}

// BatchCreateWorkouts Go基礎技術によるバッチ作成
func (r *GORMRepository) BatchCreateWorkouts(workouts []*Workout, batchSize int) error {
	if len(workouts) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("📦 BatchCreateWorkouts実行時間: %v (%d件)\n", duration, len(workouts))
	}()

	// Go基礎技術: 効率的なバッチ処理
	for i := 0; i < len(workouts); i += batchSize {
		end := i + batchSize
		if end > len(workouts) {
			end = len(workouts)
		}

		// バッチスライスを作成（容量最適化）
		batch := make([]*Workout, 0, end-i)
		batch = append(batch, workouts[i:end]...)

		// トランザクション内でバッチ処理
		if err := r.db.Create(&batch).Error; err != nil {
			return fmt.Errorf("batch create failed at index %d: %w", i, err)
		}
	}

	return nil
}

// GetWorkoutCount ワークアウト数を取得
func (r *GORMRepository) GetWorkoutCount() (int, error) {
	var count int64
	if err := r.db.Model(&Workout{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get workout count: %w", err)
	}
	return int(count), nil
}

// GetWorkoutStats ワークアウト統計を取得
func (r *GORMRepository) GetWorkoutStats(period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 期間フィルタを設定
	var timeFilter time.Time
	switch period {
	case "today":
		timeFilter = time.Now().Truncate(24 * time.Hour)
	case "week":
		timeFilter = time.Now().AddDate(0, 0, -7)
	case "month":
		timeFilter = time.Now().AddDate(0, -1, 0)
	default:
		timeFilter = time.Now().AddDate(0, 0, -30) // デフォルトは30日
	}

	// 総ワークアウト数
	var totalCount int64
	if err := r.db.Model(&Workout{}).Where("created_at >= ?", timeFilter).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total workout count: %w", err)
	}
	stats["total_workouts"] = int(totalCount)

	// 完了したワークアウト数
	var completedCount int64
	if err := r.db.Model(&Workout{}).Where("status = ? AND created_at >= ?", WorkoutStatusCompleted, timeFilter).Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed workout count: %w", err)
	}
	stats["completed_workouts"] = int(completedCount)

	// スキップしたワークアウト数
	var skippedCount int64
	if err := r.db.Model(&Workout{}).Where("status = ? AND created_at >= ?", WorkoutStatusSkipped, timeFilter).Count(&skippedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get skipped workout count: %w", err)
	}
	stats["skipped_workouts"] = int(skippedCount)

	// 総重量
	var totalWeight float64
	if err := r.db.Model(&Workout{}).Where("status = ? AND created_at >= ?", WorkoutStatusCompleted, timeFilter).Select("SUM(weight * sets * reps)").Scan(&totalWeight).Error; err != nil {
		return nil, fmt.Errorf("failed to get total weight: %w", err)
	}
	stats["total_weight_lifted"] = totalWeight

	// 筋肉群別統計
	var muscleGroupStats []struct {
		MuscleGroup string `json:"muscle_group"`
		Count       int    `json:"count"`
	}
	if err := r.db.Model(&Workout{}).Where("created_at >= ?", timeFilter).Select("muscle_group, COUNT(*) as count").Group("muscle_group").Scan(&muscleGroupStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get muscle group stats: %w", err)
	}

	muscleGroupMap := make(map[string]int)
	for _, stat := range muscleGroupStats {
		muscleGroupMap[stat.MuscleGroup] = stat.Count
	}
	stats["muscle_group_stats"] = muscleGroupMap

	return stats, nil
}

// Close リソースを解放
func (r *GORMRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}
