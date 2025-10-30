package domain

// WorkoutRepository ワークアウトリポジトリのインターフェース
// ドメイン層で定義することで依存性の逆転（Dependency Inversion）を実現
// repository層がこのインターフェースを実装する
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
