package domain

// WorkoutStatus ワークアウトステータスの型定義
type WorkoutStatus int

const (
	WorkoutStatusPlanned    WorkoutStatus = iota // 予定
	WorkoutStatusInProgress                      // 実行中
	WorkoutStatusCompleted                       // 完了
	WorkoutStatusSkipped                         // スキップ（筋肉痛で...）
)

// Difficulty 難易度の型定義
type Difficulty int

const (
	DifficultyBeginner     Difficulty = iota // 初心者
	DifficultyIntermediate                   // 中級者
	DifficultyAdvanced                       // 上級者
	DifficultyBeast                          // 化け物
)

// MuscleGroup 筋肉群の型定義
type MuscleGroup int

const (
	Unspecified MuscleGroup = iota // 未指定
	Chest                          // 胸
	Back                           // 背中
	Legs                           // 脚
	Shoulders                      // 肩
	Arms                           // 腕
	Abs                            // 腹筋
	Core                           // 体幹
	Glutes                         // 臀部
	Cardio                         // 有酸素
	FullBody                       // 全身
)

// ExerciseType トレーニング種目の型定義
type ExerciseType int

const (
	ExerciseUnspecified ExerciseType = iota // 未指定
	BenchPress                              // ベンチプレス
	Squat                                   // スクワット
	Deadlift                                // デッドリフト
	DumbbellShoulder                        // ダンベルショルダープレス
	PullUp                                  // 懸垂
	SideRaise                               // サイドレイズ
	OneHandRow                              // ワンハンドロー
	HighPull                                // ハイプル
)

// String メソッド（日本語表示）
func (mg MuscleGroup) Japanese() string {
	switch mg {
	case Chest:
		return "胸"
	case Back:
		return "背中"
	case Legs:
		return "脚"
	case Shoulders:
		return "肩"
	case Arms:
		return "腕"
	case Abs:
		return "腹筋"
	case Core:
		return "体幹"
	case Glutes:
		return "臀部"
	case Cardio:
		return "有酸素"
	case FullBody:
		return "全身"
	default:
		return "未指定"
	}
}

// English メソッド（英語表示）
func (mg MuscleGroup) English() string {
	switch mg {
	case Chest:
		return "Chest"
	case Back:
		return "Back"
	case Legs:
		return "Legs"
	case Shoulders:
		return "Shoulders"
	case Arms:
		return "Arms"
	case Abs:
		return "Abs"
	case Core:
		return "Core"
	case Glutes:
		return "Glutes"
	case Cardio:
		return "Cardio"
	case FullBody:
		return "Full Body"
	default:
		return "Unspecified"
	}
}

// Japanese メソッド（日本語表示）
func (et ExerciseType) Japanese() string {
	switch et {
	case BenchPress:
		return "ベンチプレス"
	case Squat:
		return "スクワット"
	case Deadlift:
		return "デッドリフト"
	case DumbbellShoulder:
		return "ダンベルショルダープレス"
	case PullUp:
		return "懸垂"
	case SideRaise:
		return "サイドレイズ"
	case OneHandRow:
		return "ワンハンドロー"
	case HighPull:
		return "ハイプル"
	default:
		return "未指定"
	}
}

// English メソッド（英語表示）
func (et ExerciseType) English() string {
	switch et {
	case BenchPress:
		return "Bench Press"
	case Squat:
		return "Squat"
	case Deadlift:
		return "Deadlift"
	case DumbbellShoulder:
		return "Dumbbell Shoulder Press"
	case PullUp:
		return "Pull Up"
	case SideRaise:
		return "Side Raise"
	case OneHandRow:
		return "One Hand Row"
	case HighPull:
		return "High Pull"
	default:
		return "Unspecified"
	}
}
