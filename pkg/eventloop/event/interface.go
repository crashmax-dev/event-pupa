package event

import (
	"context"

	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
)

type Interface interface {
	GetID() uuid.UUID
	GetPriority() int
	GetTriggerName() string
	RunFunction(ctx context.Context)
	After() (after.Interface, error)
	Subscriber() subscriber.Interface
	Interval() (interval.Interface, error)
	Once() (once.Interface, error)
}
