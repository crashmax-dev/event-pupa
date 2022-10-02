package main

import (
	"Biba/helpers"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"sort"
	"sync"
	"time"
)

type EventFunc string

const (
	ON      EventFunc = "ON"
	TRIGGER EventFunc = "TRIGGER"
)

type EventLoop interface {
	On(eventName string, newEvent *event, out chan<- int)
	Trigger(eventName string)
	Toggle(eventFunc ...EventFunc)
	ScheduleEvent(event *eventSchedule, out chan<- int)
	StartScheduler()
	StopScheduler()
	RemoveEvent(id int) bool
	Subscribe(triggers []*eventTrigger, listeners []*eventListener)
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
	fun      func()
	isOnce   bool

	trigger eventTrigger
}

type eventSchedule struct {
	base     event
	interval int
	quit     chan bool
}

type eventTrigger struct {
	syncGroups []*sync.WaitGroup
}

type eventListener struct {
	base   event
	waiter *sync.WaitGroup
}

//func (e *event) GetName() string {
//	return e.name
//}

func (e *eventLoop) Subscribe(triggers []*eventTrigger, listeners []*eventListener) {
	for _, v := range listeners {

		sg := sync.WaitGroup{}
		waitCount := len(triggers)
		sg.Add(waitCount)
		v.waiter = &sg

		go func(waitCount int) {
			for {
				v.waiter.Wait()
				v.base.fun()
				v.waiter.Add(waitCount)
			}
		}(waitCount)

		for _, t := range triggers {
			if t.syncGroups == nil {
				t.syncGroups = []*sync.WaitGroup{}
			}
			t.syncGroups = append(t.syncGroups, &sg)
		}
	}
}

func (e *eventLoop) On(eventName string, newEvent *event, out chan<- int) {
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

func (e *eventLoop) Trigger(eventName string) {
	if slices.Contains(e.disabled, TRIGGER) {
		fmt.Println("Can't trigger, trigger disabled!")
		return
	}

	e.mx.RLock()
	defer e.mx.RUnlock()

	for i := len(e.events[eventName]) - 1; i >= 0; i-- {
		curEvent := e.events[eventName][i]
		go func(ev event) {
			ev.fun()
			for _, sg := range ev.trigger.syncGroups {
				sg.Done()
			}
		}(*curEvent)

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

func (e *eventLoop) runScheduledEvent(event *eventSchedule) {
	fmt.Println("Scheduled event started")
	ticker := time.NewTicker(time.Duration(event.interval) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			event.base.fun()

		//Я не понял как в селекте вызывать один и тот же код для разных каналов, пока так
		//TODO переделать в func Select(cases []SelectCase) (chosen int, recv Value, recvOK bool)
		case <-event.quit:
			ticker.Stop()
			fmt.Println("Scheduled event stopped")
			return

		case <-e.stopScheduler:
			ticker.Stop()
			fmt.Println("Scheduled event stopped")
			return
		}
	}
}

func (e *eventLoop) ScheduleEvent(newEvent *eventSchedule, out chan<- int) {
	e.curEventId++
	newEvent.base.id = e.curEventId
	e.intervalEvents = append(e.intervalEvents, newEvent)
	if e.isSchedulerRunning {
		e.runScheduledEvent(newEvent)
	}
	if out != nil {
		out <- e.curEventId
	}
}

func (e *eventLoop) StartScheduler() {
	for _, evts := range e.intervalEvents {
		go e.runScheduledEvent(evts)
	}
	e.isSchedulerRunning = true
	fmt.Println("Scheduler started")
}

func (e *eventLoop) StopScheduler() {
	fmt.Println("Scheduler stopping...")
	if len(e.intervalEvents) > 0 {
		e.stopScheduler <- true
	}
	e.isSchedulerRunning = false
	fmt.Println("Scheduler stopped.")
}

func (e *eventLoop) RemoveEvent(id int) bool {

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
	quit := make(chan os.Signal)

	var evLoop EventLoop = &eventLoop{
		//events: make([]event, 0),
		events:         make(map[string][]*event, 0),
		mx:             &sync.RWMutex{},
		disabled:       []EventFunc{},
		intervalEvents: make([]*eventSchedule, 0),
		stopScheduler:  make(chan bool),
	}

	var eventDefault = event{fun: func() {
		fmt.Printf("%s\n", "lol")
	}}
	go evLoop.On("keke", &eventDefault, nil)

	eventPriority := event{fun: func() {
		fmt.Printf("%s\n", "lol2")
	}, priority: 1}
	go evLoop.On("keke", &eventPriority, nil)

	eventSingle := event{fun: func() {
		fmt.Printf("%s\n", "Lol single")
	}, isOnce: true}
	go evLoop.On("keke", &eventSingle, nil)

	time.Sleep(50)
	go evLoop.Trigger("keke")

	time.Sleep(100)
	go evLoop.Trigger("keke")

	time.Sleep(150)
	go evLoop.Toggle(TRIGGER)
	go evLoop.Trigger("keke")
	time.Sleep(200)
	go evLoop.Trigger("keke")

	intervalEvent := eventSchedule{quit: make(chan bool), interval: 500, base: event{fun: func() {
		i := 1
		fmt.Printf("Hi, scheduler fire count: %v\t", i)
		i++
	}}}

	intervalEvtCh := make(chan int)
	go evLoop.ScheduleEvent(&intervalEvent, intervalEvtCh)
	evId := <-intervalEvtCh

	go evLoop.StartScheduler()
	time.Sleep(2 * time.Second)

	go evLoop.RemoveEvent(evId)

	time.Sleep(time.Second)
	//go evLoop.StopScheduler()

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Second)
	fmt.Println("Triggering sub event")

	evListener := eventListener{base: event{fun: func() {
		fmt.Println("Event in event srabotalo N O R M A L N O")
	}}}
	eventDefault.trigger = eventTrigger{}
	go evLoop.Subscribe([]*eventTrigger{&eventDefault.trigger}, []*eventListener{&evListener})
	time.Sleep(time.Second)
	go evLoop.Trigger("keke")
	time.Sleep(time.Second)
	go evLoop.Trigger("keke")
	// create goroutine which will emulate work
	// preventing deadlock
	go func() {
		for {
			time.Sleep(100)
		}
	}()

	// easiest way to loop main forever
	<-quit
}
