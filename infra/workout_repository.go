package repository

import (
	"fmt"
	"time"

	"golv2-learning-app/domain"

	"gorm.io/gorm"
)

// GORMRepository GORMを使用したリポジトリ実装
type GORMRepository struct {
	db *gorm.DB
}

// NewGORMRepository 接続済みのGORM DBインスタンスからリポジトリを作成
func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

// CreateWorkout ワークアウトを作成
func (r *GORMRepository) CreateWorkout(workout *domain.Workout) error {
	if err := r.db.Create(workout).Error; err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}
	return nil
}

// GetWorkout ワークアウトをIDで取得
func (r *GORMRepository) GetWorkout(id domain.WorkoutID) (*domain.Workout, error) {
	var workout domain.Workout
	if err := r.db.First(&workout, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workout not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	return &workout, nil
}

// UpdateWorkout ワークアウトを更新
func (r *GORMRepository) UpdateWorkout(workout *domain.Workout) error {
	workout.UpdatedAt = time.Now()
	if err := r.db.Save(workout).Error; err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}
	return nil
}

// DeleteWorkout ワークアウトを削除
func (r *GORMRepository) DeleteWorkout(id domain.WorkoutID) error {
	if err := r.db.Delete(&domain.Workout{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}
	return nil
}

// ListWorkouts ワークアウト一覧を取得
func (r *GORMRepository) ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*domain.Workout, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		fmt.Printf("🔍 ListWorkouts実行時間: %v\n", duration)
	}()

	// ここを変えて性能評価
	workouts := make([]*domain.Workout, 0, 100)

	query := r.db.Model(&domain.Workout{})

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

	if err := query.Order("created_at DESC").Find(&workouts).Error; err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}

	fmt.Printf("🎯 取得件数: %d件\n", len(workouts))
	return workouts, nil
}

// GetWorkoutCount ワークアウト数を取得
func (r *GORMRepository) GetWorkoutCount() (int, error) {
	var count int64
	if err := r.db.Model(&domain.Workout{}).Count(&count).Error; err != nil {
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
	if err := r.db.Model(&domain.Workout{}).Where("created_at >= ?", timeFilter).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total workout count: %w", err)
	}
	stats["total_workouts"] = int(totalCount)

	// 完了したワークアウト数
	var completedCount int64
	if err := r.db.Model(&domain.Workout{}).Where("status = ? AND created_at >= ?", domain.WorkoutStatusCompleted, timeFilter).Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed workout count: %w", err)
	}
	stats["completed_workouts"] = int(completedCount)

	// スキップしたワークアウト数
	var skippedCount int64
	if err := r.db.Model(&domain.Workout{}).Where("status = ? AND created_at >= ?", domain.WorkoutStatusSkipped, timeFilter).Count(&skippedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get skipped workout count: %w", err)
	}
	stats["skipped_workouts"] = int(skippedCount)

	// 総重量
	var totalWeight float64
	if err := r.db.Model(&domain.Workout{}).Where("status = ? AND created_at >= ?", domain.WorkoutStatusCompleted, timeFilter).Select("SUM(weight * sets * reps)").Scan(&totalWeight).Error; err != nil {
		return nil, fmt.Errorf("failed to get total weight: %w", err)
	}
	stats["total_weight_lifted"] = totalWeight

	// 筋肉群別統計
	var muscleGroupStats []struct {
		MuscleGroup string `json:"muscle_group"`
		Count       int    `json:"count"`
	}
	if err := r.db.Model(&domain.Workout{}).Where("created_at >= ?", timeFilter).Select("muscle_group, COUNT(*) as count").Group("muscle_group").Scan(&muscleGroupStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get muscle group stats: %w", err)
	}

	muscleGroupMap := make(map[string]int)
	for _, stat := range muscleGroupStats {
		muscleGroupMap[stat.MuscleGroup] = stat.Count
	}
	stats["muscle_group_stats"] = muscleGroupMap

	return stats, nil
}
