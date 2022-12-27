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
	"eventloop/pkg/eventloop/internal"
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

func ctxWithValueAndTimeout(ctx context.Context, key any, val any, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithValue(ctx, key, val), timeout)
}

func TestToggleOn(t *testing.T) {
	const (
		WANT        = 2
		TRIGGERNAME = "TOGGLEON"
		TOGGLENAME  = REGISTER
	)

	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return fmt.Sprint(number)
		}
		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
	)

	var (
		ARGS                  = event.Args{Fun: numInc, TriggerName: TRIGGERNAME}
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
			return strconv.Itoa(number)
		}

		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
	)

	eventDefault, neErr := event.NewEvent(event.Args{Fun: numInc, TriggerName: EVENTNAME})
	if neErr != nil {
		t.Error(neErr)
	}

	defer cancel()

	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, eventDefault)
	})
	<-execCh

	evLoop.Toggle(TRIGGER)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME)
	})
	<-execCh

	evLoop.Toggle(TRIGGER)

	errG.Go(func() error {
		return evLoop.Trigger(ctx, EVENTNAME)
	})
	result, _ := strconv.Atoi(<-execCh)

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if result != WANT {
		t.Errorf("Number: %v; Want: %v", result, WANT)
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
	eventCh := make(chan bool)

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

	exitCh := isScheduledEventDone(ctx, eventCh, nil)
	for _, testValue := range tests {
		if testValue.init != nil {
			testValue.init()
		}
		tick := time.NewTicker(time.Millisecond * 10)
		select {
		case <-exitCh:
			if testValue.done != nil {
				testValue.done()
			}
		case <-tick.C:
			if testValue.errorFunc != nil {
				testValue.errorFunc()
			}
		}
	}
}

func TestStartScheduler(t *testing.T) {
	const (
		WANT       = 4
		INTERVAL   = time.Millisecond * 10
		EXECUTIONS = 4
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return strconv.Itoa(number)
		}
		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
	)

	defer cancel()

	errG.SetLimit(-1)

	evSched, neErr := event.NewEvent(event.Args{Fun: numInc, IntervalTime: INTERVAL})
	defer evLoop.RemoveEventByUUIDs(evSched.GetID())
	if neErr != nil {
		t.Error(neErr)
	}
	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, evSched)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})

	var result string
	for i := 0; i < EXECUTIONS; i++ {
		result = <-execCh
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if intRes, _ := strconv.Atoi(result); intRes != WANT {
		t.Errorf("Number = %d; WANT %d", intRes, WANT)
	}
}

func TestScheduleEventAfterStartAndStop(t *testing.T) {
	const (
		WANT       = 4
		INTERVAL   = time.Millisecond * 20
		EXECUTIONS = 4
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return strconv.Itoa(number)
		}
		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
	)
	defer cancel()

	evSched, neErr := event.NewEvent(event.Args{Fun: numInc, IntervalTime: INTERVAL})
	defer evLoop.RemoveEventByUUIDs(evSched.GetID())
	if neErr != nil {
		t.Error(neErr)
	}
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.RegisterEvent(ctx, evSched)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	var result string
	for i := 0; i < EXECUTIONS; i++ {
		result = <-execCh
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if intRes, _ := strconv.Atoi(result); intRes != WANT {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	} else {
		t.Log(number)
	}
}

func TestRemoveEvent(t *testing.T) {
	const (
		WANT          = 6
		TriggerName   = "RemoveEventRegularFirst"
		Interval      = time.Millisecond * 20
		IntervalExecs = 5
	)
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return strconv.Itoa(number)
		}
		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
	)
	defer cancel()
	evSched, neErr1 := event.NewEvent(event.Args{
		Fun:          numInc,
		IntervalTime: Interval})
	eventDefault, neErr2 := event.NewEvent(event.Args{Fun: numInc, TriggerName: TriggerName})
	eventDefault2, neErr3 := event.NewEvent(event.Args{Fun: numInc, TriggerName: TriggerName})
	eventDefault3, neErr4 := event.NewEvent(event.Args{Fun: numInc, TriggerName: TriggerName})

	if neErr1 != nil || neErr2 != nil || neErr3 != nil || neErr4 != nil {
		t.Error("error creating events: ", neErr1, neErr2, neErr3, neErr4)
	}

	registerErrGo(ctx, errG, evSched)
	registerErrGo(ctx, errG, eventDefault)
	registerErrGo(ctx, errG, eventDefault2)
	registerErrGo(ctx, errG, eventDefault3)

	for i := 0; i < 4; i++ {
		<-execCh
	}

	errG.Go(func() error {
		return evLoop.Trigger(ctx, string(INTERVALED))
	})
	for i := 0; i < IntervalExecs; i++ {
		<-execCh
	}

	t.Log(evLoop.RemoveEventByUUIDs(eventDefault3.GetID(), evSched.GetID(), eventDefault.GetID()))

	errG.Go(func() error {
		return evLoop.Trigger(ctx, TriggerName)
	})
	result, _ := strconv.Atoi(<-execCh)

	evLoop.RemoveEventByUUIDs(eventDefault2.GetID())

	errG.Go(func() error {
		return evLoop.Trigger(ctx, TriggerName)
	})
	<-execCh

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if result != WANT {
		t.Errorf("Number = %d; WANT %d", result, WANT)
	} else {
		t.Logf("Number = %d; WANT %d", result, WANT)
	}
}

func TestSubevent(t *testing.T) {
	const (
		WANT        = 10
		TRIGGERNAME = "SUBEVENTS_TEST"
	)
	var (
		mx          sync.Mutex
		number      int
		numIncMutex = func(ctx context.Context) string {
			mx.Lock()
			number++
			mx.Unlock()
			return strconv.Itoa(number)
		}
		execCh      = make(chan string)
		ctx, cancel = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, execCh, time.Second)
		errG        = new(errgroup.Group)
		eventArgs   = event.Args{Fun: numIncMutex, TriggerName: TRIGGERNAME}
	)

	defer cancel()

	var (
		evListener, neErr1    = event.NewEvent(eventArgs)
		evListener2, neErr2   = event.NewEvent(eventArgs)
		eventDefault, neErr3  = event.NewEvent(eventArgs)
		eventDefault2, neErr4 = event.NewEvent(eventArgs)
		eventDefault3, neErr5 = event.NewEvent(eventArgs)
	)

	if neErr1 != nil || neErr2 != nil || neErr3 != nil || neErr4 != nil || neErr5 != nil {
		t.Error(neErr1, neErr2, neErr3, neErr4, neErr5)
	}

	registerErrGo(ctx, errG, eventDefault)
	registerErrGo(ctx, errG, eventDefault2)
	registerErrGo(ctx, errG, eventDefault3)

	for i := 0; i < 3; i++ {
		<-execCh
	}
	errG.Go(func() error {
		return evLoop.Subscribe(ctx, []event.Interface{eventDefault, eventDefault2, eventDefault3},
			[]event.Interface{evListener, evListener2})
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(ctx, TRIGGERNAME)
	})
	errG.Go(func() error {
		return evLoop.Trigger(ctx, TRIGGERNAME)
	})
	var result string
	for i := 0; i < 10; i++ {
		result = <-execCh
	}
	if err := errG.Wait(); err != nil {
		t.Error(err)
	}

	if intRes, _ := strconv.Atoi(result); intRes != WANT {
		t.Errorf("Number = %v; WANT %d", result, WANT)
	}
}

func TestPrioritySync(t *testing.T) {
	const (
		WANT        = 4
		TRIGGERNAME = "TestPriorSync"
	)
	var (
		execs            int
		r                = 'A'
		defaultEventFunc = func(ctx context.Context) string {
			execs++
			r++
			return string(r)
		}
		priorityFunc = func(ctx context.Context) string {
			execs++
			return "DP"
		}
		highPriorityFunc = func(ctx context.Context) string {
			execs++
			return "HP"
		}
		resultCh         = make(chan string)
		ctx, cancel      = ctxWithValueAndTimeout(context.Background(), internal.EXEC_CH_CTX_KEY, resultCh, time.Second)
		errG             = new(errgroup.Group)
		defaultEventArgs = event.Args{Fun: defaultEventFunc, TriggerName: TRIGGERNAME}
	)

	defer cancel()

	var (
		evNormal1, neErr1   = event.NewEvent(defaultEventArgs)
		evNormal2, neErr2   = event.NewEvent(defaultEventArgs)
		evPrior, neErr3     = event.NewEvent(event.Args{Fun: priorityFunc, TriggerName: TRIGGERNAME, Priority: 1})
		evHighPrior, neErr4 = event.NewEvent(event.Args{Fun: highPriorityFunc, TriggerName: TRIGGERNAME, Priority: 2})
	)

	if neErr1 != nil || neErr2 != nil || neErr3 != nil || neErr4 != nil {
		t.Error(neErr1, neErr2, neErr3, neErr4)
	}

	registerErrGo(ctx, errG, evNormal1)
	registerErrGo(ctx, errG, evNormal2)
	registerErrGo(ctx, errG, evPrior)
	registerErrGo(ctx, errG, evHighPrior)

	for i := 0; i < 4; i++ {
		<-resultCh
	}

	errG.Go(func() error {
		return evLoop.Trigger(ctx, TRIGGERNAME)
	})

	for i := 0; i < 4; i++ {
		<-resultCh
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	if execs != WANT {
		t.Errorf("Number = %v; WANT %v", execs, WANT)
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
		globalBeforeEventArgs = event.Args{Fun: defaultEventFunc,
			TriggerName: string(BEFORE_TRIGGER),
			Priority:    BEFORE_PRIORITY}
		globalAfterEventArgs = event.Args{Fun: defaultEventFunc,
			TriggerName: string(AFTER_TRIGGER),
			Priority:    AFTER_PRIORITY}
		beforeEventArgs = event.Args{
			Fun:         defaultEventFunc,
			TriggerName: EVENTNAME,
			Priority:    BEFORE_PRIORITY,
		}
		afterEventArgs = event.Args{
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

	defer evLoop.RemoveEventByUUIDs(gaEvent.GetID(), gbEvent.GetID(), laEvent.GetID(), lbEvent.GetID())

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
	exitCode := m.Run()
	os.Exit(exitCode)
}
