package event

import (
	"context"
	"eventloop/event/schedule"
	"eventloop/event/subscriber"
	"github.com/google/uuid"
)

type Interface interface {
	GetId() uuid.UUID
	GetPriority() int
	SetPriority(prior int)
	RunFunction(ctx context.Context) string
	GetSubscriber() subscriber.Interface
	GetSchedule() (schedule.Interface, error)
	IsOnce() bool
}
