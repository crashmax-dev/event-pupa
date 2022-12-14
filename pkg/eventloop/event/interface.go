package event

import (
	"context"

	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
)

type Interface interface {
	GetID() uuid.UUID
	GetPriority() int
	SetPriority(prior int)
	RunFunction(ctx context.Context) string
	Subscriber() subscriber.Interface
	Interval() (interval.Interface, error)
	Once() (once.Interface, error)
}
