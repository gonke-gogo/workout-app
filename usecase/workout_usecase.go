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
	ExerciseType domain.ExerciseType
	Description  string
	Difficulty   domain.Difficulty
	MuscleGroup  domain.MuscleGroup
	Sets         int32
	Reps         int32
	Weight       float64
	Notes        string
}

// UpdateWorkoutRequest ワークアウト更新リクエスト
// ポインタ型を使用して「更新しない」と「明示的な値」を区別
type UpdateWorkoutRequest struct {
	ID           domain.WorkoutID      // 必須: 更新対象のID
	ExerciseType domain.ExerciseType   // 必須: ワークアウトの種目
	Description  *string               // オプション: nilなら更新しない
	Difficulty   *domain.Difficulty    // オプション: nilなら更新しない
	MuscleGroup  *domain.MuscleGroup   // オプション: nilなら更新しない
	Status       *domain.WorkoutStatus // オプション: nilなら更新しない
	Sets         *int                  // オプション: nilなら更新しない
	Reps         *int                  // オプション: nilなら更新しない
	Weight       *float64              // オプション: nilなら更新しない
	Notes        *string               // オプション: nilなら更新しない
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
	if req.Weight > 0 {
		workout.Weight = req.Weight
	}
	if req.Notes != "" {
		workout.Notes = req.Notes
	}

	// ビジネスロジック: 最終的なバリデーション
	// errValidatorを使用したバリデーション（冗長的なエラーチェックをまとめる）
	if err := wm.validateWorkoutDataWithErrValidator(workout); err != nil {
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

// ValidationErrors 複数のバリデーションエラーを保持するカスタムエラー型
// 全てのバリデーションエラーを収集して返すため
type ValidationErrors struct {
	Errors []error
}

// 注意: error()メソッドで len(ev.errors) > 0 の場合のみ ValidationErrors が作成されるため、
// len(ve.Errors) == 0 の場合は通常発生しない（防御的プログラミングのためのチェック）
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		// 通常は発生しないが、防御的プログラミングとしてチェック
		return "validation errors: no errors found (internal error)"
	}
	if len(ve.Errors) == 1 {
		return ve.Errors[0].Error()
	}
	// 複数のエラーがある場合: strings.Builderを使用した効率的な文字列結合
	var builder strings.Builder
	// 概算容量を事前確保（各エラーメッセージを平均50文字と仮定）
	estimatedSize := len(ve.Errors) * 50
	builder.Grow(estimatedSize)

	for i, err := range ve.Errors {
		if i > 0 {
			builder.WriteString("; ")
		}
		builder.WriteString(err.Error())
	}
	return builder.String()
}

// errValidator 冗長的なエラーチェックをまとめるための構造体
// 「複数のエラーをまとめる方法」の評価項目に対応
// errWriterパターンを参考に、複数のバリデーション処理を連続して行い、
// 全てのバリデーションを実行し、全てのエラーを収集して返す
type errValidator struct {
	errors []error // 内部でエラーを保持（スライスで複数保持）
}

// validate バリデーション関数を実行し、エラーが発生した場合は内部に保持
// 全てのバリデーションを実行するため、エラーがあっても処理を継続
func (ev *errValidator) validate(fn func() error) {
	err := fn()
	if err != nil {
		ev.errors = append(ev.errors, err)
	}
}

// validateSets Setsの値を検証
func (ev *errValidator) validateSets(sets int) {
	ev.validate(func() error {
		if sets < 0 {
			return fmt.Errorf("sets cannot be negative: %d", sets)
		}
		return nil
	})
}

// validateReps Repsの値を検証
func (ev *errValidator) validateReps(reps int) {
	ev.validate(func() error {
		if reps < 0 {
			return fmt.Errorf("reps cannot be negative: %d", reps)
		}
		return nil
	})
}

// validateWeight Weightの値を検証
func (ev *errValidator) validateWeight(weight float64) {
	ev.validate(func() error {
		if weight < 0 {
			return fmt.Errorf("weight cannot be negative: %.2f", weight)
		}
		return nil
	})
}

// validateID IDの値を検証
func (ev *errValidator) validateID(id domain.WorkoutID) {
	ev.validate(func() error {
		if id <= 0 {
			return fmt.Errorf("invalid workout ID: %d", id)
		}
		return nil
	})
}

// validateExerciseType ExerciseTypeの値を検証
func (ev *errValidator) validateExerciseType(exerciseType domain.ExerciseType) {
	ev.validate(func() error {
		if exerciseType == domain.ExerciseUnspecified {
			return fmt.Errorf("exercise type must be specified")
		}
		return nil
	})
}

// error 保持しているエラーを返す（nilの場合はnilを返す）
// 複数のエラーがある場合は ValidationErrors として返す
func (ev *errValidator) error() error {
	if len(ev.errors) == 0 {
		return nil
	}
	if len(ev.errors) == 1 {
		return ev.errors[0]
	}
	// 複数のエラーがある場合は ValidationErrors として返す
	return &ValidationErrors{Errors: ev.errors}
}

// validateWorkoutDataWithErrValidator errValidatorを使用したバリデーション
// 冗長的なエラーチェックを構造体にまとめることで、呼び出し元のコードがシンプルになる
// 全てのバリデーションを実行し、全てのエラーを収集して返す
func (wm *WorkoutManager) validateWorkoutDataWithErrValidator(workout *domain.Workout) error {
	validator := &errValidator{}

	// 複数のバリデーションを全て実行（エラーがあっても継続）
	// 全てのエラーを収集するため、スキップしない
	validator.validateSets(workout.Sets)
	validator.validateReps(workout.Reps)
	validator.validateWeight(workout.Weight)

	// 最終的なエラーチェック（複数のエラーがある場合はまとめて返す）
	return validator.error()
}

// logWorkoutCreated ワークアウト作成時のログ出力
func (wm *WorkoutManager) logWorkoutCreated(workout *domain.Workout) {
	difficultyNames := map[domain.Difficulty]string{
		domain.DifficultyBeginner:     "初心者",
		domain.DifficultyIntermediate: "中級者",
		domain.DifficultyAdvanced:     "上級者",
		domain.DifficultyBeast:        "お前は化け物だ、キモいです、すごいです",
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

	workout, err := wm.repo.GetWorkout(id)
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
	workout, err := wm.repo.GetWorkout(req.ID)
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

	// ビジネスロジック: 値の更新（nilでないフィールドのみ）
	workout.ExerciseType = req.ExerciseType

	if req.Description != nil {
		workout.Description = *req.Description
	}
	if req.Status != nil {
		workout.Status = *req.Status
		// ビジネスロジック: ステータス変更時の処理
		wm.handleStatusChange(workout, *req.Status, req.ExerciseType)
	}
	if req.Difficulty != nil {
		workout.Difficulty = *req.Difficulty
	}
	if req.MuscleGroup != nil {
		workout.MuscleGroup = *req.MuscleGroup
	}
	if req.Sets != nil {
		workout.Sets = *req.Sets
	}
	if req.Reps != nil {
		workout.Reps = *req.Reps
	}
	if req.Weight != nil {
		workout.Weight = *req.Weight
	}
	if req.Notes != nil {
		workout.Notes = *req.Notes
	}
	workout.UpdatedAt = time.Now()

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
func (wm *WorkoutManager) validateUpdateInput(id domain.WorkoutID, exerciseType domain.ExerciseType, sets, reps *int, weight *float64) error {
	if id <= 0 {
		return fmt.Errorf("invalid workout ID: %d", id)
	}
	if exerciseType == domain.ExerciseUnspecified {
		return fmt.Errorf("exercise type must be specified")
	}
	if sets != nil && *sets < 0 {
		return fmt.Errorf("sets cannot be negative: %d", *sets)
	}
	if reps != nil && *reps < 0 {
		return fmt.Errorf("reps cannot be negative: %d", *reps)
	}
	if weight != nil && *weight < 0 {
		return fmt.Errorf("weight cannot be negative: %.2f", *weight)
	}
	return nil
}

// validateUpdateInputWithErrValidator errValidatorを使用した更新時のバリデーション
// 冗長的なエラーチェックを構造体にまとめることで、呼び出し元のコードがシンプルになる
func (wm *WorkoutManager) validateUpdateInputWithErrValidator(id domain.WorkoutID, exerciseType domain.ExerciseType, sets, reps *int, weight *float64) error {
	validator := &errValidator{}

	// 複数のバリデーションを連続して実行
	// 1つでもエラーが発生した場合、それ以降の処理はスキップされる
	validator.validateID(id)
	validator.validateExerciseType(exerciseType)

	if sets != nil {
		validator.validateSets(*sets)
	}
	if reps != nil {
		validator.validateReps(*reps)
	}
	if weight != nil {
		validator.validateWeight(*weight)
	}

	// 最終的なエラーチェックは1回だけ
	return validator.error()
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
	workout, err := wm.repo.GetWorkout(id)
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

	highIntensityWorkouts := make([]*domain.Workout, 0)
	for _, w := range allWorkouts {
		isHighDifficulty := w.Difficulty == domain.DifficultyAdvanced || w.Difficulty == domain.DifficultyBeast
		isHeavyWeight := w.Weight >= 50.0
		if isHighDifficulty && isHeavyWeight {
			highIntensityWorkouts = append(highIntensityWorkouts, w)
		}
	}

	logMessage := wm.buildHighIntensityLogMessage(len(allWorkouts), len(highIntensityWorkouts))
	fmt.Print(logMessage)

	// --- Generics活用例: 任意の数値条件で件数を集計してログに出す ---
	// Weightが80.5以上の件数
	heavyCount := countWorkoutsBy[float64](allWorkouts,
		func(w *domain.Workout) float64 { return w.Weight },
		func(v float64) bool { return v >= 80.5 },
	)
	// Setsが5以上の件数
	highSetsCount := countWorkoutsBy[int](allWorkouts,
		func(w *domain.Workout) int { return w.Sets },
		func(v int) bool { return v >= 5 },
	)
	fmt.Printf("🔎 Generics filter summary: weight>=80.5=%d, sets>=5=%d\n", heavyCount, highSetsCount)

	return highIntensityWorkouts, nil
}

// ジェネリクス関数用
type IntOrFloat interface {
	int | float64
}

func countWorkoutsBy[T IntOrFloat](workouts []*domain.Workout, selector func(*domain.Workout) T, filter func(T) bool) int {
	if len(workouts) == 0 {
		return 0
	}
	var count int
	for _, w := range workouts {
		value := selector(w)
		if filter(value) {
			count++
		}
	}
	return count
}

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
		builder.WriteString(" - もっと重いものを持ち上げましょう！💪こんなもんじゃないだろう！")
	} else if filteredCount > totalCount/2 {
		builder.WriteString(" - 野獣レベルですね！強すぎ🦍")
	}

	builder.WriteString("\n")
	return builder.String()
}

// GetWorkoutCount ワークアウト数を取得
func (wm *WorkoutManager) GetWorkoutCount() (int, error) {
	return wm.repo.GetWorkoutCount()
}

// 後方互換性のためのエイリアス
func NewTaskManagerWithRepository(repo domain.WorkoutRepository) *WorkoutManager {
	return NewWorkoutManagerWithRepository(repo)
}
