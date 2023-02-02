package event

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
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

type Args struct {
	TriggerName string
	Priority    int
	IsOnce      bool
	Fun         Func

	IntervalTime time.Duration
	DateAfter    after.Args
	Subscriber   subscriber.Type
}

type event struct {
	uuid        string
	triggerName string
	priority    int
	fun         Func
	result      string

	disabled bool

	mx sync.Mutex

	subscriber subscriber.Interface
	interval   interval.Interface
	once       once.Interface
	after      after.Interface
}

type Func func(ctx context.Context) string

func NewEvent(args Args) (Interface, error) {
	if args.Fun == nil {
		return nil, errors.New("no run function")
	}

	// У ивента нет никаких условий для триггера
	if args.TriggerName == "" &&
		!args.IsOnce &&
		args.IntervalTime.String() == "0s" &&
		args.DateAfter == (after.Args{}) &&
		args.Subscriber == "" {
		return nil, errors.New("no event type, event will never trigger")
	}

	newEvent := &event{uuid: uuid.NewString(),
		fun:         args.Fun,
		triggerName: args.TriggerName,
		priority:    args.Priority,
	}

	if args.IsOnce {
		newEvent.once = once.NewOnce()
	}
	if args.IntervalTime.String() != "0s" {
		newEvent.interval = interval.NewIntervalEvent(args.IntervalTime)
	}
	if args.DateAfter != (after.Args{}) {
		newEvent.after = after.New(args.DateAfter)
	}

	switch args.Subscriber {
	case subscriber.Listener:
		newEvent.subscriber = subscriber.NewSubscriberEvent()
	case subscriber.Trigger:
		newEvent.subscriber = subscriber.NewTriggerEvent()
	}

	return newEvent, nil
}

func (ev *event) GetUUID() string {
	return ev.uuid
}

func (ev *event) GetTypes() (out []Type) {
	if ev.triggerName != "" {
		out = append(out, "TRIGGER")
	}
	if ev.once != nil {
		out = append(out, "ONCE")
	}
	if ev.after != nil {
		out = append(out, "AFTER")
	}
	if ev.interval != nil {
		out = append(out, "INTERVAL")
	}
	if ev.subscriber != nil {
		out = append(out, "SUBSCRIBER")
	}
	return
}

func (ev *event) GetTriggerName() string {
	return ev.triggerName
}

func (ev *event) GetPriority() int {
	return ev.priority
}

func (ev *event) GetPriorityString() string {
	return strconv.Itoa(ev.priority)
}

func (ev *event) RunFunction(ctx context.Context) {
	logger := loggerEventLoop.FromContext(ctx)

	logger.Debugw("Run event function", "eventId", ev.uuid)
	ev.result = ev.fun(ctx)
	defer internal.WriteToExecCh(ctx, ev.result)

	// Активация горутины этого триггера
	if subber, err := ev.Subscriber(); err == nil && subber.GetType() == subscriber.Trigger {
		logger.Debugw("Activating trigger goroutine", "eventId", ev.uuid)
		subber.ChanTrigger() <- struct{}{}
	}
}

var eventErrors = struct {
	subscriber error
	interval   error
	once       error
	after      error
}{
	subscriber: errors.New("subscriber"),
	interval:   errors.New("interval"),
	once:       errors.New("once"),
	after:      errors.New("after"),
}

// Subscriber
func (ev *event) Subscriber() (subscriber.Interface, error) {
	return getSubInterface(ev.subscriber, eventErrors.subscriber)
}

func (ev *event) Interval() (interval.Interface, error) {
	return getSubInterface(ev.interval, eventErrors.interval)
}

func (ev *event) Once() (once.Interface, error) {
	return getSubInterface(ev.once, eventErrors.once)
}

func (ev *event) After() (after.Interface, error) {
	return getSubInterface(ev.after, eventErrors.after)
}

func AsType(s string) (Type, error) {
	s = strings.ToUpper(s)
	switch s {
	case "TRIGGER":
		return "TRIGGER", nil
	case "ONCE":
		return "ONCE", nil
	case "INTERVAL":
		return "INTERVAL", nil
	case "AFTER":
		return "AFTER", nil
	case "SUBSCRIBER":
		return "SUBSCRIBER", nil
	default:
		return "", fmt.Errorf("no such type: %v", s)
	}
}

func getSubInterface[T any](i T, err error) (T, error) {
	if reflect.ValueOf(i).IsValid() && !reflect.ValueOf(i).IsZero() {
		return i, nil
	}
	err = fmt.Errorf("it is not an event type: %w", err)
	return *new(T), err
}
