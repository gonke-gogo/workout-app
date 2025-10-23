package repository

import "time"

// WorkoutID ワークアウトIDの型定義
type WorkoutID int64

// WorkoutStatus ワークアウトステータスの型定義
type WorkoutStatus int

// Difficulty 難易度の型定義
type Difficulty int

// MuscleGroup 筋肉群の型定義
type MuscleGroup int

// Workout ワークアウト構造体
type Workout struct {
	ID          WorkoutID     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string        `json:"name" gorm:"size:255;not null"`
	Description string        `json:"description,omitempty" gorm:"type:text"` // 空の場合はJSONから除外
	Status      WorkoutStatus `json:"status" gorm:"not null;default:0"`
	Difficulty  Difficulty    `json:"difficulty" gorm:"not null;default:1"`
	MuscleGroup MuscleGroup   `json:"muscle_group" gorm:"not null;default:0"` // enum化
	Sets        int           `json:"sets" gorm:"default:3"`
	Reps        int           `json:"reps" gorm:"default:10"`
	Weight      float64       `json:"weight,omitempty" gorm:"default:0"` // 0の場合はJSONから除外
	Notes       string        `json:"notes,omitempty" gorm:"type:text"`  // 空の場合はJSONから除外
	CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"` // nilの場合はJSONから除外
}

// WorkoutSummary ワークアウト概要（omitemptyの活用例）
// APIレスポンスでオプショナルフィールドを適切に処理するための構造体
type WorkoutSummary struct {
	ID         WorkoutID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Difficulty string    `json:"difficulty"`
	// omitemptyを使用したオプショナルフィールド
	MuscleGroup *string    `json:"muscle_group,omitempty"` // nilの場合JSONから除外
	Weight      *float64   `json:"weight,omitempty"`       // nilの場合JSONから除外
	Notes       *string    `json:"notes,omitempty"`        // nilの場合JSONから除外
	CompletedAt *time.Time `json:"completed_at,omitempty"` // nilの場合JSONから除外
	// 統計情報（条件付きで含める）
	TotalVolume *float64 `json:"total_volume,omitempty"`     // Sets × Reps × Weight（計算値）
	Duration    *int     `json:"duration_minutes,omitempty"` // 実行時間（分）
}

// WorkoutFilter フィルタ条件（omitemptyの活用例）
// クエリパラメータで空の値は送信しない
type WorkoutFilter struct {
	Status      *WorkoutStatus `json:"status,omitempty"`       // nilの場合はフィルタしない
	Difficulty  *Difficulty    `json:"difficulty,omitempty"`   // nilの場合はフィルタしない
	MuscleGroup *MuscleGroup   `json:"muscle_group,omitempty"` // nilの場合はフィルタしない
	MinWeight   *float64       `json:"min_weight,omitempty"`   // nilの場合は重量でフィルタしない
	MaxWeight   *float64       `json:"max_weight,omitempty"`   // nilの場合は重量でフィルタしない
	DateFrom    *time.Time     `json:"date_from,omitempty"`    // nilの場合は日付でフィルタしない
	DateTo      *time.Time     `json:"date_to,omitempty"`      // nilの場合は日付でフィルタしない
}

// WorkoutRepository ワークアウトリポジトリのインターフェース
type WorkoutRepository interface {
	// CreateWorkout ワークアウトを作成
	CreateWorkout(workout *Workout) error

	// GetWorkoutByID ワークアウトをIDで取得
	GetWorkoutByID(id WorkoutID) (*Workout, error)

	// UpdateWorkout ワークアウトを更新
	UpdateWorkout(workout *Workout) error

	// DeleteWorkout ワークアウトを削除
	DeleteWorkout(id WorkoutID) error

	// ListWorkouts ワークアウト一覧を取得
	ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*Workout, error)

	// GetWorkoutCount ワークアウト数を取得
	GetWorkoutCount() (int, error)

	// GetWorkoutStats ワークアウト統計を取得
	GetWorkoutStats(period string) (map[string]interface{}, error)

	// Close リソースを解放
	Close() error
}

// 定数定義
const (
	WorkoutStatusPlanned    WorkoutStatus = iota // 予定
	WorkoutStatusInProgress                      // 実行中
	WorkoutStatusCompleted                       // 完了
	WorkoutStatusSkipped                         // スキップ（筋肉痛で...）
)

const (
	DifficultyBeginner     Difficulty = iota // 初心者
	DifficultyIntermediate                   // 中級者
	DifficultyAdvanced                       // 上級者
	DifficultyBeast                          // 野獣級
)

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

// String メソッド（日本語表示）
func (mg MuscleGroup) String() string {
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
