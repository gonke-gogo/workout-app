package usecase

import (
	"fmt"
	"strings"
	"time"

	"golv2-learning-app/domain"
	appErrors "golv2-learning-app/errors"
)

// WorkoutManager ワークアウトのユースケース層（ビジネスロジック）
// WorkoutUseCaseインターフェースを実装
type WorkoutManager struct {
	repo domain.WorkoutRepository
}

// CreateWorkoutRequest ワークアウト作成リクエスト
// 全てのパラメータを1つの構造体にまとめることで、
// プレゼンテーション層からユースケース層への変換がシンプルになる
type CreateWorkoutRequest struct {
	ExerciseType domain.ExerciseType // 必須: ワークアウトの種目
	Description  string              // オプション
	Difficulty   domain.Difficulty   // オプション
	MuscleGroup  domain.MuscleGroup  // オプション
	Sets         int32               // オプション
	Reps         int32               // オプション
	Weight       float64             // オプション
	Notes        string              // オプション
}

// UpdateWorkoutRequest ワークアウト更新リクエスト
// CreateWorkoutRequestと同様に、全パラメータを構造体にまとめて可読性を向上
type UpdateWorkoutRequest struct {
	ID           domain.WorkoutID     // 必須: 更新対象のID
	ExerciseType domain.ExerciseType  // 必須: ワークアウトの種目
	Description  string               // オプション
	Difficulty   domain.Difficulty    // オプション
	MuscleGroup  domain.MuscleGroup   // オプション
	Status       domain.WorkoutStatus // オプション
	Sets         int                  // オプション
	Reps         int                  // オプション
	Weight       float64              // オプション
	Notes        string               // オプション
}

// ファクトリー関数
func NewWorkoutManager() *WorkoutManager {
	return &WorkoutManager{
		repo: nil, // メモリベース（後方互換性のため）
	}
}

// NewWorkoutManagerWithRepository リポジトリを使用するファクトリー関数
func NewWorkoutManagerWithRepository(repo domain.WorkoutRepository) *WorkoutManager {
	return &WorkoutManager{
		repo: repo,
	}
}

func (wm *WorkoutManager) CreateWorkout(req CreateWorkoutRequest) (*domain.Workout, error) {
	// defer でのログ記録とエラーハンドリング
	start := time.Now()
	fmt.Printf("🏃 ワークアウト作成開始: %s\n", req.ExerciseType.Japanese())

	defer func() {
		duration := time.Since(start)
		fmt.Printf("🏁 ワークアウト作成処理終了: %s (実行時間: %v)\n", req.ExerciseType.Japanese(), duration)
	}()

	// panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("💥 ワークアウト作成中にpanic発生: %s - %v\n", req.ExerciseType.Japanese(), r)
		}
	}()

	// ビジネスロジック: 入力値のバリデーション
	if req.ExerciseType == domain.ExerciseUnspecified {
		return nil, fmt.Errorf("exercise type must be specified")
	}

	// ビジネスロジック: デフォルト値の設定
	workout := &domain.Workout{
		ExerciseType: req.ExerciseType,
		Status:       domain.WorkoutStatusPlanned,
		Difficulty:   domain.DifficultyBeginner,
		Sets:         3,
		Reps:         10,
		Weight:       0.0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ビジネスロジック: リクエストの値を適用（0値でない場合のみ）
	if req.Description != "" {
		workout.Description = req.Description
	}
	if req.Difficulty != 0 {
		workout.Difficulty = req.Difficulty
	}
	if req.MuscleGroup != domain.Unspecified {
		workout.MuscleGroup = req.MuscleGroup
	}
	if req.Sets > 0 {
		workout.Sets = int(req.Sets)
	}
	if req.Reps > 0 {
		workout.Reps = int(req.Reps)
	}
	if req.Weight >= 0 {
		workout.Weight = req.Weight
	}
	if req.Notes != "" {
		workout.Notes = req.Notes
	}

	// ビジネスロジック: 最終的なバリデーション
	if err := wm.validateWorkoutData(workout); err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "CreateWorkout",
			ExerciseType: req.ExerciseType,
			Message:      "workout data validation failed",
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	err := wm.repo.CreateWorkout(workout)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "CreateWorkout",
			ExerciseType: req.ExerciseType,
			Message:      "failed to create workout in repository",
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	// ビジネスロジック: 作成成功ログ
	wm.logWorkoutCreated(workout)

	return workout, nil
}

// validateWorkoutData ワークアウトデータの妥当性チェック
func (wm *WorkoutManager) validateWorkoutData(workout *domain.Workout) error {
	if workout.Sets < 0 {
		return fmt.Errorf("sets cannot be negative: %d", workout.Sets)
	}
	if workout.Reps < 0 {
		return fmt.Errorf("reps cannot be negative: %d", workout.Reps)
	}
	if workout.Weight < 0 {
		return fmt.Errorf("weight cannot be negative: %.2f", workout.Weight)
	}
	return nil
}

// logWorkoutCreated ワークアウト作成時のログ出力
func (wm *WorkoutManager) logWorkoutCreated(workout *domain.Workout) {
	difficultyNames := map[domain.Difficulty]string{
		domain.DifficultyBeginner:     "初心者",
		domain.DifficultyIntermediate: "中級者",
		domain.DifficultyAdvanced:     "上級者",
		domain.DifficultyBeast:        "野獣級",
	}

	fmt.Printf("💪 新しいワークアウト「%s」を作成しました！難易度: %s\n", workout.ExerciseType.Japanese(), difficultyNames[workout.Difficulty])
	if workout.Weight > 0 {
		fmt.Printf("🔥 重量: %.1fkg × %dセット × %d回\n", workout.Weight, workout.Sets, workout.Reps)
	}
}

// GetWorkout ワークアウトを取得（ビジネスロジック層）
func (wm *WorkoutManager) GetWorkout(id domain.WorkoutID) (*domain.Workout, error) {
	// ビジネスロジック: 入力値のバリデーション
	if id <= 0 {
		workoutErr := &appErrors.WorkoutError{
			Op:      "GetWorkout",
			Message: fmt.Sprintf("workout ID must be positive (got: %d)", id),
			Err:     nil,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	workout, err := wm.repo.GetWorkoutByID(id)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:      "GetWorkout",
			Message: fmt.Sprintf("failed to retrieve workout from repository (ID: %d)", id),
			Err:     err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	// ビジネスロジック: 取得したデータの妥当性チェック
	if !wm.isValidWorkout(workout) {
		workoutErr := &appErrors.WorkoutError{
			Op:           "GetWorkout",
			ExerciseType: workout.ExerciseType,
			Message:      fmt.Sprintf("workout data validation failed after retrieval (ID: %d)", id),
			Err:          nil,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	return workout, nil
}

// UpdateWorkout ワークアウトを更新（ビジネスロジック層）
func (wm *WorkoutManager) UpdateWorkout(req UpdateWorkoutRequest) error {
	// ビジネスロジック: 入力値のバリデーション
	if err := wm.validateUpdateInput(req.ID, req.ExerciseType, req.Sets, req.Reps, req.Weight); err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "UpdateWorkout",
			ExerciseType: req.ExerciseType,
			Message:      fmt.Sprintf("update input validation failed (ID: %d)", req.ID),
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	// 既存のワークアウトを取得
	workout, err := wm.repo.GetWorkoutByID(req.ID)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "UpdateWorkout",
			ExerciseType: req.ExerciseType,
			Message:      fmt.Sprintf("failed to get workout for update (ID: %d)", req.ID),
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	// ビジネスロジック: 値の更新
	workout.ExerciseType = req.ExerciseType
	workout.Description = req.Description
	workout.Status = req.Status
	workout.Difficulty = req.Difficulty
	workout.MuscleGroup = req.MuscleGroup
	workout.Sets = req.Sets
	workout.Reps = req.Reps
	workout.Weight = req.Weight
	workout.Notes = req.Notes
	workout.UpdatedAt = time.Now()

	// ビジネスロジック: ステータス変更時の処理
	wm.handleStatusChange(workout, req.Status, req.ExerciseType)

	err = wm.repo.UpdateWorkout(workout)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "UpdateWorkout",
			ExerciseType: req.ExerciseType,
			Message:      fmt.Sprintf("failed to persist workout update (ID: %d)", req.ID),
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	fmt.Printf("✅ ワークアウト「%s」を更新しました\n", req.ExerciseType.Japanese())
	return nil
}

// validateUpdateInput 更新時の入力値バリデーション
func (wm *WorkoutManager) validateUpdateInput(id domain.WorkoutID, exerciseType domain.ExerciseType, sets, reps int, weight float64) error {
	if id <= 0 {
		return fmt.Errorf("invalid workout ID: %d", id)
	}
	if exerciseType == domain.ExerciseUnspecified {
		return fmt.Errorf("exercise type must be specified")
	}
	if sets < 0 {
		return fmt.Errorf("sets cannot be negative: %d", sets)
	}
	if reps < 0 {
		return fmt.Errorf("reps cannot be negative: %d", reps)
	}
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative: %.2f", weight)
	}
	return nil
}

// handleStatusChange ステータス変更時のビジネスロジック
func (wm *WorkoutManager) handleStatusChange(workout *domain.Workout, newStatus domain.WorkoutStatus, exerciseType domain.ExerciseType) {
	// ステータスが完了に変更された場合
	if newStatus == domain.WorkoutStatusCompleted && workout.CompletedAt == nil {
		now := time.Now()
		workout.CompletedAt = &now
		fmt.Printf("🎉 ワークアウト「%s」完了！お疲れ様でした！\n", exerciseType.Japanese())
	}

	// ステータスがスキップに変更された場合
	if newStatus == domain.WorkoutStatusSkipped {
		fmt.Printf("😅 ワークアウト「%s」をスキップしました。筋肉痛ですか？\n", exerciseType.Japanese())
	}
}

// DeleteWorkout ワークアウトを削除（ビジネスロジック層）
func (wm *WorkoutManager) DeleteWorkout(id domain.WorkoutID) error {
	// ビジネスロジック: 入力値のバリデーション
	if id <= 0 {
		workoutErr := &appErrors.WorkoutError{
			Op:      "DeleteWorkout",
			Message: fmt.Sprintf("workout ID must be positive (got: %d)", id),
			Err:     nil,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	// ビジネスロジック: 削除前に存在確認
	workout, err := wm.repo.GetWorkoutByID(id)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:      "DeleteWorkout",
			Message: fmt.Sprintf("failed to get workout before deletion (ID: %d)", id),
			Err:     err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	// ビジネスロジック: 完了済みワークアウトの削除警告
	if workout.Status == domain.WorkoutStatusCompleted {
		fmt.Printf("⚠️  完了済みのワークアウトを削除します: 「%s」\n", workout.ExerciseType.Japanese())
	}

	err = wm.repo.DeleteWorkout(id)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:           "DeleteWorkout",
			ExerciseType: workout.ExerciseType,
			Message:      fmt.Sprintf("failed to delete workout from repository (ID: %d)", id),
			Err:          err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return workoutErr
	}

	fmt.Printf("🗑️ ワークアウト「%s」を削除しました\n", workout.ExerciseType.Japanese())
	return nil
}

// ListWorkouts ワークアウト一覧を取得（ビジネスロジック層）
func (wm *WorkoutManager) ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*domain.Workout, error) {
	// リポジトリから全データを取得
	workouts, err := wm.repo.ListWorkouts(statusFilter, difficultyFilter, muscleGroupFilter)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:      "ListWorkouts",
			Message: "failed to retrieve workouts from repository",
			Err:     err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	// ビジネスロジック: 無効なデータをフィルタリング
	validWorkouts := make([]*domain.Workout, 0, len(workouts))
	for _, workout := range workouts {
		if wm.isValidWorkout(workout) {
			validWorkouts = append(validWorkouts, workout)
		}
	}

	fmt.Printf("🔍 フィルタリング結果: 全%d件中、有効なワークアウト%d件を返します\n", len(workouts), len(validWorkouts))
	return validWorkouts, nil
}

// isValidWorkout ビジネスルール: ワークアウトの妥当性チェック
func (wm *WorkoutManager) isValidWorkout(workout *domain.Workout) bool {
	// 必須項目のチェック
	if workout == nil {
		return false
	}
	if workout.ExerciseType == domain.ExerciseUnspecified {
		return false
	}
	if workout.ID <= 0 {
		return false
	}
	return true
}

// GetHighIntensityWorkouts 高強度ワークアウトのみを取得（Go基礎技術使用例）
func (wm *WorkoutManager) GetHighIntensityWorkouts() ([]*domain.Workout, error) {
	// 全ワークアウトを取得
	allWorkouts, err := wm.repo.ListWorkouts(nil, nil, nil)
	if err != nil {
		workoutErr := &appErrors.WorkoutError{
			Op:      "GetHighIntensityWorkouts",
			Message: "failed to get all workouts for filtering",
			Err:     err,
		}
		fmt.Printf("❌ %s\n", workoutErr.Error())
		return nil, workoutErr
	}

	// Go基礎技術1: 効率的なフィルタリング処理
	highIntensityWorkouts := make([]*domain.Workout, 0)
	for _, w := range allWorkouts {
		isHighDifficulty := w.Difficulty == domain.DifficultyAdvanced || w.Difficulty == domain.DifficultyBeast
		isHeavyWeight := w.Weight >= 50.0
		if isHighDifficulty && isHeavyWeight {
			highIntensityWorkouts = append(highIntensityWorkouts, w)
		}
	}

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
	return wm.repo.GetWorkoutCount()
}

// GetWorkoutStats ワークアウト統計を取得
func (wm *WorkoutManager) GetWorkoutStats(period string) (map[string]interface{}, error) {
	return wm.repo.GetWorkoutStats(period)
}

// 後方互換性のためのエイリアス
func NewTaskManagerWithRepository(repo domain.WorkoutRepository) *WorkoutManager {
	return NewWorkoutManagerWithRepository(repo)
}
