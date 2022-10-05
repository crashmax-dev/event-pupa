package main

import (
	"Biba/helpers"
	"context"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

type EventFunc string

const (
	ON      EventFunc = "ON"
	TRIGGER EventFunc = "TRIGGER"
)

type EventLoop interface {
	On(ctx context.Context, eventName string, newEvent *event, out chan<- int)
	Trigger(ctx context.Context, eventName string)
	Toggle(eventFunc ...EventFunc)
	ScheduleEvent(ctx context.Context, newEvent *eventSchedule, out chan<- int)
	StartScheduler(ctx context.Context)
	StopScheduler()
	RemoveEvent(id int) bool
	Subscribe(ctx context.Context, triggers []*eventTrigger, listeners []*eventListener)
}

//type Event interface {
//	GetName() string
//}

type eventLoop struct {
	//events []event
	events             map[string][]*event
	intervalEvents     []*eventSchedule
	mx                 *sync.RWMutex
	disabled           []EventFunc
	isSchedulerRunning bool
	curEventId         int
	stopScheduler      chan bool
}

type event struct {
	//name string
	id       int
	priority int
	fun      func(ctx context.Context)
	isOnce   bool

	trigger eventTrigger
}

type eventSchedule struct {
	base     event
	interval time.Duration
	quit     chan bool
}

type eventTrigger struct {
	channels []chan int
	mx       sync.Mutex
}

type eventListener struct {
	base event
	ch   []chan int
	mx   sync.Mutex
}

//func (e *event) GetName() string {
//	return e.name
//}

func (e *eventLoop) Subscribe(ctx context.Context, triggers []*eventTrigger, listeners []*eventListener) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}
	for _, v := range listeners {

		for _, t := range triggers {
			ch := make(chan int, 1)
			if v.ch == nil {
				v.ch = []chan int{}
			}
			v.ch = append(v.ch, ch)

			if t.channels == nil {
				t.channels = []chan int{}
			}
			t.channels = append(t.channels, ch)
		}

		go func(ctx context.Context, v *eventListener) {
			for {
				select {
				case <-ctx.Done():
					//TODO выводить в логи пердупреждение, что контекст закрыт
					return
				default:
					v.mx.Lock()
					fmt.Println("Waiting for triggers...")
					fmt.Printf("Reading channels: %v\n", v.ch)
					for _, ch := range v.ch {
						fmt.Printf("Reading from %v\n", ch)
						<-ch
					}
					fmt.Println("go viponlnyatsa")
					v.base.fun(ctx)
					v.mx.Unlock()
				}
			}
		}(ctx, v)
	}
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *eventLoop) On(ctx context.Context, eventName string, newEvent *event, out chan<- int) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}

	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- -1
		}
		fmt.Println("Can't attach listener, On disabled!")
		return
	}
	e.mx.Lock()
	defer e.mx.Unlock()

	e.curEventId++
	newEvent.id = e.curEventId
	e.events[eventName] = append(e.events[eventName], newEvent)
	if newEvent.priority > 0 {
		sort.Slice(e.events[eventName], func(i, j int) bool {
			return e.events[eventName][i].priority < e.events[eventName][j].priority
		})
	}

	if out != nil {
		out <- e.curEventId
	}

}

func (e *eventLoop) Trigger(ctx context.Context, eventName string) {

	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}

	if slices.Contains(e.disabled, TRIGGER) {
		fmt.Println("Can't trigger, trigger disabled!")
		return
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	for i := len(e.events[eventName]) - 1; i >= 0; i-- {
		curEvent := e.events[eventName][i]

		go func(ev *event) {
			ev.fun(ctx)
			ev.trigger.mx.Lock()
			fmt.Println("Sending messages...")
			fmt.Println(ev.trigger.channels)
			for _, ch := range ev.trigger.channels {
				fmt.Printf("Writing to %v\n", ch)
				ch <- 1
			}
			fmt.Println("All messages send")
			ev.trigger.mx.Unlock()
		}(curEvent)

		if curEvent.isOnce {
			//TODO пофиксить, иногда вылетает с ошибонькой
			e.events[eventName] = helpers.RemoveIndex(e.events[eventName], i)
		}
	}
}

func (e *eventLoop) Toggle(eventFuncs ...EventFunc) {
	for _, v := range eventFuncs {
		if x := slices.Index(e.disabled, v); x != -1 {
			e.disabled = helpers.RemoveIndex(e.disabled, x)
		} else {
			e.disabled = append(e.disabled, v)
		}
	}
}

// done нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - гг (канал самого ивента, канал ивентлупа и context.Done()
func done(eventCh, eventLoopCh <-chan bool, ctx context.Context) <-chan struct{} {
	result := make(chan struct{}, 1)
	result <- struct{}{}
	//fmt.Println("Bobs")
	select {
	case <-ctx.Done():
		fmt.Println("Scheduler stopped because of context")
		return result
	case <-eventCh:
		fmt.Println("Scheduler stopped because of event want to stop")
		return result
	case <-eventLoopCh:
		fmt.Println("Scheduler stopped because of event manager commands")
		return result
	default:
		return make(chan struct{})
	}
}

func (e *eventLoop) runScheduledEvent(ctx context.Context, event *eventSchedule) {
	fmt.Printf("Scheduled event started with interval %v\n", event.interval)
	ticker := time.NewTicker(event.interval)
	defer ticker.Stop()
	fmt.Println("Run infinite cycle")
	for {
		select {
		//case <-ticker.C:
		case <-ticker.C:
			event.base.fun(ctx)
		case <-done(event.quit, e.stopScheduler, ctx):
			fmt.Println("Scheduled event stopped")
			return
		case <-event.quit:
			fmt.Println("Event quited")
		case <-e.stopScheduler:
			fmt.Println("Scheduler stopped")
		case <-ctx.Done():
			fmt.Println("Context stopped")
		}
	}
	//fmt.Printf("Scheduled event finished")
}

// ScheduleEvent добавляет ивент в список ивентов-таймеров. Если шедулер запущен - запускает этот ивент.
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent *eventSchedule, out chan<- int) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}

	e.curEventId++
	newEvent.base.id = e.curEventId
	e.intervalEvents = append(e.intervalEvents, newEvent)

	if e.isSchedulerRunning {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- e.curEventId
	}

}

func (e *eventLoop) StartScheduler(ctx context.Context) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		fmt.Println("TEST TEST")
		return
	}

	for _, evts := range e.intervalEvents {
		go e.runScheduledEvent(ctx, evts)
	}
	e.isSchedulerRunning = true
	fmt.Println("Scheduler started")
}

func (e *eventLoop) StopScheduler() {
	fmt.Println("Scheduler stopping...")
	e.mx.Lock()
	defer e.mx.Unlock()
	if len(e.intervalEvents) > 0 && e.isSchedulerRunning {
		fmt.Println("Send signal to stop")
		e.stopScheduler <- true
	}
	e.isSchedulerRunning = false
	fmt.Println("Scheduler stopped.")
}

func (e *eventLoop) RemoveEvent(id int) bool {
	e.mx.Lock()
	defer e.mx.Unlock()

	for key, events := range e.events {
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].id == id {
				e.events[key] = helpers.RemoveIndex(e.events[key], i)
				return true
			}
		}
	}

	for i := len(e.intervalEvents) - 1; i >= 0; i-- {
		if e.intervalEvents[i].base.id == id {
			e.intervalEvents[i].quit <- true
			e.intervalEvents = helpers.RemoveIndex(e.intervalEvents, i)
			return true
		}
	}
	return false
}

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var evLoop EventLoop = &eventLoop{
		//events: make([]event, 0),
		events:         make(map[string][]*event, 0),
		mx:             &sync.RWMutex{},
		disabled:       []EventFunc{},
		intervalEvents: make([]*eventSchedule, 0),
		stopScheduler:  make(chan bool),
	}

	var eventDefault = event{fun: func(ctx context.Context) {
		fmt.Printf("%s\n", "lol")
	}}
	go evLoop.On(ctx, "keke", &eventDefault, nil)

	eventPriority := event{fun: func(ctx context.Context) {
		fmt.Printf("%s\n", "lol2")
	}, priority: 1}
	go evLoop.On(ctx, "keke", &eventPriority, nil)

	time.Sleep(50)
	go evLoop.Trigger(ctx, "keke")

	time.Sleep(100)
	go evLoop.Trigger(ctx, "keke")

	time.Sleep(150)
	go evLoop.Toggle(TRIGGER)
	go evLoop.Trigger(ctx, "keke")
	time.Sleep(200)
	go evLoop.Trigger(ctx, "keke")

	intervalEvent := eventSchedule{quit: make(chan bool), interval: 500 * time.Millisecond, base: event{fun: func(ctx context.Context) func(ctx context.Context) {
		i := 1
		return func(ctx context.Context) {
			fmt.Printf("Hi, scheduler fire count: %v\t", i)
			i++
		}
	}(ctx)}}

	intervalEvtCh := make(chan int)
	go evLoop.ScheduleEvent(ctx, &intervalEvent, intervalEvtCh)
	//evId := <-intervalEvtCh

	go evLoop.StartScheduler(ctx)
	time.Sleep(2 * time.Second)

	//go evLoop.RemoveEvent(evId)

	time.Sleep(time.Second)
	//go evLoop.StopScheduler()

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Second)
	fmt.Println("Triggering sub event")

	evListener := eventListener{base: event{fun: func(ctx context.Context) {
		fmt.Println("Event in event srabotalo N O R M A L N O")
	}}}
	eventDefault.trigger = eventTrigger{}
	go evLoop.Subscribe(ctx, []*eventTrigger{&eventDefault.trigger}, []*eventListener{&evListener})
	time.Sleep(time.Second)
	go evLoop.Trigger(ctx, "keke")
	//time.Sleep(time.Second)
	go evLoop.Trigger(ctx, "keke")
	// create goroutine which will emulate work
	// preventing deadlock
	//go func() {
	//	for {
	//		time.Sleep(100)
	//	}
	//}()

	// easiest way to loop main forever
	for {
		select {
		case sType := <-quit:
			signal.Stop(quit)
			fmt.Printf("Program finished: %v", sType)
			evLoop.StopScheduler()
			cancel()
			//evLoop.StopScheduler()
			os.Exit(0)
			return
		default:
			time.Sleep(time.Second)
		}
	}

	//go func() {
	//	signalType := <-quit
	//	signal.Stop(quit)
	//	log.Println("Exit command received. Exiting...")
	//
	//	// this is a good place to flush everything to disk
	//	// before terminating.
	//	log.Println("Signal type : ", signalType)
	//
	//	os.Exit(0)
	//
	//}()

}
