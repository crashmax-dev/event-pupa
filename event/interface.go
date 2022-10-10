package event

import (
	"context"
	"eventloop/event/schedule"
	"eventloop/event/subscriber"
)

type Interface interface {
	GetId() int
	GetPriority() int
	SetPriority(prior int)
	RunFunction(ctx context.Context) string
	GetSubscriber() subscriber.Interface
	GetSchedule() schedule.Interface
	IsOnce() bool
}
