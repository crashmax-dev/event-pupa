package LearnProject

import (
	"context"
	"eventloop/event"
	"testing"
	"time"
)

func TestOnAndTrigger(t *testing.T) {
	const WANT = 1
	var (
		number      int
		evLoop      = NewEventLoop()
		ctx, cancel = context.WithCancel(context.Background())
		numInc      = func(ctx context.Context) {
			number++
		}
	)

	var (
		eventDefault = event.NewEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventDefault, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
	cancel()
}

func TestOnce(t *testing.T) {
	const WANT = 1
	var (
		number      int
		evLoop      = NewEventLoop()
		ctx, cancel = context.WithCancel(context.Background())
		numInc      = func(ctx context.Context) {
			number++
		}
	)

	eventSingle := event.NewOnceEvent(numInc)
	evLoop.On(ctx, "test", eventSingle, nil)
	evLoop.Trigger(ctx, "test")
	evLoop.Trigger(ctx, "test")
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}

	cancel()
}

func TestMultipleDefaultAndOnce(t *testing.T) {
	const WANT = 7
	var (
		number      int
		evLoop      = NewEventLoop()
		ctx, cancel = context.WithCancel(context.Background())
		numInc      = func(ctx context.Context) {
			number++
		}
	)

	var (
		eventFirst  = event.NewEvent(numInc)
		eventSecond = event.NewEvent(numInc)
		eventOnce   = event.NewOnceEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventFirst, nil)
	go evLoop.On(ctx, "test", eventSecond, nil)
	go evLoop.On(ctx, "test", eventOnce, nil)

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

	cancel()
}

func TestStartScheduler(t *testing.T) {
	const WANT = 3
	var (
		number      int
		evLoop      = NewEventLoop()
		ctx, cancel = context.WithCancel(context.Background())
		numInc      = func(ctx context.Context) {
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

	cancel()
}

func TestSubevent(t *testing.T) {
	const WANT = 10
	var (
		number      int
		evLoop      = NewEventLoop()
		ctx, cancel = context.WithCancel(context.Background())
		numInc      = func(ctx context.Context) {
			evLoop.LockMutex()
			number++
			evLoop.UnlockMutex()
		}
	)

	var (
		evListener    = event.NewEvent(numInc)
		evListener2   = event.NewEvent(numInc)
		eventDefault  = event.NewEvent(numInc)
		eventDefault2 = event.NewEvent(numInc)
		eventDefault3 = event.NewEvent(numInc)
	)

	go evLoop.On(ctx, "test", eventDefault, nil)
	go evLoop.On(ctx, "test", eventDefault2, nil)
	go evLoop.On(ctx, "test", eventDefault3, nil)
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

	cancel()
}
