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
		return fmt.Errorf("failed to create workout (exercise_type=%d): %w", workout.ExerciseType, err)
	}
	return nil
}

// GetWorkout ワークアウトをIDで取得
func (r *GORMRepository) GetWorkout(id domain.WorkoutID) (*domain.Workout, error) {
	var workout domain.Workout
	if err := r.db.First(&workout, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workout not found (id=%d): %w", id, err)
		}
		return nil, fmt.Errorf("failed to get workout (id=%d): %w", id, err)
	}

	return &workout, nil
}

// UpdateWorkout ワークアウトを更新
func (r *GORMRepository) UpdateWorkout(workout *domain.Workout) error {
	workout.UpdatedAt = time.Now()
	if err := r.db.Save(workout).Error; err != nil {
		return fmt.Errorf("failed to update workout (id=%d, exercise_type=%d): %w", workout.ID, workout.ExerciseType, err)
	}
	return nil
}

// DeleteWorkout ワークアウトを削除
func (r *GORMRepository) DeleteWorkout(id domain.WorkoutID) error {
	if err := r.db.Delete(&domain.Workout{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete workout (id=%d): %w", id, err)
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
