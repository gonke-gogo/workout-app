package repository

import (
	"fmt"

	"golv2-learning-app/domain"
)

// MockWorkoutRepository テスト用のモック実装
type MockWorkoutRepository struct {
	workouts map[domain.WorkoutID]*domain.Workout
	nextID   domain.WorkoutID
}

// NewMockWorkoutRepository 新しいモックリポジトリを作成
func NewMockWorkoutRepository() *MockWorkoutRepository {
	return &MockWorkoutRepository{
		workouts: make(map[domain.WorkoutID]*domain.Workout),
		nextID:   1,
	}
}

// CreateWorkout ワークアウトを作成（メモリ上）
func (m *MockWorkoutRepository) CreateWorkout(workout *domain.Workout) error {
	workout.ID = m.nextID
	m.workouts[m.nextID] = workout
	m.nextID++
	return nil
}

// GetWorkout ワークアウトをIDで取得
func (m *MockWorkoutRepository) GetWorkout(id domain.WorkoutID) (*domain.Workout, error) {
	workout, exists := m.workouts[id]
	if !exists {
		return nil, fmt.Errorf("workout not found: id=%d", id)
	}
	return workout, nil
}

// UpdateWorkout ワークアウトを更新
func (m *MockWorkoutRepository) UpdateWorkout(workout *domain.Workout) error {
	if _, exists := m.workouts[workout.ID]; !exists {
		return fmt.Errorf("workout not found: id=%d", workout.ID)
	}
	m.workouts[workout.ID] = workout
	return nil
}

// DeleteWorkout ワークアウトを削除
func (m *MockWorkoutRepository) DeleteWorkout(id domain.WorkoutID) error {
	if _, exists := m.workouts[id]; !exists {
		return fmt.Errorf("workout not found: id=%d", id)
	}
	delete(m.workouts, id)
	return nil
}

// ListWorkouts ワークアウト一覧を取得
func (m *MockWorkoutRepository) ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*domain.Workout, error) {
	result := make([]*domain.Workout, 0, len(m.workouts))
	for _, workout := range m.workouts {
		// フィルタリング処理（簡易版）
		if statusFilter != nil && int(workout.Status) != *statusFilter {
			continue
		}
		if difficultyFilter != nil && int(workout.Difficulty) != *difficultyFilter {
			continue
		}
		if muscleGroupFilter != nil && int(workout.MuscleGroup) != *muscleGroupFilter {
			continue
		}
		result = append(result, workout)
	}
	return result, nil
}

// GetWorkoutCount ワークアウト数を取得
func (m *MockWorkoutRepository) GetWorkoutCount() (int, error) {
	return len(m.workouts), nil
}
