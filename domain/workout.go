package domain

import "time"

// WorkoutID ワークアウトIDの型定義
type WorkoutID int64

// Workout ワークアウトのドメインモデル（エンティティ）
type Workout struct {
	ID           WorkoutID     `json:"id" gorm:"primaryKey;autoIncrement"`
	ExerciseType ExerciseType  `json:"exercise_type" gorm:"not null;default:0"` // トレーニング種目（enum化）
	Description  string        `json:"description,omitempty" gorm:"type:text"`  // 空の場合はJSONから除外
	Status       WorkoutStatus `json:"status" gorm:"not null;default:0"`
	Difficulty   Difficulty    `json:"difficulty" gorm:"not null;default:1"`
	MuscleGroup  MuscleGroup   `json:"muscle_group" gorm:"not null;default:0"` // enum化
	Sets         int           `json:"sets" gorm:"default:3"`
	Reps         int           `json:"reps" gorm:"default:10"`
	Weight       float64       `json:"weight,omitempty" gorm:"default:0"` // 0の場合はJSONから除外
	Notes        string        `json:"notes,omitempty" gorm:"type:text"`  // 空の場合はJSONから除外
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"` // nilの場合はJSONから除外
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
