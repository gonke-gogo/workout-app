package server

import (
	"fmt"
	"strings"
	"time"

	"golv2-learning-app/repository"
	"golv2-learning-app/utils"
)

// WorkoutManager リポジトリベースのワークアウトマネージャー
type WorkoutManager struct {
	repo repository.WorkoutRepository
}

// WorkoutOption ワークアウト作成オプション
type WorkoutOption struct {
	Description string
	Difficulty  repository.Difficulty
	MuscleGroup repository.MuscleGroup
	Sets        int
	Reps        int
	Weight      float64
	Notes       string
}

// ファクトリー関数
func NewWorkoutManager() *WorkoutManager {
	return &WorkoutManager{
		repo: nil, // メモリベース（後方互換性のため）
	}
}

// リポジトリを使用するファクトリー関数
func NewWorkoutManagerWithRepository(repo repository.WorkoutRepository) *WorkoutManager {
	return &WorkoutManager{
		repo: repo,
	}
}

// GORMリポジトリを使用するファクトリー関数
func NewWorkoutManagerWithGORM(dsn string) (*WorkoutManager, error) {
	repo, err := repository.NewGORMRepository(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM repository: %w", err)
	}
	return &WorkoutManager{
		repo: repo,
	}, nil
}

func (wm *WorkoutManager) CreateWorkout(name string, opts ...WorkoutOption) (*repository.Workout, error) {
	// defer でのログ記録とエラーハンドリング
	start := time.Now()
	fmt.Printf("🏃 ワークアウト作成開始: %s\n", name)

	defer func() {
		duration := time.Since(start)
		fmt.Printf("🏁 ワークアウト作成処理終了: %s (実行時間: %v)\n", name, duration)
	}()

	// panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("💥 ワークアウト作成中にpanic発生: %s - %v\n", name, r)
		}
	}()

	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("workout name cannot be empty")
	}

	if wm.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	workout := &repository.Workout{
		Name:       name,
		Status:     repository.WorkoutStatusPlanned,
		Difficulty: repository.DifficultyBeginner,
		Sets:       3,
		Reps:       10,
		Weight:     0.0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// オプション引数の処理
	for _, opt := range opts {
		if opt.Description != "" {
			workout.Description = opt.Description
		}
		if opt.Difficulty != 0 {
			workout.Difficulty = opt.Difficulty
		}
		if opt.MuscleGroup != repository.Unspecified {
			workout.MuscleGroup = opt.MuscleGroup
		}
		if opt.Sets > 0 {
			workout.Sets = opt.Sets
		}
		if opt.Reps > 0 {
			workout.Reps = opt.Reps
		}
		if opt.Weight >= 0 {
			workout.Weight = opt.Weight
		}
		if opt.Notes != "" {
			workout.Notes = opt.Notes
		}
	}

	err := wm.repo.CreateWorkout(workout)
	if err != nil {
		return nil, fmt.Errorf("failed to create workout: %v", err)
	}

	// 楽しいメッセージをログに出力
	difficultyNames := map[repository.Difficulty]string{
		repository.DifficultyBeginner:     "初心者",
		repository.DifficultyIntermediate: "中級者",
		repository.DifficultyAdvanced:     "上級者",
		repository.DifficultyBeast:        "野獣級",
	}

	fmt.Printf("💪 新しいワークアウト「%s」を作成しました！難易度: %s\n", name, difficultyNames[workout.Difficulty])
	if workout.Weight > 0 {
		fmt.Printf("🔥 重量: %.1fkg × %dセット × %d回\n", workout.Weight, workout.Sets, workout.Reps)
	}

	return workout, nil
}

// GetWorkout ワークアウトを取得
func (wm *WorkoutManager) GetWorkout(id repository.WorkoutID) (*repository.Workout, error) {
	if wm.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	return wm.repo.GetWorkoutByID(id)
}

// UpdateWorkout ワークアウトを更新
func (wm *WorkoutManager) UpdateWorkout(id repository.WorkoutID, name, description string, status repository.WorkoutStatus, difficulty repository.Difficulty, muscleGroup repository.MuscleGroup, sets int, reps int, weight float64, notes string) error {
	if wm.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("workout name cannot be empty")
	}

	workout, err := wm.repo.GetWorkoutByID(id)
	if err != nil {
		return fmt.Errorf("failed to get workout: %v", err)
	}

	workout.Name = name
	workout.Description = description
	workout.Status = status
	workout.Difficulty = difficulty
	workout.MuscleGroup = muscleGroup
	workout.Sets = sets
	workout.Reps = reps
	workout.Weight = weight
	workout.Notes = notes
	workout.UpdatedAt = time.Now()

	// ステータスが完了に変更された場合
	if status == repository.WorkoutStatusCompleted && workout.CompletedAt == nil {
		now := time.Now()
		workout.CompletedAt = &now
		fmt.Printf("🎉 ワークアウト「%s」完了！お疲れ様でした！\n", name)
	}

	// ステータスがスキップに変更された場合
	if status == repository.WorkoutStatusSkipped {
		fmt.Printf("😅 ワークアウト「%s」をスキップしました。筋肉痛ですか？\n", name)
	}

	return wm.repo.UpdateWorkout(workout)
}

// DeleteWorkout ワークアウトを削除
func (wm *WorkoutManager) DeleteWorkout(id repository.WorkoutID) error {
	if wm.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	workout, err := wm.repo.GetWorkoutByID(id)
	if err != nil {
		return fmt.Errorf("failed to get workout: %v", err)
	}

	fmt.Printf("🗑️ ワークアウト「%s」を削除しました。\n", workout.Name)
	return wm.repo.DeleteWorkout(id)
}

// ListWorkouts ワークアウト一覧を取得
func (wm *WorkoutManager) ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*repository.Workout, error) {
	if wm.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	return wm.repo.ListWorkouts(statusFilter, difficultyFilter, muscleGroupFilter)
}

// GetHighIntensityWorkouts 高強度ワークアウトのみを取得（Go基礎技術使用例）
func (wm *WorkoutManager) GetHighIntensityWorkouts() ([]*repository.Workout, error) {
	if wm.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// 全ワークアウトを取得
	allWorkouts, err := wm.repo.ListWorkouts(nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get all workouts: %v", err)
	}

	// Go基礎技術1: ジェネリクス関数を使用してフィルタリング
	highIntensityWorkouts := utils.Filter(allWorkouts, func(w *repository.Workout) bool {
		isHighDifficulty := w.Difficulty == repository.DifficultyAdvanced || w.Difficulty == repository.DifficultyBeast
		isHeavyWeight := w.Weight >= 50.0
		return isHighDifficulty && isHeavyWeight
	})

	// Go基礎技術2: strings.Builder + 事前容量確保でログメッセージ構築
	logMessage := wm.buildHighIntensityLogMessage(len(allWorkouts), len(highIntensityWorkouts))
	fmt.Print(logMessage)

	return highIntensityWorkouts, nil
}

// buildHighIntensityLogMessage Go基礎技術による効率的なログメッセージ構築
func (wm *WorkoutManager) buildHighIntensityLogMessage(totalCount, filteredCount int) string {
	var builder strings.Builder
	// 概算容量を事前確保
	builder.Grow(100)

	builder.WriteString("🔥 高強度ワークアウト: 全")
	builder.WriteString(fmt.Sprintf("%d", totalCount))
	builder.WriteString("件中")
	builder.WriteString(fmt.Sprintf("%d", filteredCount))
	builder.WriteString("件を抽出しました")

	if filteredCount == 0 {
		builder.WriteString(" - もっと重いものを持ち上げましょう！💪")
	} else if filteredCount > totalCount/2 {
		builder.WriteString(" - 野獣レベルですね！🦍")
	}

	builder.WriteString("\n")
	return builder.String()
}

// GetWorkoutCount ワークアウト数を取得
func (wm *WorkoutManager) GetWorkoutCount() (int, error) {
	if wm.repo == nil {
		return 0, fmt.Errorf("repository not initialized")
	}

	return wm.repo.GetWorkoutCount()
}

// GetWorkoutStats ワークアウト統計を取得
func (wm *WorkoutManager) GetWorkoutStats(period string) (map[string]interface{}, error) {
	if wm.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	return wm.repo.GetWorkoutStats(period)
}

// 後方互換性のためのエイリアス
func NewTaskManagerWithRepository(repo repository.WorkoutRepository) *WorkoutManager {
	return NewWorkoutManagerWithRepository(repo)
}

func NewTaskManagerWithGORM(dsn string) (*WorkoutManager, error) {
	return NewWorkoutManagerWithGORM(dsn)
}
