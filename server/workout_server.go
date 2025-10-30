package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"golv2-learning-app/domain"
	"golv2-learning-app/proto"
	"golv2-learning-app/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer gRPCサーバー構造体
type GRPCServer struct {
	proto.UnimplementedWorkoutServiceServer
	workoutManager *usecase.WorkoutManager
}

// NewGRPCServer 新しいgRPCサーバーを作成
func NewGRPCServer(workoutManager *usecase.WorkoutManager) *GRPCServer {
	return &GRPCServer{
		workoutManager: workoutManager,
	}
}

// Start サーバーを起動
func (s *GRPCServer) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterWorkoutServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("💪 筋トレアプリのgRPCサーバーが起動しました！")
	log.Printf("🔥 ポート %d でリッスン中...", port)
	log.Printf("🎯 Evansで接続: evans -r repl -p %d", port)

	return grpcServer.Serve(lis)
}

// CreateWorkout ワークアウトを作成（プレゼンテーション層）
func (s *GRPCServer) CreateWorkout(ctx context.Context, req *proto.CreateWorkoutRequest) (*proto.CreateWorkoutResponse, error) {
	exerciseType := convertProtoExerciseType(req.ExerciseType)
	log.Printf("💪 新しいワークアウトを作成中: %s", exerciseType.Japanese())

	// proto → usecase.CreateWorkoutRequest への変換
	usecaseReq := usecase.CreateWorkoutRequest{
		ExerciseType: exerciseType,
		Description:  req.Description,
		Difficulty:   convertProtoDifficulty(req.Difficulty),
		MuscleGroup:  convertProtoMuscleGroup(req.MuscleGroup),
		Sets:         req.Sets,
		Reps:         req.Reps,
		Weight:       req.Weight,
		Notes:        req.Notes,
	}

	// ビジネスロジック層に処理を委譲
	workout, err := s.workoutManager.CreateWorkout(usecaseReq)
	if err != nil {
		return &proto.CreateWorkoutResponse{
			Workout: nil,
			Message: s.buildErrorMessage("ワークアウト作成", exerciseType.Japanese(), err.Error()),
		}, nil
	}

	// domain → proto への変換（プレゼンテーション層の責務）
	protoWorkout := convertToProtoWorkout(workout)
	return &proto.CreateWorkoutResponse{
		Workout: protoWorkout,
		Message: s.buildSuccessMessage("作成", exerciseType.Japanese()),
	}, nil
}

// GetWorkout ワークアウトを取得（プレゼンテーション層）
func (s *GRPCServer) GetWorkout(ctx context.Context, req *proto.GetWorkoutRequest) (*proto.GetWorkoutResponse, error) {
	log.Printf("🔍 ワークアウトを取得中: ID %d", req.Id)

	// ビジネスロジック層に処理を委譲
	workout, err := s.workoutManager.GetWorkout(domain.WorkoutID(req.Id))
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %v", err)
	}

	// domain → proto への変換（プレゼンテーション層の責務）
	protoWorkout := convertToProtoWorkout(workout)
	return &proto.GetWorkoutResponse{
		Workout: protoWorkout,
	}, nil
}

// UpdateWorkout ワークアウトを更新（プレゼンテーション層）
func (s *GRPCServer) UpdateWorkout(ctx context.Context, req *proto.UpdateWorkoutRequest) (*proto.UpdateWorkoutResponse, error) {
	// proto → domain への変換
	exerciseType := convertProtoExerciseType(req.ExerciseType)
	log.Printf("✏️ ワークアウトを更新中: ID %d (%s)", req.Id, exerciseType.Japanese())

	// proto → usecase.UpdateWorkoutRequest への変換
	usecaseReq := usecase.UpdateWorkoutRequest{
		ID:           domain.WorkoutID(req.Id),
		ExerciseType: exerciseType,
		Description:  req.Description,
		Difficulty:   convertProtoDifficulty(req.Difficulty),
		MuscleGroup:  convertProtoMuscleGroup(req.MuscleGroup),
		Status:       convertProtoWorkoutStatus(req.Status),
		Sets:         int(req.Sets),
		Reps:         int(req.Reps),
		Weight:       req.Weight,
		Notes:        req.Notes,
	}

	// ビジネスロジック層に処理を委譲
	err := s.workoutManager.UpdateWorkout(usecaseReq)
	if err != nil {
		return &proto.UpdateWorkoutResponse{
			Workout: nil,
			Message: fmt.Sprintf("❌ ワークアウト更新に失敗しました: %v", err),
		}, nil
	}

	// 更新されたワークアウトを取得（表示用）
	workout, err := s.workoutManager.GetWorkout(domain.WorkoutID(req.Id))
	if err != nil {
		return &proto.UpdateWorkoutResponse{
			Workout: nil,
			Message: fmt.Sprintf("❌ 更新されたワークアウトの取得に失敗しました: %v", err),
		}, nil
	}

	// domain → proto への変換（プレゼンテーション層の責務）
	protoWorkout := convertToProtoWorkout(workout)
	return &proto.UpdateWorkoutResponse{
		Workout: protoWorkout,
		Message: fmt.Sprintf("✅ ワークアウト「%s」が更新されました！", exerciseType.Japanese()),
	}, nil
}

// DeleteWorkout ワークアウトを削除（プレゼンテーション層）
func (s *GRPCServer) DeleteWorkout(ctx context.Context, req *proto.DeleteWorkoutRequest) (*proto.DeleteWorkoutResponse, error) {
	log.Printf("🗑️ ワークアウトを削除中: ID %d", req.Id)

	// ビジネスロジック層に処理を委譲
	err := s.workoutManager.DeleteWorkout(domain.WorkoutID(req.Id))
	if err != nil {
		return &proto.DeleteWorkoutResponse{
			Message: fmt.Sprintf("❌ ワークアウト削除に失敗しました: %v", err),
		}, nil
	}

	return &proto.DeleteWorkoutResponse{
		Message: "✅ ワークアウトが削除されました！",
	}, nil
}

// ListWorkouts ワークアウト一覧を取得（プレゼンテーション層）
func (s *GRPCServer) ListWorkouts(ctx context.Context, req *proto.ListWorkoutsRequest) (*proto.ListWorkoutsResponse, error) {
	log.Printf("📋 ワークアウト一覧を取得中...")

	// フィルター条件の変換（proto → domain）
	var statusFilter *int
	var difficultyFilter *int
	var muscleGroupFilter *int

	if req.StatusFilter != proto.WorkoutStatus_WORKOUT_STATUS_UNSPECIFIED {
		status := int(convertProtoWorkoutStatus(req.StatusFilter))
		statusFilter = &status
	}
	if req.DifficultyFilter != proto.Difficulty_DIFFICULTY_UNSPECIFIED {
		difficulty := int(convertProtoDifficulty(req.DifficultyFilter))
		difficultyFilter = &difficulty
	}
	if req.MuscleGroupFilter != proto.MuscleGroup_UNSPECIFIED {
		muscleGroup := int(convertProtoMuscleGroup(req.MuscleGroupFilter))
		muscleGroupFilter = &muscleGroup
	}

	workouts, err := s.workoutManager.ListWorkouts(statusFilter, difficultyFilter, muscleGroupFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %v", err)
	}

	convertedWorkouts := make([]*proto.Workout, 0, len(workouts))
	for _, workout := range workouts {
		convertedWorkouts = append(convertedWorkouts, convertToProtoWorkout(workout))
	}

	// サマリーメッセージを生成
	summary := s.buildWorkoutSummary(workouts)

	log.Printf("✅ %d件のワークアウトを返却します", len(convertedWorkouts))

	return &proto.ListWorkoutsResponse{
		Workouts:   convertedWorkouts,
		TotalCount: int32(len(convertedWorkouts)),
		Message:    summary,
	}, nil
}

// buildErrorMessage Go基礎技術による効率的なエラーメッセージ構築
func (s *GRPCServer) buildErrorMessage(operation, target, errorDetail string) string {
	var builder strings.Builder
	// 概算容量を事前確保
	builder.Grow(len(operation) + len(target) + len(errorDetail) + 20)

	builder.WriteString("❌ ")
	builder.WriteString(operation)
	builder.WriteString("エラー: ")
	builder.WriteString(target)
	builder.WriteString(" - ")
	builder.WriteString(errorDetail)

	return builder.String()
}

// buildSuccessMessage Go基礎技術による効率的な成功メッセージ構築
func (s *GRPCServer) buildSuccessMessage(operation, target string) string {
	var builder strings.Builder
	// 概算容量を事前確保
	builder.Grow(len(operation) + len(target) + 30)

	builder.WriteString("🎉 ワークアウト「")
	builder.WriteString(target)
	builder.WriteString("」が")
	builder.WriteString(operation)
	builder.WriteString("されました！")

	return builder.String()
}

// buildWorkoutSummary Go基礎技術による効率的なワークアウトサマリー構築
func (s *GRPCServer) buildWorkoutSummary(workouts []*domain.Workout) string {
	if len(workouts) == 0 {
		return "📋 ワークアウトがありません"
	}

	var builder strings.Builder
	// 概算容量を計算（各ワークアウト約50文字と仮定）
	estimatedSize := len(workouts)*50 + 100
	builder.Grow(estimatedSize)

	builder.WriteString("📋 ワークアウト一覧 (")
	builder.WriteString(fmt.Sprintf("%d件", len(workouts)))
	builder.WriteString("):\n")

	// Go基礎技術: 効率的なループ処理
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

// GetWorkoutStats ワークアウト統計を取得
func (s *GRPCServer) GetWorkoutStats(ctx context.Context, req *proto.GetWorkoutStatsRequest) (*proto.GetWorkoutStatsResponse, error) {
	log.Printf("📊 ワークアウト統計を取得中: 期間 %s", req.Period)

	stats, err := s.workoutManager.GetWorkoutStats(req.Period)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout stats: %v", err)
	}

	response := &proto.GetWorkoutStatsResponse{
		TotalWorkouts:     int32(stats["total_workouts"].(int)),
		CompletedWorkouts: int32(stats["completed_workouts"].(int)),
		SkippedWorkouts:   int32(stats["skipped_workouts"].(int)),
		TotalWeightLifted: stats["total_weight_lifted"].(float64),
		MuscleGroupStats:  make(map[string]int32),
		Message:           fmt.Sprintf("📈 %sの統計データです！", req.Period),
	}

	// 筋肉群別統計を変換
	if muscleGroupStats, ok := stats["muscle_group_stats"].(map[string]int); ok {
		for group, count := range muscleGroupStats {
			response.MuscleGroupStats[group] = int32(count)
		}
	}

	return response, nil
}

// GetHighIntensityWorkouts 高強度ワークアウト一覧を取得（ジェネリクス使用例）
func (s *GRPCServer) GetHighIntensityWorkouts(ctx context.Context, req *proto.GetHighIntensityWorkoutsRequest) (*proto.GetHighIntensityWorkoutsResponse, error) {
	log.Printf("🔥 高強度ワークアウトを取得中...")

	// ビジネスロジック層で高強度ワークアウトを取得（ジェネリクス関数使用）
	workouts, err := s.workoutManager.GetHighIntensityWorkouts()
	if err != nil {
		return nil, fmt.Errorf("failed to get high intensity workouts: %v", err)
	}

	// プロトコル形式に変換
	protoWorkouts := make([]*proto.Workout, len(workouts))
	for i, workout := range workouts {
		protoWorkouts[i] = convertToProtoWorkout(workout)
	}

	message := fmt.Sprintf("💪 高強度ワークアウト%d件を取得しました！野獣たちの記録です", len(workouts))
	if len(workouts) == 0 {
		message = "😅 高強度ワークアウトがありません。もっと重いものを持ち上げましょう！"
	}

	return &proto.GetHighIntensityWorkoutsResponse{
		Workouts:   protoWorkouts,
		TotalCount: int32(len(workouts)),
		Message:    message,
	}, nil
}

// 変換関数
func convertToProtoWorkout(workout *domain.Workout) *proto.Workout {
	protoWorkout := &proto.Workout{
		Id:           int32(workout.ID),
		ExerciseType: convertToProtoExerciseType(workout.ExerciseType),
		Description:  workout.Description,
		Status:       convertToProtoWorkoutStatus(workout.Status),
		Difficulty:   convertToProtoDifficulty(workout.Difficulty),
		MuscleGroup:  convertToProtoMuscleGroup(workout.MuscleGroup),
		Sets:         int32(workout.Sets),
		Reps:         int32(workout.Reps),
		Weight:       workout.Weight,
		Notes:        workout.Notes,
		CreatedAt:    workout.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    workout.UpdatedAt.Format(time.RFC3339),
	}

	if workout.CompletedAt != nil {
		protoWorkout.CompletedAt = workout.CompletedAt.Format(time.RFC3339)
	}

	return protoWorkout
}

func convertToProtoWorkoutStatus(status domain.WorkoutStatus) proto.WorkoutStatus {
	switch status {
	case domain.WorkoutStatusPlanned:
		return proto.WorkoutStatus_WORKOUT_STATUS_PLANNED
	case domain.WorkoutStatusInProgress:
		return proto.WorkoutStatus_WORKOUT_STATUS_IN_PROGRESS
	case domain.WorkoutStatusCompleted:
		return proto.WorkoutStatus_WORKOUT_STATUS_COMPLETED
	case domain.WorkoutStatusSkipped:
		return proto.WorkoutStatus_WORKOUT_STATUS_SKIPPED
	default:
		return proto.WorkoutStatus_WORKOUT_STATUS_UNSPECIFIED
	}
}

func convertProtoWorkoutStatus(status proto.WorkoutStatus) domain.WorkoutStatus {
	switch status {
	case proto.WorkoutStatus_WORKOUT_STATUS_PLANNED:
		return domain.WorkoutStatusPlanned
	case proto.WorkoutStatus_WORKOUT_STATUS_IN_PROGRESS:
		return domain.WorkoutStatusInProgress
	case proto.WorkoutStatus_WORKOUT_STATUS_COMPLETED:
		return domain.WorkoutStatusCompleted
	case proto.WorkoutStatus_WORKOUT_STATUS_SKIPPED:
		return domain.WorkoutStatusSkipped
	default:
		return domain.WorkoutStatusPlanned
	}
}

func convertToProtoDifficulty(difficulty domain.Difficulty) proto.Difficulty {
	switch difficulty {
	case domain.DifficultyBeginner:
		return proto.Difficulty_DIFFICULTY_BEGINNER
	case domain.DifficultyIntermediate:
		return proto.Difficulty_DIFFICULTY_INTERMEDIATE
	case domain.DifficultyAdvanced:
		return proto.Difficulty_DIFFICULTY_ADVANCED
	case domain.DifficultyBeast:
		return proto.Difficulty_DIFFICULTY_BEAST
	default:
		return proto.Difficulty_DIFFICULTY_UNSPECIFIED
	}
}

func convertToProtoMuscleGroup(muscleGroup domain.MuscleGroup) proto.MuscleGroup {
	switch muscleGroup {
	case domain.Chest:
		return proto.MuscleGroup_CHEST
	case domain.Back:
		return proto.MuscleGroup_BACK
	case domain.Legs:
		return proto.MuscleGroup_LEGS
	case domain.Shoulders:
		return proto.MuscleGroup_SHOULDERS
	case domain.Arms:
		return proto.MuscleGroup_ARMS
	case domain.Abs:
		return proto.MuscleGroup_ABS
	case domain.Core:
		return proto.MuscleGroup_CORE
	case domain.Glutes:
		return proto.MuscleGroup_GLUTES
	case domain.Cardio:
		return proto.MuscleGroup_CARDIO
	case domain.FullBody:
		return proto.MuscleGroup_FULL_BODY
	default:
		return proto.MuscleGroup_UNSPECIFIED
	}
}

func convertProtoDifficulty(difficulty proto.Difficulty) domain.Difficulty {
	switch difficulty {
	case proto.Difficulty_DIFFICULTY_BEGINNER:
		return domain.DifficultyBeginner
	case proto.Difficulty_DIFFICULTY_INTERMEDIATE:
		return domain.DifficultyIntermediate
	case proto.Difficulty_DIFFICULTY_ADVANCED:
		return domain.DifficultyAdvanced
	case proto.Difficulty_DIFFICULTY_BEAST:
		return domain.DifficultyBeast
	default:
		return domain.DifficultyBeginner
	}
}

func convertProtoMuscleGroup(muscleGroup proto.MuscleGroup) domain.MuscleGroup {
	switch muscleGroup {
	case proto.MuscleGroup_CHEST:
		return domain.Chest
	case proto.MuscleGroup_BACK:
		return domain.Back
	case proto.MuscleGroup_LEGS:
		return domain.Legs
	case proto.MuscleGroup_SHOULDERS:
		return domain.Shoulders
	case proto.MuscleGroup_ARMS:
		return domain.Arms
	case proto.MuscleGroup_ABS:
		return domain.Abs
	case proto.MuscleGroup_CORE:
		return domain.Core
	case proto.MuscleGroup_GLUTES:
		return domain.Glutes
	case proto.MuscleGroup_CARDIO:
		return domain.Cardio
	case proto.MuscleGroup_FULL_BODY:
		return domain.FullBody
	default:
		return domain.Unspecified
	}
}

// ExerciseType変換関数（domain → proto）
func convertToProtoExerciseType(exerciseType domain.ExerciseType) proto.ExerciseType {
	switch exerciseType {
	case domain.BenchPress:
		return proto.ExerciseType_EXERCISE_BENCH_PRESS
	case domain.Squat:
		return proto.ExerciseType_EXERCISE_SQUAT
	case domain.Deadlift:
		return proto.ExerciseType_EXERCISE_DEADLIFT
	case domain.DumbbellShoulder:
		return proto.ExerciseType_EXERCISE_DUMBBELL_SHOULDER
	case domain.PullUp:
		return proto.ExerciseType_EXERCISE_PULL_UP
	case domain.SideRaise:
		return proto.ExerciseType_EXERCISE_SIDE_RAISE
	case domain.OneHandRow:
		return proto.ExerciseType_EXERCISE_ONE_HAND_ROW
	case domain.HighPull:
		return proto.ExerciseType_EXERCISE_HIGH_PULL
	default:
		return proto.ExerciseType_EXERCISE_UNSPECIFIED
	}
}

// ExerciseType変換関数（proto → domain）
func convertProtoExerciseType(exerciseType proto.ExerciseType) domain.ExerciseType {
	switch exerciseType {
	case proto.ExerciseType_EXERCISE_BENCH_PRESS:
		return domain.BenchPress
	case proto.ExerciseType_EXERCISE_SQUAT:
		return domain.Squat
	case proto.ExerciseType_EXERCISE_DEADLIFT:
		return domain.Deadlift
	case proto.ExerciseType_EXERCISE_DUMBBELL_SHOULDER:
		return domain.DumbbellShoulder
	case proto.ExerciseType_EXERCISE_PULL_UP:
		return domain.PullUp
	case proto.ExerciseType_EXERCISE_SIDE_RAISE:
		return domain.SideRaise
	case proto.ExerciseType_EXERCISE_ONE_HAND_ROW:
		return domain.OneHandRow
	case proto.ExerciseType_EXERCISE_HIGH_PULL:
		return domain.HighPull
	default:
		return domain.ExerciseUnspecified
	}
}
