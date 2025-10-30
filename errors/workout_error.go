package errors

import (
	"fmt"

	"golv2-learning-app/domain"
)

// WorkoutError カスタムエラー型：構造化された詳細情報を保持
type WorkoutError struct {
	Op           string              // 操作名（"CreateWorkout", "UpdateWorkout"など）
	ExerciseType domain.ExerciseType // エクササイズタイプ
	Message      string              // 人間が読める説明メッセージ
	Err          error               // 元のエラー（wrap用）
}

// Error エラーメッセージを返す（errorインターフェースの実装）
func (e *WorkoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("workout error: op=%s, type=%s, message=%s: %v",
			e.Op, e.ExerciseType.Japanese(), e.Message, e.Err)
	}
	return fmt.Sprintf("workout error: op=%s, type=%s, message=%s",
		e.Op, e.ExerciseType.Japanese(), e.Message)
}

// Unwrap 元のエラーを返す（errors.Unwrapで使用）
func (e *WorkoutError) Unwrap() error {
	return e.Err
}
