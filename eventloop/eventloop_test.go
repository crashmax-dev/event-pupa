package eventloop

import (
	"EventManager/event"
	"EventManager/eventloop/types"
	"context"
	"sync"
	"testing"
	"time"
)

func newDefaultEventLoop() *eventLoop {
	return &eventLoop{
		//events: make([]event, 0),
		events:         make(map[string][]*types.event, 0),
		mx:             &sync.RWMutex{},
		disabled:       []EventFunc{},
		intervalEvents: make([]*event.eventSchedule, 0),
		stopScheduler:  make(chan bool),
	}
}

func TestOnce(t *testing.T) {
	var (
		number int
		evLoop EventLoop = NewEventLoop()
	)

	ctx, _ := context.WithCancel(context.Background())

	eventSingle := types.event{fun: func(ctx context.Context) {
		number++
	}, isOnce: true}
	evLoop.On(ctx, "test", &eventSingle, nil)
	evLoop.Trigger(ctx, "test")
	evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 200)
	if number != 1 {
		t.Errorf("Number = %d; want 1", number)
	}
}

func TestMultipleDefaultAndOnce(t *testing.T) {
	var (
		number int
		evLoop EventLoop = NewEventLoop()
	)

	numInc := func(ctx context.Context) {
		number++
	}

	ctx, _ := context.WithCancel(context.Background())

	eventFirst := types.event{fun: numInc}
	go evLoop.On(ctx, "test", &eventFirst, nil)
	eventSecond := types.event{fun: numInc}
	go evLoop.On(ctx, "test", &eventSecond, nil)
	eventOnce := types.event{fun: numInc, isOnce: true}
	go evLoop.On(ctx, "test", &eventOnce, nil)
	time.Sleep(time.Millisecond * 200)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 50)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 50)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 200)
	if number != 7 {
		t.Errorf("Number = %d; want 7", number)
	}
}

func TestStartScheduler(t *testing.T) {
	var (
		number int
		evLoop EventLoop = NewEventLoop()
		ctx, _           = context.WithCancel(context.Background())
		numInc           = func(ctx context.Context) {
			number++
		}
	)

	evSched := event.eventSchedule{
		base: types.event{
			fun: numInc,
		},
		interval: 500 * time.Millisecond,
		quit:     nil,
	}
	evLoop.ScheduleEvent(ctx, &evSched, nil)
	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 1900)
	if number != 3 {
		t.Errorf("Number = %d; want 3", number)
	}
}

func TestSubevent(t *testing.T) {
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

	evListener := event.eventListener{base: types.event{fun: numInc}}
	evListener2 := event.eventListener{base: types.event{fun: numInc}}
	var eventDefault = types.event{fun: numInc}
	var eventDefault2 = types.event{fun: numInc}
	var eventDefault3 = types.event{fun: numInc}
	go evLoop.On(ctx, "test", &eventDefault, nil)
	go evLoop.On(ctx, "test", &eventDefault2, nil)
	go evLoop.On(ctx, "test", &eventDefault3, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Subscribe(ctx, []*event.eventTrigger{&eventDefault.trigger, &eventDefault2.trigger, &eventDefault3.trigger}, []*event.eventListener{&evListener, &evListener2})
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)
	if number != 10 {
		t.Errorf("Number = %d; want 10", number)
	}
}
