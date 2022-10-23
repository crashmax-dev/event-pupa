package eventloop

import (
	"context"
	"eventloop/event"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
	"time"
)

var (
	evLoop Interface
)

type Test struct {
	name string
	f    func(ctx context.Context, eventName string, farg func(ctx context.Context) string) string
	want int
}

func TestToggleOn(t *testing.T) {
	const (
		WANT      = 2
		EVENTNAME = "TOGGLEON"
	)

	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		eventDefault  = event.NewEvent(numInc)
		eventDefault2 = event.NewEvent(numInc)
		ctx, cancel   = context.WithTimeout(context.Background(), time.Second)
	)

	defer cancel()

	go evLoop.On(ctx, EVENTNAME, eventDefault, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 10)

	go evLoop.On(ctx, EVENTNAME, eventDefault2, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 10)

	go evLoop.On(ctx, EVENTNAME, eventDefault2, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.Trigger(ctx, EVENTNAME, nil)
	time.Sleep(time.Millisecond * 10)

	if number != WANT {
		t.Errorf("Number: %v; Want: %v", number, WANT)
	}
}

func TestToggleTrigger(t *testing.T) {
	const (
		WANT      = 1
		EVENTNAME = "toggletrigger"
	)

	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		eventDefault = event.NewEvent(numInc)
		ctx, cancel  = context.WithTimeout(context.Background(), time.Second)
	)

	defer cancel()

	go evLoop.On(ctx, EVENTNAME, eventDefault, nil)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Trigger(ctx, EVENTNAME, nil)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Trigger(ctx, EVENTNAME, nil)
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number: %v; Want: %v", number, WANT)
	}
}

func TestIsContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if isContextDone(ctx) {
		t.Errorf("Context isScheduledEventDone: true; Want: false")
	}
	cancel()
	if !isContextDone(ctx) {
		t.Errorf("Context isScheduledEventDone: false; Want: true")
	}
}

func TestIsScheduledEventDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eventCh, eventLoopCh := make(chan bool, 1), make(chan bool, 1)

	tests := []struct {
		init func()
		done func()
		dflt func()
	}{{
		done: func() {
			t.Errorf("isScheduledEventDone: true; Want: false")
		},
	},
		{
			init: func() {
				eventCh <- true
			},
			dflt: func() {
				t.Errorf("isScheduledEventDone by event channel: false; Want: true")
			},
		},
		{
			init: func() {
				eventLoopCh <- true
			},
			dflt: func() {
				t.Errorf("isScheduledEventDone by eventloop channel: false; Want: true")
			},
		},
		{
			init: func() {
				cancel()
			},
			dflt: func() {
				t.Errorf("isScheduledEventDone by context: false; Want: true")
			},
		}}

	for _, t := range tests {
		if t.init != nil {
			t.init()
		}
		select {
		case <-isScheduledEventDone(eventCh, eventLoopCh, ctx, nil):
			if t.done != nil {
				t.done()
			}
		default:
			if t.dflt != nil {
				t.dflt()
			}
		}
	}
}

func TestStartScheduler(t *testing.T) {
	const WANT = 4
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)

	evSched := event.NewIntervalEvent(numInc, time.Millisecond*20)
	go evLoop.ScheduleEvent(ctx, evSched, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 100)
	cancel()

	if number != WANT && number != WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestScheduleEventAfterStartAndStop(t *testing.T) {
	const WANT = 5
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)
	defer cancel()

	evSched := event.NewIntervalEvent(numInc, time.Millisecond*20)
	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 10)
	go evLoop.ScheduleEvent(ctx, evSched, nil)
	time.Sleep(time.Millisecond * 100)
	go evLoop.StopScheduler()
	time.Sleep(time.Millisecond * 500)

	if number != WANT && number != WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestRemoveEvent(t *testing.T) {
	const (
		WANT      = 7
		EVENTNAME = "RemoveEventRegularFirst"
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)
	defer cancel()
	var (
		evSched       = event.NewIntervalEvent(numInc, time.Millisecond*20)
		eventDefault  = event.NewEvent(numInc)
		eventDefault2 = event.NewEvent(numInc)
		eventDefault3 = event.NewEvent(numInc)
	)
	evLoop.On(ctx, EVENTNAME, eventDefault, nil)
	evLoop.On(ctx, EVENTNAME, eventDefault2, nil)
	evLoop.On(ctx, EVENTNAME, eventDefault3, nil)

	go evLoop.ScheduleEvent(ctx, evSched, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 100)

	go evLoop.RemoveEvent(eventDefault3.GetId())
	go evLoop.RemoveEvent(evSched.GetId())
	go evLoop.RemoveEvent(eventDefault.GetId())
	time.Sleep(time.Millisecond * 10)

	go evLoop.Trigger(ctx, EVENTNAME, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.RemoveEvent(eventDefault2.GetId())
	time.Sleep(time.Millisecond * 10)

	go evLoop.Trigger(ctx, EVENTNAME, nil)
	time.Sleep(time.Millisecond * 10)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	} else {
		t.Logf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestSubevent(t *testing.T) {
	const WANT = 10
	var (
		number      int
		numIncMutex = func(ctx context.Context) string {
			evLoop.LockMutex()
			number++
			evLoop.UnlockMutex()
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)

	var (
		evListener    = event.NewEvent(numIncMutex)
		evListener2   = event.NewEvent(numIncMutex)
		eventDefault  = event.NewEvent(numIncMutex)
		eventDefault2 = event.NewEvent(numIncMutex)
		eventDefault3 = event.NewEvent(numIncMutex)
	)

	go evLoop.On(ctx, "test", eventDefault, nil)
	go evLoop.On(ctx, "test", eventDefault2, nil)
	go evLoop.On(ctx, "test", eventDefault3, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Subscribe(ctx, []event.Interface{eventDefault, eventDefault2, eventDefault3},
		[]event.Interface{evListener, evListener2})
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test", nil)
	go evLoop.Trigger(ctx, "test", nil)
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}

	cancel()
}

func TestMain(m *testing.M) {
	evLoop = NewEventLoop(zapcore.DebugLevel)
	os.Exit(m.Run())
}
