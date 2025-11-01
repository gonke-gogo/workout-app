package server

import (
	"fmt"
	"strings"
	"testing"

	"golv2-learning-app/domain"
)

// テスト用のワークアウトデータを生成
func generateTestWorkouts(count int) []*domain.Workout {
	workouts := make([]*domain.Workout, count)
	for i := 0; i < count; i++ {
		workouts[i] = &domain.Workout{
			ID:           domain.WorkoutID(i + 1),
			ExerciseType: domain.BenchPress,
			MuscleGroup:  domain.Chest,
			Sets:         3,
			Reps:         10,
			Weight:       60.0 + float64(i),
		}
	}
	return workouts
}

// バッドパターン: + 演算子を使った文字列結合
func buildWorkoutSummaryBad(workouts []*domain.Workout) string {
	if len(workouts) == 0 {
		return "📋 ワークアウトがありません"
	}

	var result string

	result += "📋 ワークアウト一覧 ("
	result += fmt.Sprintf("%d件", len(workouts))
	result += "):\n"

	for i, workout := range workouts {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("  %d. %s (%s)", i+1, workout.ExerciseType.Japanese(), workout.MuscleGroup.Japanese())

		if workout.Sets > 0 && workout.Reps > 0 {
			result += fmt.Sprintf(" - %dセット×%d回", workout.Sets, workout.Reps)
		}

		if workout.Weight > 0 {
			result += fmt.Sprintf(" @ %.1fkg", workout.Weight)
		}
	}

	return result
}

// グッドパターン: strings.Builderを使った文字列結合（現在の実装）
func buildWorkoutSummaryGood(workouts []*domain.Workout) string {
	if len(workouts) == 0 {
		return "📋 ワークアウトがありません"
	}

	var builder strings.Builder
	estimatedSize := len(workouts)*50 + 100
	builder.Grow(estimatedSize)

	builder.WriteString("📋 ワークアウト一覧 (")
	builder.WriteString(fmt.Sprintf("%d件", len(workouts)))
	builder.WriteString("):\n")

	for i, workout := range workouts {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("  %d. %s (%s)", i+1, workout.ExerciseType.Japanese(), workout.MuscleGroup.Japanese()))

		if workout.Sets > 0 && workout.Reps > 0 {
			builder.WriteString(fmt.Sprintf(" - %dセット×%d回", workout.Sets, workout.Reps))
		}

		if workout.Weight > 0 {
			builder.WriteString(fmt.Sprintf(" @ %.1fkg", workout.Weight))
		}
	}

	return builder.String()
}

// Bad += 演算子を使った文字列結合
func BenchmarkBuildWorkoutSummary_Bad(b *testing.B) {
	workouts := generateTestWorkouts(100) // 100個のワークアウト
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = buildWorkoutSummaryBad(workouts)
	}
}

// Good strings.Builderを使った文字列結合
func BenchmarkBuildWorkoutSummary_Good(b *testing.B) {
	workouts := generateTestWorkouts(100) // 100個のワークアウト
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = buildWorkoutSummaryGood(workouts)
	}
}
