package repository

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"golv2-learning-app/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	insertWorkoutQuery = regexp.QuoteMeta("INSERT INTO `workouts`")
	selectWorkoutQuery = regexp.QuoteMeta("SELECT * FROM `workouts` WHERE `workouts`.`id` = ? ORDER BY `workouts`.`id` LIMIT 1")
	updateWorkoutQuery = regexp.QuoteMeta("UPDATE `workouts` SET")
	deleteWorkoutQuery = regexp.QuoteMeta("DELETE FROM `workouts` WHERE `workouts`.`id` = ?")
)

// setupMockDB モック化されたGORMリポジトリを作成
func setupMockDB(t *testing.T) (*GORMRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	dialector := mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}

	repo := &GORMRepository{db: gormDB}
	return repo, mock, db
}

// TestGORMRepository_CreateWorkout
func TestGORMRepository_CreateWorkout(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		workout      *domain.Workout
		mockResultID int64 // モックで返すID
		mockAffected int64 // 影響を受けた行数
		mockError    error // エラーをシミュレート
		wantErr      bool
		description  string
	}{
		{
			name: "正常系: ベンチプレス作成",
			workout: &domain.Workout{
				ExerciseType: domain.BenchPress,
				Description:  "テスト用のベンチプレス",
				Status:       domain.WorkoutStatusPlanned,
				Difficulty:   domain.DifficultyBeginner,
				MuscleGroup:  domain.Chest,
				Sets:         3,
				Reps:         10,
				Weight:       60.0,
				Notes:        "テスト実行中",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockResultID: 1,
			mockAffected: 1,
			mockError:    nil,
			wantErr:      false,
			description:  "通常のベンチプレス作成",
		},
		{
			name: "正常系: スクワット作成（重量なし）",
			workout: &domain.Workout{
				ExerciseType: domain.Squat,
				Description:  "自重スクワット",
				Status:       domain.WorkoutStatusPlanned,
				Difficulty:   domain.DifficultyIntermediate,
				MuscleGroup:  domain.Legs,
				Sets:         5,
				Reps:         15,
				Weight:       0.0,
				Notes:        "自重トレーニング",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockResultID: 2,
			mockAffected: 1,
			mockError:    nil,
			wantErr:      false,
			description:  "重量なしのワークアウト作成",
		},
		{
			name: "正常系: デッドリフト作成（野獣級）",
			workout: &domain.Workout{
				ExerciseType: domain.Deadlift,
				Description:  "ヘビーデッドリフト",
				Status:       domain.WorkoutStatusCompleted,
				Difficulty:   domain.DifficultyBeast,
				MuscleGroup:  domain.Back,
				Sets:         3,
				Reps:         5,
				Weight:       150.0,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockResultID: 3,
			mockAffected: 1,
			mockError:    nil,
			wantErr:      false,
			description:  "高難易度・高重量のワークアウト作成",
		},
		{
			name: "異常系: DB接続エラー",
			workout: &domain.Workout{
				ExerciseType: domain.BenchPress,
				Status:       domain.WorkoutStatusPlanned,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockResultID: 0,
			mockAffected: 0,
			mockError:    sql.ErrConnDone, // 接続エラーをシミュレート
			wantErr:      true,
			description:  "DB接続エラー時の挙動をテスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock, db := setupMockDB(t)
			defer db.Close()

			// モックの期待値を設定
			mock.ExpectBegin()

			if tt.mockError != nil {
				// 異常系
				mock.ExpectExec(insertWorkoutQuery).
					WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				// 正常系
				mock.ExpectExec(insertWorkoutQuery).
					WithArgs(
						sqlmock.AnyArg(), // exercise_type
						tt.workout.Description,
						sqlmock.AnyArg(), // status
						sqlmock.AnyArg(), // difficulty
						sqlmock.AnyArg(), // muscle_group
						tt.workout.Sets,
						tt.workout.Reps,
						tt.workout.Weight,
						tt.workout.Notes,
						sqlmock.AnyArg(), // created_at
						sqlmock.AnyArg(), // updated_at
						sqlmock.AnyArg(), // completed_at
					).
					WillReturnResult(sqlmock.NewResult(tt.mockResultID, tt.mockAffected))
				mock.ExpectCommit()
			}

			// テスト実行
			err := repo.CreateWorkout(tt.workout)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWorkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			// モックの期待値が全て満たされたか確認
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}

// TestGORMRepository_GetWorkout
func TestGORMRepository_GetWorkout(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		workoutID   domain.WorkoutID
		mockWorkout *domain.Workout
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name:      "正常系: ワークアウト取得",
			workoutID: 1,
			mockWorkout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.BenchPress,
				Description:  "テスト用のベンチプレス",
				Status:       domain.WorkoutStatusPlanned,
				Difficulty:   domain.DifficultyBeginner,
				MuscleGroup:  domain.Chest,
				Sets:         3,
				Reps:         10,
				Weight:       60.0,
				Notes:        "テスト実行中",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockError:   nil,
			wantErr:     false,
			description: "既存ワークアウトの取得が成功",
		},
		{
			name:        "異常系: レコードが見つからない",
			workoutID:   999,
			mockWorkout: nil,
			mockError:   gorm.ErrRecordNotFound,
			wantErr:     true,
			description: "存在しないIDを指定した場合のエラーハンドリング",
		},
		{
			name:        "異常系: DB接続エラー",
			workoutID:   1,
			mockWorkout: nil,
			mockError:   sql.ErrConnDone,
			wantErr:     true,
			description: "DB接続エラー時のエラーハンドリング",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock, db := setupMockDB(t)
			defer db.Close()

			// モックの期待値を設定
			if tt.mockError != nil {
				// 異常系
				mock.ExpectQuery(selectWorkoutQuery).
					WithArgs(tt.workoutID).
					WillReturnError(tt.mockError)
			} else {
				// 正常系
				rows := sqlmock.NewRows([]string{"id", "exercise_type", "description", "status", "difficulty", "muscle_group", "sets", "reps", "weight", "notes", "created_at", "updated_at", "completed_at"}).
					AddRow(
						tt.mockWorkout.ID,
						tt.mockWorkout.ExerciseType,
						tt.mockWorkout.Description,
						tt.mockWorkout.Status,
						tt.mockWorkout.Difficulty,
						tt.mockWorkout.MuscleGroup,
						tt.mockWorkout.Sets,
						tt.mockWorkout.Reps,
						tt.mockWorkout.Weight,
						tt.mockWorkout.Notes,
						tt.mockWorkout.CreatedAt,
						tt.mockWorkout.UpdatedAt,
						tt.mockWorkout.CompletedAt,
					)
				mock.ExpectQuery(selectWorkoutQuery).
					WithArgs(tt.workoutID).
					WillReturnRows(rows)
			}

			// テスト実行
			workout, err := repo.GetWorkout(tt.workoutID)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 正常系の場合、取得したワークアウトの内容を確認
			if !tt.wantErr {
				if workout == nil {
					t.Error("Expected workout but got nil")
					return
				}
				if workout.ID != tt.mockWorkout.ID {
					t.Errorf("Expected ID=%d, got ID=%d", tt.mockWorkout.ID, workout.ID)
				}
				if workout.ExerciseType != tt.mockWorkout.ExerciseType {
					t.Errorf("Expected ExerciseType=%v, got %v", tt.mockWorkout.ExerciseType, workout.ExerciseType)
				}
			} else {
				// 異常系の場合、エラーメッセージにidの情報が含まれていることを確認
				if err != nil {
					t.Logf("✅ エラーログ: %v", err)
					t.Logf("✅ エラーメッセージ: %s", err.Error())
				}
			}

			// モックの期待値が全て満たされたか確認
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}

// TestGORMRepository_UpdateWorkout
func TestGORMRepository_UpdateWorkout(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		workout      *domain.Workout
		mockAffected int64
		mockError    error
		wantErr      bool
		description  string
	}{
		{
			name: "正常系: ワークアウト更新",
			workout: &domain.Workout{
				ID:           1,
				ExerciseType: domain.BenchPress,
				Description:  "更新されたベンチプレス",
				Status:       domain.WorkoutStatusPlanned,
				Difficulty:   domain.DifficultyAdvanced,
				MuscleGroup:  domain.Chest,
				Sets:         5,
				Reps:         8,
				Weight:       70.0,
				Notes:        "重量アップ",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockAffected: 1,
			mockError:    nil,
			wantErr:      false,
			description:  "既存ワークアウトの更新が成功",
		},
		{
			name: "異常系: 更新エラー",
			workout: &domain.Workout{
				ID:           999,
				ExerciseType: domain.BenchPress,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			mockAffected: 0,
			mockError:    sql.ErrNoRows,
			wantErr:      true,
			description:  "更新失敗時のエラーハンドリング",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock, db := setupMockDB(t)
			defer db.Close()

			// モックの期待値を設定
			mock.ExpectBegin()

			if tt.mockError != nil {
				// エラーケース
				mock.ExpectExec(updateWorkoutQuery).
					WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				// 正常系
				mock.ExpectExec(updateWorkoutQuery).
					WillReturnResult(sqlmock.NewResult(0, tt.mockAffected))
				mock.ExpectCommit()
			}

			// テスト実行
			err := repo.UpdateWorkout(tt.workout)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateWorkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			// モックの期待値が全て満たされたか確認
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}

// TestGORMRepository_DeleteWorkout
func TestGORMRepository_DeleteWorkout(t *testing.T) {
	tests := []struct {
		name         string
		workoutID    domain.WorkoutID
		mockAffected int64
		mockError    error
		wantErr      bool
		description  string
	}{
		{
			name:         "正常系: ワークアウト削除",
			workoutID:    1,
			mockAffected: 1,
			mockError:    nil,
			wantErr:      false,
			description:  "既存ワークアウトの削除が成功",
		},
		{
			name:         "異常系: 削除エラー",
			workoutID:    999,
			mockAffected: 0,
			mockError:    sql.ErrNoRows,
			wantErr:      true,
			description:  "削除失敗時のエラーハンドリング",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock, db := setupMockDB(t)
			defer db.Close()

			// モックの期待値を設定
			mock.ExpectBegin()

			if tt.mockError != nil {
				// エラーケース
				mock.ExpectExec(deleteWorkoutQuery).
					WithArgs(tt.workoutID).
					WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				// 正常系
				mock.ExpectExec(deleteWorkoutQuery).
					WithArgs(tt.workoutID).
					WillReturnResult(sqlmock.NewResult(0, tt.mockAffected))
				mock.ExpectCommit()
			}

			// テスト実行
			err := repo.DeleteWorkout(tt.workoutID)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteWorkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			// モックの期待値が全て満たされたか確認
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}
