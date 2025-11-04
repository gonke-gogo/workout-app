package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"golv2-learning-app/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// コマンドライン引数の定義
	var (
		addr    = flag.String("addr", "localhost:50051", "gRPCサーバーのアドレス (例: localhost:50051)")
		timeout = flag.Duration("timeout", 5*time.Second, "リクエストのタイムアウト時間")
	)
	flag.Parse()

	log.Printf("🔌 gRPCサーバーに接続: %s", *addr)
	log.Printf("⏱️  タイムアウト: %v", *timeout)

	// gRPCサーバーに接続
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("接続に失敗: %v", err)
	}
	defer conn.Close()

	client := proto.NewWorkoutServiceClient(conn)

	workouts := []struct {
		exerciseType proto.ExerciseType
		description  string
		notes        string
		muscleGroup  proto.MuscleGroup
		difficulty   proto.Difficulty
		sets         int32
		reps         int32
		weight       float64
	}{
		{proto.ExerciseType_EXERCISE_BENCH_PRESS, "胸筋を鍛える王道種目", "今日は調子がいい！💪", proto.MuscleGroup_CHEST, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 10, 60.0},
		{proto.ExerciseType_EXERCISE_SQUAT, "下半身を鍛える王道種目", "自重でもキツい...😅", proto.MuscleGroup_LEGS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 4, 15, 0.0},
		{proto.ExerciseType_EXERCISE_DEADLIFT, "背中を鍛える最強種目", "野獣級の重量に挑戦！🔥", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 3, 8, 80.0},
		{proto.ExerciseType_EXERCISE_PULL_UP, "背中と腕を同時に鍛える", "自重トレーニングの王様", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 3, 8, 0.0},
		{proto.ExerciseType_EXERCISE_DUMBBELL_SHOULDER, "肩を大きくして逞しく見せる", "肩が丸くなってきた💪", proto.MuscleGroup_SHOULDERS, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 10, 30.0},
		{proto.ExerciseType_EXERCISE_SIDE_RAISE, "肩の外側を鍛える", "肩の外側が効いてる！", proto.MuscleGroup_SHOULDERS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 3, 15, 8.0},
		{proto.ExerciseType_EXERCISE_ONE_HAND_ROW, "背中の厚みを作る", "背中が効いてる！", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 12, 20.0},
		{proto.ExerciseType_EXERCISE_HIGH_PULL, "肩と僧帽筋を鍛える", "肩が熱い！🔥", proto.MuscleGroup_SHOULDERS, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 10, 40.0},
		// 2回目のセット（バリエーション追加）
		{proto.ExerciseType_EXERCISE_BENCH_PRESS, "上部胸筋を重点的に鍛える", "上部胸筋が効いてる！", proto.MuscleGroup_CHEST, proto.Difficulty_DIFFICULTY_BEAST, 4, 8, 65.0},
		{proto.ExerciseType_EXERCISE_SQUAT, "深くしゃがんで効かせる", "脚がパンパン！💪", proto.MuscleGroup_LEGS, proto.Difficulty_DIFFICULTY_ADVANCED, 4, 12, 80.0},
		{proto.ExerciseType_EXERCISE_DEADLIFT, "高重量チャレンジ", "100kg目指す！", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 5, 5, 90.0},
		{proto.ExerciseType_EXERCISE_PULL_UP, "加重懸垂でパワーアップ", "重りつけて挑戦！", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 4, 6, 10.0},
	}

	fmt.Printf("🚀 %d個のワークアウトを作成開始！\n", len(workouts))

	successCount := 0
	for i, workout := range workouts {
		exerciseName := getExerciseTypeName(workout.exerciseType)
		fmt.Printf("📝 [%d/%d] %s を作成中...\n", i+1, len(workouts), exerciseName)

		req := &proto.CreateWorkoutRequest{
			ExerciseType: workout.exerciseType,
			Description:  workout.description,
			Notes:        workout.notes,
			MuscleGroup:  workout.muscleGroup,
			Difficulty:   workout.difficulty,
			Sets:         workout.sets,
			Reps:         workout.reps,
			Weight:       workout.weight,
		}

		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		resp, err := client.CreateWorkout(ctx, req)
		cancel()

		if err != nil {
			fmt.Printf("❌ %s の作成に失敗: %v\n", exerciseName, err)
			continue
		}

		successCount++
		fmt.Printf("✅ %s (ID: %d) を作成しました\n", getExerciseTypeName(resp.Workout.ExerciseType), resp.Workout.Id)

		// API負荷軽減の待機
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\n🎉 完了！ %d/%d個のワークアウトを作成しました！\n", successCount, len(workouts))
	fmt.Printf("📊 結果を確認するには: export LC_ALL=ja_JP.UTF-8 && evans -r repl -p %s\n", *addr)
	fmt.Println("   そして call ListWorkouts を実行してください")
}

// getExerciseTypeName ExerciseTypeの日本語名を取得
func getExerciseTypeName(exerciseType proto.ExerciseType) string {
	switch exerciseType {
	case proto.ExerciseType_EXERCISE_BENCH_PRESS:
		return "ベンチプレス"
	case proto.ExerciseType_EXERCISE_SQUAT:
		return "スクワット"
	case proto.ExerciseType_EXERCISE_DEADLIFT:
		return "デッドリフト"
	case proto.ExerciseType_EXERCISE_DUMBBELL_SHOULDER:
		return "ダンベルショルダープレス"
	case proto.ExerciseType_EXERCISE_PULL_UP:
		return "懸垂"
	case proto.ExerciseType_EXERCISE_SIDE_RAISE:
		return "サイドレイズ"
	case proto.ExerciseType_EXERCISE_ONE_HAND_ROW:
		return "ワンハンドロー"
	case proto.ExerciseType_EXERCISE_HIGH_PULL:
		return "ハイプル"
	default:
		return "未指定"
	}
}
