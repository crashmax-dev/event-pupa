package event

import (
	"context"

	"gitlab.com/YSX/eventloop/pkg/eventloop/event/after"
	"gitlab.com/YSX/eventloop/pkg/eventloop/event/interval"
	"gitlab.com/YSX/eventloop/pkg/eventloop/event/once"
	"gitlab.com/YSX/eventloop/pkg/eventloop/event/subscriber"
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
