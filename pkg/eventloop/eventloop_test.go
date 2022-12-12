package eventloop

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"eventloop/pkg/channelEx"
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

var (
	evLoop Interface
)

type test struct {
	name string
	f    func(ctx context.Context, t *testing.T, eventName string, farg func(ctx context.Context) string) string
	want int
}

func handleError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
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
		errG          = new(errgroup.Group)
	)

	defer cancel()

	errG.Go(func() error {
		return evLoop.On(ctx, EVENTNAME, eventDefault, nil)
	})
	time.Sleep(time.Millisecond * 10)

	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.On(ctx, EVENTNAME, eventDefault2, nil)
	})
	time.Sleep(time.Millisecond * 10)

	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.On(ctx, EVENTNAME, eventDefault2, nil)
	})
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

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
		errG         = new(errgroup.Group)
	)

	defer cancel()

	errG.Go(func() error {
		return evLoop.On(ctx, EVENTNAME, eventDefault, nil)
	})
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

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

// Проверяет функцию isScheduledEventDone
func TestIsScheduledEventDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eventCh, eventLoopCh := make(chan bool, 1), make(chan bool, 1)

	tests := []struct {
		init      func()
		done      func()
		errorFunc func()
	}{{ // Незаканчивающееся событие
		done: func() {
			t.Errorf("isScheduledEventDone: true; Want: false")
		},
	},
		{ // ====
			init: func() {
				eventCh <- true
			},
			errorFunc: func() {
				t.Errorf("isScheduledEventDone by event channel: false; Want: true")
			},
		},
		{ // ====
			init: func() {
				eventLoopCh <- true
			},
			errorFunc: func() {
				t.Errorf("isScheduledEventDone by eventloop channel: false; Want: true")
			},
		},
		{ // =====
			init: func() {
				cancel()
			},
			errorFunc: func() {
				t.Errorf("isScheduledEventDone by context: false; Want: true")
			},
		}}

	for _, testValue := range tests {
		if testValue.init != nil {
			testValue.init()
		}
		select {
		case <-isScheduledEventDone(ctx, eventCh, eventLoopCh, nil):
			if testValue.done != nil {
				testValue.done()
			}
		default:
			if testValue.errorFunc != nil {
				testValue.errorFunc()
			}
		}
	}
}

func TestStartScheduler(t *testing.T) {
	const (
		WANT       = 3
		INTERVALMS = time.Millisecond * 20
		EXECUTIONS = 4
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	evSched := event.NewIntervalEvent(numInc, INTERVALMS)
	errG.Go(func() error {
		return evLoop.ScheduleEvent(ctx, evSched, nil)
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.StartScheduler(ctx)
	})
	time.Sleep(INTERVALMS * EXECUTIONS)
	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if number != WANT && number != WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestScheduleEventAfterStartAndStop(t *testing.T) {
	const (
		WANT       = 4
		INTERVALMS = time.Millisecond * 20
		EXECUTIONS = 4
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)
	defer cancel()

	evSched := event.NewIntervalEvent(numInc, INTERVALMS)
	errG.Go(func() error {
		return evLoop.StartScheduler(ctx)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.ScheduleEvent(ctx, evSched, nil)
	})
	time.Sleep(EXECUTIONS * INTERVALMS)
	errG.Go(func() error {
		evLoop.StopScheduler()
		return nil
	})

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if number != WANT && number != WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	} else {
		t.Log(number)
	}
}

func TestRemoveEvent(t *testing.T) {
	const (
		WANT           = 7
		EVENTNAME      = "RemoveEventRegularFirst"
		INTERVAL_MS    = time.Millisecond * 20
		INTERVAL_EXECS = 5
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)
	defer cancel()
	var (
		evSched       = event.NewIntervalEvent(numInc, INTERVAL_MS)
		eventDefault  = event.NewEvent(numInc)
		eventDefault2 = event.NewEvent(numInc)
		eventDefault3 = event.NewEvent(numInc)
	)
	handleError(t, evLoop.On(ctx, EVENTNAME, eventDefault, nil))
	handleError(t, evLoop.On(ctx, EVENTNAME, eventDefault2, nil))
	handleError(t, evLoop.On(ctx, EVENTNAME, eventDefault3, nil))

	errG.Go(func() error {
		return evLoop.ScheduleEvent(ctx, evSched, nil)
	})
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.StartScheduler(ctx)
	})
	time.Sleep(INTERVAL_MS * INTERVAL_EXECS)

	go evLoop.RemoveEventByUUIDs([]uuid.UUID{eventDefault3.GetID(), evSched.GetID(), eventDefault.GetID()})
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})
	time.Sleep(time.Millisecond * 10)

	go evLoop.RemoveEventByUUIDs([]uuid.UUID{eventDefault2.GetID()})
	time.Sleep(time.Millisecond * 10)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if number < WANT || number > WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	} else {
		t.Logf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestSubevent(t *testing.T) {
	const WANT = 10
	var (
		mx          sync.Mutex
		number      int
		numIncMutex = func(ctx context.Context) string {
			mx.Lock()
			number++
			mx.Unlock()
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	var (
		evListener    = event.NewEvent(numIncMutex)
		evListener2   = event.NewEvent(numIncMutex)
		eventDefault  = event.NewEvent(numIncMutex)
		eventDefault2 = event.NewEvent(numIncMutex)
		eventDefault3 = event.NewEvent(numIncMutex)
	)

	errG.Go(func() error {
		return evLoop.On(ctx, "test", eventDefault, nil)
	})
	errG.Go(func() error {
		return evLoop.On(ctx, "test", eventDefault2, nil)
	})
	errG.Go(func() error {
		return evLoop.On(ctx, "test", eventDefault3, nil)
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Subscribe(ctx, []event.Interface{eventDefault, eventDefault2, eventDefault3},
			[]event.Interface{evListener, evListener2})
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Trigger(ctx, "test", nil)
	})
	errG.Go(func() error {
		return evLoop.Trigger(ctx, "test", nil)
	})

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	// Нужна задержка, т.к. мы дожидаемся выполнения ивентов-тригеров, но не дожидаемся ивентов-слушателей
	time.Sleep(time.Millisecond * 10)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}
}

func TestPrioritySync(t *testing.T) {
	const WANT = 4
	var (
		r                = 'A'
		defaultEventFunc = func(ctx context.Context) string {
			r++
			return string(r)
		}
		priorityFunc = func(ctx context.Context) string {
			return "DP"
		}
		highPriorityFunc = func(ctx context.Context) string {
			return "HP"
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	var (
		evNormal1   = event.NewEvent(defaultEventFunc)
		evNormal2   = event.NewEvent(defaultEventFunc)
		evPrior     = event.NewPriorityEvent(priorityFunc, 1)
		evHighPrior = event.NewPriorityEvent(highPriorityFunc, 2)
	)

	handleError(t, evLoop.On(ctx, "TestPriorSync", evNormal1, nil))
	handleError(t, evLoop.On(ctx, "TestPriorSync", evNormal2, nil))
	handleError(t, evLoop.On(ctx, "TestPriorSync", evHighPrior, nil))
	handleError(t, evLoop.On(ctx, "TestPriorSync", evPrior, nil))

	ch := channelEx.NewChannel(0)
	var numExecs int
	errG.Go(func() error {
		return evLoop.Trigger(ctx, "TestPriorSync", ch)
	})
	for data := range ch.Channel() {
		fmt.Println(data)
		numExecs++
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if numExecs != WANT {
		t.Errorf("Number = %d; WANT %d", numExecs, WANT)
	}
}

// Before-After
func TestBeforeAfter(t *testing.T) {
	const (
		WANT      = 6
		EVENTNAME = "BEFORE_AFTER_EVENT"
	)
	var (
		result           = 0
		defaultEventFunc = func(ctx context.Context) string {
			result++
			return fmt.Sprintf("%v", result)
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	evLoop.OnAfter(ctx, EVENTNAME, defaultEventFunc)
	evLoop.OnAfter(ctx, "", defaultEventFunc)
	evLoop.OnBefore(ctx, EVENTNAME, defaultEventFunc)
	evLoop.OnBefore(ctx, "", defaultEventFunc)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME, nil)
	})

	errG.Go(func() error {
		return evLoop.Trigger(ctx, "Random Event", nil)
	})

	if err := errG.Wait(); err != nil {
		t.Errorf(err.Error())
	}

	if result != WANT {
		t.Errorf("Number = %d; WANT %d", result, WANT)
	}
}

func TestMain(m *testing.M) {
	evLoop = NewEventLoop(zapcore.DebugLevel.String())
	os.Exit(m.Run())
}
