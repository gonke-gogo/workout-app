package repository

import (
	"testing"
	"time"
)

func TestGORMRepository_CreateWorkout(t *testing.T) {
	// テスト用のDSN（実際のテストではテスト用DBを使用）
	dsn := "workoutuser:workoutpass@tcp(localhost:3306)/workoutdb_test?parseTime=true&charset=utf8mb4"

	repo, err := NewGORMRepository(dsn)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to database: %v", err)
	}
	defer repo.Close()

	workout := &Workout{
		Name:        "テストベンチプレス",
		Description: "テスト用のワークアウト",
		Status:      WorkoutStatusPlanned,
		Difficulty:  DifficultyBeginner,
		MuscleGroup: Chest,
		Sets:        3,
		Reps:        10,
		Weight:      60.0,
		Notes:       "テスト実行中",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.CreateWorkout(workout)
	if err != nil {
		t.Errorf("CreateWorkout failed: %v", err)
	}

	if workout.ID == 0 {
		t.Error("Expected workout ID to be set")
	}
}

func TestGORMRepository_GetWorkoutByID(t *testing.T) {
	dsn := "workoutuser:workoutpass@tcp(localhost:3306)/workoutdb_test?parseTime=true&charset=utf8mb4"

	repo, err := NewGORMRepository(dsn)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to database: %v", err)
	}
	defer repo.Close()

	// ワークアウトを作成
	workout := &Workout{
		Name:        "テストスクワット",
		Description: "テスト用のワークアウト取得",
		Status:      WorkoutStatusPlanned,
		Difficulty:  DifficultyIntermediate,
		MuscleGroup: Legs,
		Sets:        4,
		Reps:        15,
		Weight:      0.0,
		Notes:       "自重トレーニング",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.CreateWorkout(workout)
	if err != nil {
		t.Fatalf("CreateWorkout failed: %v", err)
	}

	// ワークアウトを取得
	retrievedWorkout, err := repo.GetWorkoutByID(workout.ID)
	if err != nil {
		t.Errorf("GetWorkoutByID failed: %v", err)
	}

	if retrievedWorkout.Name != workout.Name {
		t.Errorf("Expected name %s, got %s", workout.Name, retrievedWorkout.Name)
	}

	if retrievedWorkout.MuscleGroup != workout.MuscleGroup {
		t.Errorf("Expected muscle group %s, got %s", workout.MuscleGroup, retrievedWorkout.MuscleGroup)
	}
}

func TestGORMRepository_ListWorkouts(t *testing.T) {
	dsn := "workoutuser:workoutpass@tcp(localhost:3306)/workoutdb_test?parseTime=true&charset=utf8mb4"

	repo, err := NewGORMRepository(dsn)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to database: %v", err)
	}
	defer repo.Close()

	// テスト用ワークアウトを複数作成
	workouts := []*Workout{
		{
			Name:        "テストベンチプレス",
			Status:      WorkoutStatusPlanned,
			Difficulty:  DifficultyAdvanced,
			MuscleGroup: Chest,
			Sets:        3,
			Reps:        10,
			Weight:      80.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "テストデッドリフト",
			Status:      WorkoutStatusCompleted,
			Difficulty:  DifficultyBeast,
			MuscleGroup: Back,
			Sets:        3,
			Reps:        8,
			Weight:      100.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// ワークアウトを作成
	for _, workout := range workouts {
		err = repo.CreateWorkout(workout)
		if err != nil {
			t.Fatalf("CreateWorkout failed: %v", err)
		}
	}

	// フィルタなしでワークアウト一覧を取得
	allWorkouts, err := repo.ListWorkouts(nil, nil, nil)
	if err != nil {
		t.Errorf("ListWorkouts failed: %v", err)
	}

	if len(allWorkouts) < 2 {
		t.Errorf("Expected at least 2 workouts, got %d", len(allWorkouts))
	}

	// ステータスフィルタでワークアウト一覧を取得
	statusFilter := int(WorkoutStatusCompleted)
	completedWorkouts, err := repo.ListWorkouts(&statusFilter, nil, nil)
	if err != nil {
		t.Errorf("ListWorkouts with status filter failed: %v", err)
	}

	// 完了済みワークアウトが含まれているか確認
	found := false
	for _, workout := range completedWorkouts {
		if workout.Status == WorkoutStatusCompleted {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find completed workouts")
	}
}
