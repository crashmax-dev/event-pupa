package eventloop

import (
	"EventManager/event"
	"context"
	"sync"
	"testing"
	"time"
)

func newDefaultEventLoop() *eventLoop {
	return &eventLoop{
		//events: make([]event, 0),
		events:         make(map[string][]event.Interface, 0),
		mx:             &sync.RWMutex{},
		disabled:       []EventFunction{},
		intervalEvents: make([]event.Interface, 0),
		stopScheduler:  make(chan bool),
	}
}

func TestOnAndTrigger(t *testing.T) {
	const WANT = 1
	var (
		number int
		evLoop = NewEventLoop()
		ctx, _ = context.WithCancel(context.Background())
		numInc = func(ctx context.Context) {
			number++
		}
	)

	var (
		eventDefault = event.NewEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventDefault)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}

func TestOnce(t *testing.T) {
	const WANT = 1
	var (
		number int
		evLoop = NewEventLoop()
		ctx, _ = context.WithCancel(context.Background())
		numInc = func(ctx context.Context) {
			number++
		}
	)

	eventSingle := event.NewOnceEvent(numInc)
	evLoop.On(ctx, "test", eventSingle)
	evLoop.Trigger(ctx, "test")
	evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}

func TestMultipleDefaultAndOnce(t *testing.T) {
	const WANT = 7
	var (
		number int
		evLoop = NewEventLoop()
		ctx, _ = context.WithCancel(context.Background())
		numInc = func(ctx context.Context) {
			number++
		}
	)

	var (
		eventFirst  = event.NewEvent(numInc)
		eventSecond = event.NewEvent(numInc)
		eventOnce   = event.NewOnceEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventFirst)
	go evLoop.On(ctx, "test", eventSecond)
	go evLoop.On(ctx, "test", eventOnce)

	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}

func TestStartScheduler(t *testing.T) {
	const WANT = 3
	var (
		number int
		evLoop = NewEventLoop()
		ctx, _ = context.WithCancel(context.Background())
		numInc = func(ctx context.Context) {
			number++
		}
	)

	evSched := event.NewIntervalEvent(numInc, time.Millisecond*500)
	evLoop.ScheduleEvent(ctx, evSched, nil)
	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 1900)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}

func TestSubevent(t *testing.T) {
	const WANT = 10
	var (
		number int
		evLoop = NewEventLoop()
		ctx, _ = context.WithCancel(context.Background())
		numInc = func(ctx context.Context) {
			evLoop.mx.Lock()
			number++
			evLoop.mx.Unlock()
		}
	)

	var (
		evListener    = event.NewEvent(numInc)
		evListener2   = event.NewEvent(numInc)
		eventDefault  = event.NewEvent(numInc)
		eventDefault2 = event.NewEvent(numInc)
		eventDefault3 = event.NewEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventDefault)
	go evLoop.On(ctx, "test", eventDefault2)
	go evLoop.On(ctx, "test", eventDefault3)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Subscribe(ctx, []event.Interface{eventDefault, eventDefault2, eventDefault3},
		[]event.Interface{evListener, evListener2})
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}
