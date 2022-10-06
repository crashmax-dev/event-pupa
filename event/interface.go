package event

import (
	"EventManager/event/schedule"
	"EventManager/event/subscriber"
	"context"
)

type Interface interface {
	GetPriority() int
	SetPriority(prior int)
	RunFunction(ctx context.Context)
	GetSubscriber() subscriber.Interface
	GetSchedule() schedule.Interface
}
