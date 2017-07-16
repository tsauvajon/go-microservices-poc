package task

const (
	// StatusNotStarted : this task isn't started yet
	StatusNotStarted = 0
	// StatusInProgress : this task is in progress
	StatusInProgress = 1
	// StatusFinished : this task is done
	StatusFinished = 2
)

/*
Task :
Consecutive IDs
state :
	0 – not started
	1 – in progress
	2 – finished
*/
type Task struct {
	ID    int `json:"id"`
	State int `json:"state"`
}
