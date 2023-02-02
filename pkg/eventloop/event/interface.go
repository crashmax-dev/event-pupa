package event

import (
	"context"

	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
)

type Interface interface {
	GetUUID() string
	GetPriority() int
	GetPriorityString() string
	GetTriggerName() string
	RunFunction(ctx context.Context)
	After() (after.Interface, error)
	Subscriber() (subscriber.Interface, error)
	Interval() (interval.Interface, error)
	Once() (once.Interface, error)
	GetTypes() (out []Type)
}
