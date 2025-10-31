package domain

type WorkoutRepository interface {
	CreateWorkout(workout *Workout) error

	GetWorkout(id WorkoutID) (*Workout, error)

	UpdateWorkout(workout *Workout) error

	DeleteWorkout(id WorkoutID) error

	ListWorkouts(statusFilter *int, difficultyFilter *int, muscleGroupFilter *int) ([]*Workout, error)

	GetWorkoutCount() (int, error)
}
