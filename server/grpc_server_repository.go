package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"golv2-learning-app/proto"
	"golv2-learning-app/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer gRPCサーバー構造体
type GRPCServer struct {
	proto.UnimplementedWorkoutServiceServer
	workoutManager *WorkoutManager
}

// NewGRPCServer 新しいgRPCサーバーを作成
func NewGRPCServer(workoutManager *WorkoutManager) *GRPCServer {
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

// CreateWorkout ワークアウトを作成（Go基礎技術による最適化）
func (s *GRPCServer) CreateWorkout(ctx context.Context, req *proto.CreateWorkoutRequest) (*proto.CreateWorkoutResponse, error) {
	log.Printf("💪 新しいワークアウトを作成中: %s", req.Name)

	// Go基礎技術1: append最適化 - 事前容量確保
	// 最大8個のオプションがあるので事前に容量を確保
	opts := make([]WorkoutOption, 0, 8)

	if req.Description != "" {
		opts = append(opts, WorkoutOption{Description: req.Description})
	}
	if req.Difficulty != proto.Difficulty_DIFFICULTY_UNSPECIFIED {
		opts = append(opts, WorkoutOption{Difficulty: convertProtoDifficulty(req.Difficulty)})
	}
	if req.MuscleGroup != proto.MuscleGroup_UNSPECIFIED {
		opts = append(opts, WorkoutOption{MuscleGroup: convertProtoMuscleGroup(req.MuscleGroup)})
	}
	if req.Sets > 0 {
		opts = append(opts, WorkoutOption{Sets: int(req.Sets)})
	}
	if req.Reps > 0 {
		opts = append(opts, WorkoutOption{Reps: int(req.Reps)})
	}
	if req.Weight > 0 {
		opts = append(opts, WorkoutOption{Weight: req.Weight})
	}
	if req.Notes != "" {
		opts = append(opts, WorkoutOption{Notes: req.Notes})
	}

	workout, err := s.workoutManager.CreateWorkout(req.Name, opts...)
	if err != nil {
		// Go基礎技術2: strings.Builder + 事前容量確保でエラーメッセージ構築
		return &proto.CreateWorkoutResponse{
			Workout: nil,
			Message: s.buildErrorMessage("ワークアウト作成", req.Name, err.Error()),
		}, nil
	}

	protoWorkout := convertToProtoWorkout(workout)
	return &proto.CreateWorkoutResponse{
		Workout: protoWorkout,
		Message: s.buildSuccessMessage("作成", req.Name),
	}, nil
}

// GetWorkout ワークアウトを取得
func (s *GRPCServer) GetWorkout(ctx context.Context, req *proto.GetWorkoutRequest) (*proto.GetWorkoutResponse, error) {
	log.Printf("🔍 ワークアウトを取得中: ID %d", req.Id)

	workout, err := s.workoutManager.GetWorkout(repository.WorkoutID(req.Id))
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %v", err)
	}

	protoWorkout := convertToProtoWorkout(workout)
	return &proto.GetWorkoutResponse{
		Workout: protoWorkout,
	}, nil
}

// UpdateWorkout ワークアウトを更新
func (s *GRPCServer) UpdateWorkout(ctx context.Context, req *proto.UpdateWorkoutRequest) (*proto.UpdateWorkoutResponse, error) {
	log.Printf("✏️ ワークアウトを更新中: ID %d", req.Id)

	err := s.workoutManager.UpdateWorkout(
		repository.WorkoutID(req.Id),
		req.Name,
		req.Description,
		convertProtoWorkoutStatus(req.Status),
		convertProtoDifficulty(req.Difficulty),
		convertProtoMuscleGroup(req.MuscleGroup),
		int(req.Sets),
		int(req.Reps),
		req.Weight,
		req.Notes,
	)
	if err != nil {
		return &proto.UpdateWorkoutResponse{
			Workout: nil,
			Message: fmt.Sprintf("❌ ワークアウト更新に失敗しました: %v", err),
		}, nil
	}

	// 更新されたワークアウトを取得
	workout, err := s.workoutManager.GetWorkout(repository.WorkoutID(req.Id))
	if err != nil {
		return &proto.UpdateWorkoutResponse{
			Workout: nil,
			Message: fmt.Sprintf("❌ 更新されたワークアウトの取得に失敗しました: %v", err),
		}, nil
	}

	protoWorkout := convertToProtoWorkout(workout)
	return &proto.UpdateWorkoutResponse{
		Workout: protoWorkout,
		Message: fmt.Sprintf("✅ ワークアウト「%s」が更新されました！", req.Name),
	}, nil
}

// DeleteWorkout ワークアウトを削除
func (s *GRPCServer) DeleteWorkout(ctx context.Context, req *proto.DeleteWorkoutRequest) (*proto.DeleteWorkoutResponse, error) {
	log.Printf("🗑️ ワークアウトを削除中: ID %d", req.Id)

	err := s.workoutManager.DeleteWorkout(repository.WorkoutID(req.Id))
	if err != nil {
		return &proto.DeleteWorkoutResponse{
			Message: fmt.Sprintf("❌ ワークアウト削除に失敗しました: %v", err),
		}, nil
	}

	return &proto.DeleteWorkoutResponse{
		Message: "✅ ワークアウトが削除されました！",
	}, nil
}

// ListWorkouts ワークアウト一覧を取得
func (s *GRPCServer) ListWorkouts(ctx context.Context, req *proto.ListWorkoutsRequest) (*proto.ListWorkoutsResponse, error) {
	log.Printf("📋 ワークアウト一覧を取得中...")

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

	// Go基礎技術: make関数でcapacity事前指定 + 効率的なフィルタリング
	// 全件数を基準に容量を確保（フィルタリング後は通常8-9割程度残る）
	convertedWorkouts := make([]*proto.Workout, 0, len(workouts))
	validCount := 0

	// 1回のループでフィルタリング + 変換を同時実行（効率化）
	for _, workout := range workouts {
		// Go基礎技術: 条件チェック（ジェネリクス関数を使わず直接処理で高速化）
		if workout.Name != "" && workout.ID > 0 {
			convertedWorkouts = append(convertedWorkouts, convertToProtoWorkout(workout))
			validCount++
		}
	}

	log.Printf("🔍 フィルタリング結果: 全%d件中、有効なワークアウト%d件を返します", len(workouts), validCount)

	count, err := s.workoutManager.GetWorkoutCount()
	if err != nil {
		count = len(workouts) // フォールバック
	}

	return &proto.ListWorkoutsResponse{
		Workouts:   convertedWorkouts,
		TotalCount: int32(count),
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
func (s *GRPCServer) buildWorkoutSummary(workouts []*repository.Workout) string {
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
		builder.WriteString(fmt.Sprintf("  %d. %s (%s)", i+1, workout.Name, workout.MuscleGroup))

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
func convertToProtoWorkout(workout *repository.Workout) *proto.Workout {
	protoWorkout := &proto.Workout{
		Id:          int32(workout.ID),
		Name:        workout.Name,
		Description: workout.Description,
		Status:      convertToProtoWorkoutStatus(workout.Status),
		Difficulty:  convertToProtoDifficulty(workout.Difficulty),
		MuscleGroup: convertToProtoMuscleGroup(workout.MuscleGroup),
		Sets:        int32(workout.Sets),
		Reps:        int32(workout.Reps),
		Weight:      workout.Weight,
		Notes:       workout.Notes,
		CreatedAt:   workout.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workout.UpdatedAt.Format(time.RFC3339),
	}

	if workout.CompletedAt != nil {
		protoWorkout.CompletedAt = workout.CompletedAt.Format(time.RFC3339)
	}

	return protoWorkout
}

func convertToProtoWorkoutStatus(status repository.WorkoutStatus) proto.WorkoutStatus {
	switch status {
	case repository.WorkoutStatusPlanned:
		return proto.WorkoutStatus_WORKOUT_STATUS_PLANNED
	case repository.WorkoutStatusInProgress:
		return proto.WorkoutStatus_WORKOUT_STATUS_IN_PROGRESS
	case repository.WorkoutStatusCompleted:
		return proto.WorkoutStatus_WORKOUT_STATUS_COMPLETED
	case repository.WorkoutStatusSkipped:
		return proto.WorkoutStatus_WORKOUT_STATUS_SKIPPED
	default:
		return proto.WorkoutStatus_WORKOUT_STATUS_UNSPECIFIED
	}
}

func convertProtoWorkoutStatus(status proto.WorkoutStatus) repository.WorkoutStatus {
	switch status {
	case proto.WorkoutStatus_WORKOUT_STATUS_PLANNED:
		return repository.WorkoutStatusPlanned
	case proto.WorkoutStatus_WORKOUT_STATUS_IN_PROGRESS:
		return repository.WorkoutStatusInProgress
	case proto.WorkoutStatus_WORKOUT_STATUS_COMPLETED:
		return repository.WorkoutStatusCompleted
	case proto.WorkoutStatus_WORKOUT_STATUS_SKIPPED:
		return repository.WorkoutStatusSkipped
	default:
		return repository.WorkoutStatusPlanned
	}
}

func convertToProtoDifficulty(difficulty repository.Difficulty) proto.Difficulty {
	switch difficulty {
	case repository.DifficultyBeginner:
		return proto.Difficulty_DIFFICULTY_BEGINNER
	case repository.DifficultyIntermediate:
		return proto.Difficulty_DIFFICULTY_INTERMEDIATE
	case repository.DifficultyAdvanced:
		return proto.Difficulty_DIFFICULTY_ADVANCED
	case repository.DifficultyBeast:
		return proto.Difficulty_DIFFICULTY_BEAST
	default:
		return proto.Difficulty_DIFFICULTY_UNSPECIFIED
	}
}

func convertToProtoMuscleGroup(muscleGroup repository.MuscleGroup) proto.MuscleGroup {
	switch muscleGroup {
	case repository.Chest:
		return proto.MuscleGroup_CHEST
	case repository.Back:
		return proto.MuscleGroup_BACK
	case repository.Legs:
		return proto.MuscleGroup_LEGS
	case repository.Shoulders:
		return proto.MuscleGroup_SHOULDERS
	case repository.Arms:
		return proto.MuscleGroup_ARMS
	case repository.Abs:
		return proto.MuscleGroup_ABS
	case repository.Core:
		return proto.MuscleGroup_CORE
	case repository.Glutes:
		return proto.MuscleGroup_GLUTES
	case repository.Cardio:
		return proto.MuscleGroup_CARDIO
	case repository.FullBody:
		return proto.MuscleGroup_FULL_BODY
	default:
		return proto.MuscleGroup_UNSPECIFIED
	}
}

func convertProtoDifficulty(difficulty proto.Difficulty) repository.Difficulty {
	switch difficulty {
	case proto.Difficulty_DIFFICULTY_BEGINNER:
		return repository.DifficultyBeginner
	case proto.Difficulty_DIFFICULTY_INTERMEDIATE:
		return repository.DifficultyIntermediate
	case proto.Difficulty_DIFFICULTY_ADVANCED:
		return repository.DifficultyAdvanced
	case proto.Difficulty_DIFFICULTY_BEAST:
		return repository.DifficultyBeast
	default:
		return repository.DifficultyBeginner
	}
}

func convertProtoMuscleGroup(muscleGroup proto.MuscleGroup) repository.MuscleGroup {
	switch muscleGroup {
	case proto.MuscleGroup_CHEST:
		return repository.Chest
	case proto.MuscleGroup_BACK:
		return repository.Back
	case proto.MuscleGroup_LEGS:
		return repository.Legs
	case proto.MuscleGroup_SHOULDERS:
		return repository.Shoulders
	case proto.MuscleGroup_ARMS:
		return repository.Arms
	case proto.MuscleGroup_ABS:
		return repository.Abs
	case proto.MuscleGroup_CORE:
		return repository.Core
	case proto.MuscleGroup_GLUTES:
		return repository.Glutes
	case proto.MuscleGroup_CARDIO:
		return repository.Cardio
	case proto.MuscleGroup_FULL_BODY:
		return repository.FullBody
	default:
		return repository.Unspecified
	}
}
