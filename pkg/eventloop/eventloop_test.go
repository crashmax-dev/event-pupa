package eventloop

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

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
		WANT        = 2
		TRIGGERNAME = "TOGGLEON"
		TOGGLENAME  = REGISTER
	)

	var (
		number int
		execCh = make(chan string)
		numInc = func(ctx context.Context) string {
			number++
			return fmt.Sprint(number)
		}
		ctx, cancel = context.WithTimeout(context.WithValue(context.Background(), "execCh", execCh), time.Second)
		errG        = new(errgroup.Group)
	)

	var (
		ARGS                  = event.EventArgs{Fun: numInc, TriggerName: TRIGGERNAME}
		eventDefault, errNE1  = event.NewEvent(ARGS)
		eventDefault2, errNE2 = event.NewEvent(ARGS)
	)

	defer cancel()

	if errNE1 != nil || errNE2 != nil {
		t.Error("Error creating events: ", errNE1, errNE2)
	}

	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, eventDefault)
	})
	<-execCh

	evLoop.Toggle(TOGGLENAME)

	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, eventDefault2)
	})
	<-execCh

	evLoop.Toggle(TOGGLENAME)

	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, eventDefault2)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(ctx, TRIGGERNAME)
	})
	<-execCh
	result, _ := strconv.Atoi(<-execCh)

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}
	if number != WANT || result != WANT {
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
	eventCh := make(chan bool, 1)

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
		case <-isScheduledEventDone(ctx, eventCh, nil):
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
		INTERVALMS = time.Millisecond * 200
		EXECUTIONS = 4
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*200)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	errG.SetLimit(-1)

	evSched, neErr := event.NewEvent(event.EventArgs{Fun: numInc, IntervalTime: INTERVALMS})
	if neErr != nil {
		t.Error(neErr)
	}
	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, evSched)
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	time.Sleep(INTERVALMS * EXECUTIONS)
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})

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

	evSched, neErr := event.NewEvent(event.EventArgs{Fun: numInc, IntervalTime: INTERVALMS})
	if neErr != nil {
		t.Error(neErr)
	}
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, evSched)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	time.Sleep(EXECUTIONS * INTERVALMS)
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
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
		ctx, cancel           = context.WithTimeout(context.Background(), time.Second)
		errG                  = new(errgroup.Group)
		globalBeforeEventArgs = event.EventArgs{Fun: defaultEventFunc,
			TriggerName: string(BEFORE_TRIGGER),
			Priority:    BEFORE_PRIORITY}
		globalAfterEventArgs = event.EventArgs{Fun: defaultEventFunc,
			TriggerName: string(AFTER_TRIGGER),
			Priority:    AFTER_PRIORITY}
		beforeEventArgs = event.EventArgs{
			Fun:         defaultEventFunc,
			TriggerName: EVENTNAME,
			Priority:    BEFORE_PRIORITY,
		}
		afterEventArgs = event.EventArgs{
			Fun:         defaultEventFunc,
			TriggerName: EVENTNAME,
			Priority:    AFTER_PRIORITY,
		}
	)

	defer cancel()

	gaEvent, neErr1 := event.NewEvent(globalAfterEventArgs)
	gbEvent, neErr2 := event.NewEvent(globalBeforeEventArgs)
	laEvent, neErr3 := event.NewEvent(afterEventArgs)
	lbEvent, neErr4 := event.NewEvent(beforeEventArgs)

	if neErr1 != nil || neErr2 != nil || neErr3 != nil || neErr4 != nil {
		t.Error(neErr1, neErr2, neErr3, neErr4)
	}

	evLoop.RegisterEvent(ctx, gaEvent)
	evLoop.RegisterEvent(ctx, gbEvent)
	evLoop.RegisterEvent(ctx, laEvent)
	evLoop.RegisterEvent(ctx, lbEvent)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME)
	})

	errG.Go(func() error {
		return evLoop.Trigger(ctx, "Random Event")
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
