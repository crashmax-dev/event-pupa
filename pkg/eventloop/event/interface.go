package event

import (
	"context"
	"eventloop/pkg/eventloop/event/schedule"
	"eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
	"sync"
)

type Interface interface {
	GetID() uuid.UUID
	GetPriority() int
	SetPriority(prior int)
	RunFunction(ctx context.Context) string
	GetSubscriber() subscriber.Interface
	GetSchedule() (schedule.Interface, error)
	IsOnce() bool
	GetOnce() *sync.Once
}
