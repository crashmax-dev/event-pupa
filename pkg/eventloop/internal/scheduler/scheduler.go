package scheduler

import (
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"sync"
)

type Scheduler struct {
	mx                 sync.Mutex
	isSchedulerRunning bool
	schedulerResults   map[string]string

	Stopper chan bool
}

func NewScheduler() Scheduler {
	return Scheduler{
		Stopper: make(chan bool)}
}

func (e *Scheduler) ResetResults() {
	e.schedulerResults = make(map[string]string)
}

func (e *Scheduler) IsSchedulerRunning() bool {
	return e.isSchedulerRunning
}

func (e *Scheduler) GetSchedulerResults() []string {
	e.mx.Lock()
	defer e.mx.Unlock()
	return maps.Values(e.schedulerResults)
}

func (e *Scheduler) RunSchedule(isRun bool) {
	e.isSchedulerRunning = isRun
}

func (e *Scheduler) SetSchedulerResults(key uuid.UUID, value string) {
	e.mx.Lock()
	defer e.mx.Unlock()
	e.schedulerResults[key.String()] = value
}
