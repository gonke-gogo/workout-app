package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"golv2-learning-app/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// gRPCサーバーに接続
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("接続に失敗: %v", err)
	}
	defer conn.Close()

	client := proto.NewWorkoutServiceClient(conn)

	// ワークアウトデータ
	workouts := []struct {
		name        string
		description string
		notes       string
		muscleGroup proto.MuscleGroup
		difficulty  proto.Difficulty
		sets        int32
		reps        int32
		weight      float64
	}{
		{"ベンチプレス", "胸筋を鍛える王道種目", "今日は調子がいい！💪", proto.MuscleGroup_CHEST, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 10, 60.0},
		{"スクワット", "下半身を鍛える王道種目", "自重でもキツい...😅", proto.MuscleGroup_LEGS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 4, 15, 0.0},
		{"デッドリフト", "背中を鍛える最強種目", "野獣級の重量に挑戦！🔥", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 3, 8, 80.0},
		{"プランク", "体幹を鍛える基本種目", "1分間が永遠に感じる⏰", proto.MuscleGroup_CORE, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 3, 60, 0.0},
		{"懸垂", "背中と腕を同時に鍛える", "自重トレーニングの王様", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_BEAST, 3, 8, 0.0},
		{"ショルダープレス", "肩を大きくして逞しく見せる", "肩が丸くなってきた💪", proto.MuscleGroup_SHOULDERS, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 10, 30.0},
		{"カール", "腕を太くして逞しく", "腕が太くなってきた💪", proto.MuscleGroup_ARMS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 3, 15, 15.0},
		{"クランチ", "腹筋を割ってモテ男に", "腹筋が少し見えてきた😊", proto.MuscleGroup_ABS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 3, 20, 0.0},
		{"ランニング", "有酸素運動で脂肪燃焼", "雨でスキップ😢", proto.MuscleGroup_CARDIO, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 1, 30, 0.0},
		{"ヨガ", "柔軟性を高める", "時間がなくてスキップ😅", proto.MuscleGroup_CORE, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 1, 60, 0.0},
		{"インクラインベンチプレス", "上部胸筋を重点的に鍛える", "上部胸筋が効いてる！", proto.MuscleGroup_CHEST, proto.Difficulty_DIFFICULTY_BEAST, 4, 8, 45.0},
		{"ラットプルダウン", "広い背中で逆三角形を目指す", "背中が広がってきた！", proto.MuscleGroup_BACK, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 12, 45.0},
		{"サイドレイズ", "肩の外側を鍛える", "肩の外側が効いてる！", proto.MuscleGroup_SHOULDERS, proto.Difficulty_DIFFICULTY_INTERMEDIATE, 3, 15, 8.0},
		{"レッグプレス", "太ももを太くして逞しく", "脚が太くなってきた！", proto.MuscleGroup_LEGS, proto.Difficulty_DIFFICULTY_ADVANCED, 4, 12, 100.0},
		{"トライセップスプッシュダウン", "腕の後ろ側を鍛える", "腕の後ろ側が効いてる！", proto.MuscleGroup_ARMS, proto.Difficulty_DIFFICULTY_ADVANCED, 3, 12, 35.0},
	}

	fmt.Printf("🚀 %d個のワークアウトを作成開始！\n", len(workouts))

	successCount := 0
	for i, workout := range workouts {
		fmt.Printf("📝 [%d/%d] %s を作成中...\n", i+1, len(workouts), workout.name)

		req := &proto.CreateWorkoutRequest{
			Name:        workout.name,
			Description: workout.description,
			Notes:       workout.notes,
			MuscleGroup: workout.muscleGroup,
			Difficulty:  workout.difficulty,
			Sets:        workout.sets,
			Reps:        workout.reps,
			Weight:      workout.weight,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := client.CreateWorkout(ctx, req)
		cancel()

		if err != nil {
			fmt.Printf("❌ %s の作成に失敗: %v\n", workout.name, err)
			continue
		}

		successCount++
		fmt.Printf("✅ %s (ID: %d) を作成しました\n", resp.Workout.Name, resp.Workout.Id)

		// API負荷を避けるため少し待機
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\n🎉 完了！ %d/%d個のワークアウトを作成しました！\n", successCount, len(workouts))
	fmt.Println("📊 結果を確認するには: export LC_ALL=ja_JP.UTF-8 && evans -r repl -p 50051")
	fmt.Println("   そして call ListWorkouts を実行してください")
}
