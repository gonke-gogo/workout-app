package usecase

import (
	"golv2-learning-app/domain"
	"testing"
)

// TestCreateWorkout_WithMock_TableDriven テーブル駆動テストでワークアウト作成をテスト
// モックリポジトリを使用して、DBなしでビジネスロジックをテスト
func TestCreateWorkout_WithMock_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateWorkoutRequest
		wantErr     bool
		wantID      domain.WorkoutID
		wantStatus  domain.WorkoutStatus
		wantSets    int
		wantReps    int
		description string
	}{
		{
			name: "正常系: ベンチプレス作成",
			request: CreateWorkoutRequest{
				ExerciseType: domain.BenchPress,
				Description:  "テスト用ベンチプレス",
				Difficulty:   domain.DifficultyBeginner,
				MuscleGroup:  domain.Chest,
				Sets:         3,
				Reps:         10,
				Weight:       60.0,
			},
			wantErr:     false,
			wantID:      1,
			wantStatus:  domain.WorkoutStatusPlanned,
			wantSets:    3,
			wantReps:    10,
			description: "通常のベンチプレス作成",
		},
		{
			name: "正常系: デフォルト値適用（Sets/Repsなし）",
			request: CreateWorkoutRequest{
				ExerciseType: domain.Squat,
				// Sets, Repsを指定しない → デフォルト値が適用される
			},
			wantErr:     false,
			wantID:      1,
			wantStatus:  domain.WorkoutStatusPlanned,
			wantSets:    3,  // デフォルト値
			wantReps:    10, // デフォルト値
			description: "Sets/Reps未指定時のデフォルト値適用を確認",
		},
		{
			name: "正常系: 高難易度デッドリフト",
			request: CreateWorkoutRequest{
				ExerciseType: domain.Deadlift,
				Difficulty:   domain.DifficultyBeast,
				MuscleGroup:  domain.Back,
				Sets:         5,
				Reps:         3,
				Weight:       150.0,
				Notes:        "ヘビーセット",
			},
			wantErr:     false,
			wantID:      1,
			wantStatus:  domain.WorkoutStatusPlanned,
			wantSets:    5,
			wantReps:    3,
			description: "高難易度・高重量のワークアウト",
		},
		{
			name: "異常系: ExerciseType未指定",
			request: CreateWorkoutRequest{
				ExerciseType: domain.ExerciseUnspecified,
				Sets:         3,
				Reps:         10,
			},
			wantErr:     true,
			description: "ExerciseTypeが未指定の場合、エラーを返す",
		},
		{
			name: "正常系: 負の値は無視される（デフォルト値が適用）",
			request: CreateWorkoutRequest{
				ExerciseType: domain.BenchPress,
				Sets:         -5,    // 負の値 → 無視される
				Reps:         -10,   // 負の値 → 無視される
				Weight:       -50.0, // 負の値 → 無視される
			},
			wantErr:     false,
			wantID:      1,
			wantStatus:  domain.WorkoutStatusPlanned,
			wantSets:    3,  // デフォルト値が適用
			wantReps:    10, // デフォルト値が適用
			description: "負の値は無視され、デフォルト値が適用されることを確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックリポジトリを作成（各テストケースで独立）
			mockRepo := NewMockWorkoutRepository()
			manager := NewWorkoutManagerWithRepository(mockRepo)

			// テスト実行
			workout, err := manager.CreateWorkout(tt.request)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWorkout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// エラーケースの場合はここで終了
			if tt.wantErr {
				return
			}

			// 正常系の詳細チェック
			if workout.ID != tt.wantID {
				t.Errorf("Expected ID=%d, got ID=%d", tt.wantID, workout.ID)
			}

			if workout.ExerciseType != tt.request.ExerciseType {
				t.Errorf("Expected ExerciseType=%v, got %v", tt.request.ExerciseType, workout.ExerciseType)
			}

			if workout.Status != tt.wantStatus {
				t.Errorf("Expected Status=%v, got %v", tt.wantStatus, workout.Status)
			}

			if workout.Sets != tt.wantSets {
				t.Errorf("Expected Sets=%d, got %d", tt.wantSets, workout.Sets)
			}

			if workout.Reps != tt.wantReps {
				t.Errorf("Expected Reps=%d, got %d", tt.wantReps, workout.Reps)
			}
		})
	}
}

// TestUpdateWorkout_WithMock_TableDriven テーブル駆動テストでワークアウト更新をテスト
func TestUpdateWorkout_WithMock_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		setupWorkout *domain.Workout // 事前に作成するワークアウト
		updateID     domain.WorkoutID
		updateType   domain.ExerciseType
		updateSets   int
		updateReps   int
		updateWeight float64
		wantErr      bool
		description  string
	}{
		{
			name: "正常系: セット数・レップ数の更新",
			setupWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.BenchPress,
				Status:       domain.WorkoutStatusPlanned,
				Sets:         3,
				Reps:         10,
				Weight:       60.0,
			},
			updateID:     1,
			updateType:   domain.BenchPress,
			updateSets:   5,    // 3 → 5に変更
			updateReps:   8,    // 10 → 8に変更
			updateWeight: 70.0, // 60 → 70に変更
			wantErr:      false,
			description:  "既存ワークアウトの数値を更新",
		},
		{
			name: "正常系: 重量のみ更新",
			setupWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.Squat,
				Status:       domain.WorkoutStatusPlanned,
				Sets:         4,
				Reps:         10,
				Weight:       80.0,
			},
			updateID:     1,
			updateType:   domain.Squat,
			updateSets:   4,    // 変更なし
			updateReps:   10,   // 変更なし
			updateWeight: 85.0, // 80 → 85に変更
			wantErr:      false,
			description:  "重量のみを更新",
		},
		{
			name:         "異常系: 存在しないID",
			setupWorkout: nil, // データなし
			updateID:     999,
			updateType:   domain.BenchPress,
			updateSets:   3,
			updateReps:   10,
			updateWeight: 60.0,
			wantErr:      true,
			description:  "存在しないIDを指定した場合、エラーを返す",
		},
		{
			name: "異常系: 負のセット数",
			setupWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.BenchPress,
				Status:       domain.WorkoutStatusPlanned,
				Sets:         3,
				Reps:         10,
			},
			updateID:     1,
			updateType:   domain.BenchPress,
			updateSets:   -5, // 負の値
			updateReps:   10,
			updateWeight: 60.0,
			wantErr:      true,
			description:  "負のセット数は許可されない",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockWorkoutRepository()
			manager := NewWorkoutManagerWithRepository(mockRepo)

			// 事前データ作成
			if tt.setupWorkout != nil {
				err := mockRepo.CreateWorkout(tt.setupWorkout)
				if err != nil {
					t.Fatalf("Failed to setup workout: %v", err)
				}
			}

			// テスト実行
			err := manager.UpdateWorkout(UpdateWorkoutRequest{
				ID:           tt.updateID,
				ExerciseType: tt.updateType,
				Sets:         tt.updateSets,
				Reps:         tt.updateReps,
				Weight:       tt.updateWeight,
				Notes:        "",
				Description:  "",
				Status:       domain.WorkoutStatusPlanned,
				Difficulty:   domain.DifficultyIntermediate,
				MuscleGroup:  domain.Chest,
			})

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateWorkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 正常系の場合、更新が反映されているか確認
			if !tt.wantErr {
				updated, err := mockRepo.GetWorkoutByID(tt.updateID)
				if err != nil {
					t.Fatalf("Failed to get updated workout: %v", err)
				}

				if updated.Sets != tt.updateSets {
					t.Errorf("Expected Sets=%d, got %d", tt.updateSets, updated.Sets)
				}
				if updated.Reps != tt.updateReps {
					t.Errorf("Expected Reps=%d, got %d", tt.updateReps, updated.Reps)
				}
				if updated.Weight != tt.updateWeight {
					t.Errorf("Expected Weight=%.1f, got %.1f", tt.updateWeight, updated.Weight)
				}
			}
		})
	}
}

// TestDeleteWorkout_WithMock_TableDriven テーブル駆動テストでワークアウト削除をテスト
func TestDeleteWorkout_WithMock_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		setupWorkout *domain.Workout
		deleteID     domain.WorkoutID
		wantErr      bool
		description  string
	}{
		{
			name: "正常系: 予定中のワークアウト削除",
			setupWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.BenchPress,
				Status:       domain.WorkoutStatusPlanned,
			},
			deleteID:    1,
			wantErr:     false,
			description: "予定中のワークアウトを削除",
		},
		{
			name: "正常系: 完了済みワークアウト削除（警告あり）",
			setupWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.Squat,
				Status:       domain.WorkoutStatusCompleted,
			},
			deleteID:    1,
			wantErr:     false,
			description: "完了済みワークアウトも削除可能（警告ログが出る）",
		},
		{
			name:         "異常系: 存在しないID",
			setupWorkout: nil,
			deleteID:     999,
			wantErr:      true,
			description:  "存在しないIDを指定した場合、エラーを返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockWorkoutRepository()
			manager := NewWorkoutManagerWithRepository(mockRepo)

			// 事前データ作成
			if tt.setupWorkout != nil {
				err := mockRepo.CreateWorkout(tt.setupWorkout)
				if err != nil {
					t.Fatalf("Failed to setup workout: %v", err)
				}
			}

			// テスト実行
			err := manager.DeleteWorkout(tt.deleteID)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteWorkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 正常系の場合、削除されていることを確認
			if !tt.wantErr {
				_, err := mockRepo.GetWorkoutByID(tt.deleteID)
				if err == nil {
					t.Error("Expected workout to be deleted, but it still exists")
				}
			}
		})
	}
}

// TestListWorkouts_WithMock_TableDriven テーブル駆動テストでワークアウト一覧取得をテスト
func TestListWorkouts_WithMock_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		setupWorkouts     []*domain.Workout
		statusFilter      *int
		difficultyFilter  *int
		muscleGroupFilter *int
		wantCount         int
		wantErr           bool
		description       string
	}{
		{
			name: "正常系: フィルタなし（全件取得）",
			setupWorkouts: []*domain.Workout{
				{ExerciseType: domain.BenchPress, Status: domain.WorkoutStatusPlanned},
				{ExerciseType: domain.Squat, Status: domain.WorkoutStatusCompleted},
				{ExerciseType: domain.Deadlift, Status: domain.WorkoutStatusSkipped},
			},
			statusFilter:      nil,
			difficultyFilter:  nil,
			muscleGroupFilter: nil,
			wantCount:         3,
			wantErr:           false,
			description:       "全てのワークアウトが取得される",
		},
		{
			name: "正常系: ステータスフィルタ（完了済みのみ）",
			setupWorkouts: []*domain.Workout{
				{ExerciseType: domain.BenchPress, Status: domain.WorkoutStatusPlanned},
				{ExerciseType: domain.Squat, Status: domain.WorkoutStatusCompleted},
				{ExerciseType: domain.Deadlift, Status: domain.WorkoutStatusCompleted},
			},
			statusFilter:      intPtr(int(domain.WorkoutStatusCompleted)),
			difficultyFilter:  nil,
			muscleGroupFilter: nil,
			wantCount:         2,
			wantErr:           false,
			description:       "完了済みのワークアウトのみ取得",
		},
		{
			name: "正常系: 難易度フィルタ（野獣級のみ）",
			setupWorkouts: []*domain.Workout{
				{ExerciseType: domain.BenchPress, Difficulty: domain.DifficultyBeginner},
				{ExerciseType: domain.Deadlift, Difficulty: domain.DifficultyBeast},
			},
			statusFilter:      nil,
			difficultyFilter:  intPtr(int(domain.DifficultyBeast)),
			muscleGroupFilter: nil,
			wantCount:         1,
			wantErr:           false,
			description:       "野獣級のワークアウトのみ取得",
		},
		{
			name: "正常系: 部位フィルタ（胸のみ）",
			setupWorkouts: []*domain.Workout{
				{ExerciseType: domain.BenchPress, MuscleGroup: domain.Chest},
				{ExerciseType: domain.Squat, MuscleGroup: domain.Legs},
				{ExerciseType: domain.PullUp, MuscleGroup: domain.Back},
			},
			statusFilter:      nil,
			difficultyFilter:  nil,
			muscleGroupFilter: intPtr(int(domain.Chest)),
			wantCount:         1,
			wantErr:           false,
			description:       "胸のワークアウトのみ取得",
		},
		{
			name:              "正常系: データなし",
			setupWorkouts:     []*domain.Workout{},
			statusFilter:      nil,
			difficultyFilter:  nil,
			muscleGroupFilter: nil,
			wantCount:         0,
			wantErr:           false,
			description:       "データがない場合、空配列が返る",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockWorkoutRepository()
			manager := NewWorkoutManagerWithRepository(mockRepo)

			// 事前データ作成
			for _, workout := range tt.setupWorkouts {
				err := mockRepo.CreateWorkout(workout)
				if err != nil {
					t.Fatalf("Failed to setup workout: %v", err)
				}
			}

			// テスト実行
			workouts, err := manager.ListWorkouts(tt.statusFilter, tt.difficultyFilter, tt.muscleGroupFilter)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ListWorkouts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 件数チェック
			if len(workouts) != tt.wantCount {
				t.Errorf("Expected %d workouts, got %d", tt.wantCount, len(workouts))
			}
		})
	}
}

// intPtr int値のポインタを返すヘルパー関数
func intPtr(i int) *int {
	return &i
}

// 以下、既存のシンプルなテストも残す（互換性のため）

// TestCreateWorkout_WithMock 基本的なワークアウト作成テスト（後方互換性のため残す）
func TestCreateWorkout_WithMock(t *testing.T) {
	mockRepo := NewMockWorkoutRepository()
	manager := NewWorkoutManagerWithRepository(mockRepo)

	workout, err := manager.CreateWorkout(CreateWorkoutRequest{
		ExerciseType: domain.BenchPress,
	})
	if err != nil {
		t.Fatalf("CreateWorkout failed: %v", err)
	}

	if workout.ID == 0 {
		t.Error("Expected workout ID to be set, got 0")
	}
	if workout.ExerciseType != domain.BenchPress {
		t.Errorf("Expected ExerciseType %s, got %s", domain.BenchPress.Japanese(), workout.ExerciseType.Japanese())
	}
}
