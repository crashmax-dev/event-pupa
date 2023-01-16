package event

import (
	"context"
	"errors"
	"sync"
	"time"

	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
	"eventloop/pkg/eventloop/internal"
	loggerEventLoop "eventloop/pkg/logger"
	"github.com/google/uuid"
)

// event - обычное событие, которое может иметь свойства других событий (одноразовых, интервальных, зависимых)

type Type string

const (
	TRIGGER  Type = "TRIGGER"
	ONCE     Type = "ONCE"
	INTERVAL Type = "INTERVAL"
	AFTER    Type = "AFTER"
)

type Args struct {
	TriggerName string
	Priority    int
	IsOnce      bool
	Fun         Func

	IntervalTime time.Duration
	after.DateAfter
	subscriber subscriber.Type
}

type event struct {
	id          uuid.UUID
	triggerName string
	priority    int
	fun         Func
	result      string

	mx sync.Mutex

	subscriber subscriber.Interface
	interval   interval.Interface
	once       once.Interface
	after      after.Interface
}

type Func func(ctx context.Context) string

func NewEvent(args Args) (Interface, error) {
	if args.Fun == nil {
		return nil, errors.New("no function, please add")
	}

	newEvent := &event{id: uuid.New(),
		fun:         args.Fun,
		triggerName: args.TriggerName,
		priority:    args.Priority}

	if args.IsOnce {
		newEvent.once = once.NewOnce()
	}
	if args.IntervalTime.String() != "0s" {
		newEvent.interval = interval.NewIntervalEvent(args.IntervalTime)
	}
	if args.DateAfter != (after.DateAfter{}) {
		newEvent.after = after.New(args.DateAfter)
	}

	switch args.subscriber {
	case subscriber.Listener:
		newEvent.subscriber = subscriber.NewSubscriberEvent()
	case subscriber.Trigger:
		newEvent.subscriber = subscriber.NewTriggerEvent()
	}

	return newEvent, nil
}

func (ev *event) GetID() uuid.UUID {
	return ev.id
}

func (ev *event) GetTypes() (out []Type) {
	if ev.triggerName != "" {
		out = append(out, TRIGGER)
	}
	if ev.once != nil {
		out = append(out, ONCE)
	}
	if ev.after != nil {
		out = append(out, AFTER)
	}
	if ev.interval != nil {
		out = append(out, INTERVAL)
	}
	if ev.subscriber != nil {
		out = append(out, SUBSCRIBER)
	}
	return
}

func (ev *event) GetTriggerName() string {
	return ev.triggerName
}

func (ev *event) GetPriority() int {
	return ev.priority
}

func (ev *event) RunFunction(ctx context.Context) {
	logger := ctx.Value(internal.LOGGER_CTX_KEY).(loggerEventLoop.Interface)

	logger.Debugw("Run event function", "eventId", ev.id)
	ev.result = ev.fun(ctx)
	defer internal.WriteToExecCh(ctx, ev.result)

	// Активация горутины этого триггера

	if subber, err := ev.Subscriber(); err == nil && subber.GetType() == subscriber.Trigger {
		logger.Debugw("Activating trigger goroutine", "eventId", ev.id)
		subber.ChanTrigger() <- struct{}{}
	}
}

var subError = errors.New("subscriber")

// Subscriber
func (ev *event) Subscriber() (subscriber.Interface, error) {
	return getSubInterface(ev.subscriber, subError)
}

var intError = errors.New("interval")

func (ev *event) Interval() (interval.Interface, error) {
	return getSubInterface(ev.interval, intError)
}

var onceError = errors.New("once")

func (ev *event) Once() (once.Interface, error) {
	return getSubInterface(ev.once, onceError)
}

var afterError = errors.New("after")

func (ev *event) After() (after.Interface, error) {
	return getSubInterface(ev.after, afterError)
}

func getSubInterface[T any](i T, err error) (T, error) {
	if reflect.ValueOf(i).IsValid() && !reflect.ValueOf(i).IsZero() {
		return i, nil
	}
	err = fmt.Errorf("it is not an event type: %w", err)
	return *new(T), err
}
