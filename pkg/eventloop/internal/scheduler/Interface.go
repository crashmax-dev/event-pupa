package scheduler

type Interface interface {
	GetSchedulerResults() []string
	IsSchedulerRunning() bool
}
