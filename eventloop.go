package eventloop

import (
	"context"
	"eventloop/event"
	"eventloop/helpers"
	"fmt"
	"golang.org/x/exp/slices"
	"sort"
	"sync"
	"time"
)

type eventLoop struct {
	//events []event
	events             map[string][]event.Interface
	intervalEvents     []event.Interface
	mx                 *sync.RWMutex
	disabled           []EventFunction
	isSchedulerRunning bool
	stopScheduler      chan bool
}

func NewEventLoop() Interface {
	//var evLoop eventloop = &eventLoop{
	//	events:         make(map[string][]*event, 0),
	//	mx:             &sync.RWMutex{},
	//	disabled:       []EventFunc{},
	//	intervalEvents: make([]*eventSchedule, 0),
	//	stopScheduler:  make(chan bool),
	//}
	return &eventLoop{mx: &sync.RWMutex{}, events: make(map[string][]event.Interface, 0)}
}

func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}
	for _, v := range listeners {
		trigger := v.GetSubscriber()
		for _, t := range triggers {
			ch := make(chan int, 1)
			trigger.AddChannel(ch)

			t.GetSubscriber().AddChannel(ch)
		}

		go func(ctx context.Context, v event.Interface) {
			for {
				select {
				case <-ctx.Done():
					//TODO выводить в логи пердупреждение, что контекст закрыт
					return
				default:

					trigger.LockMutex()
					//v.mx.Lock()
					fmt.Println("Waiting for triggers...")
					channels := trigger.GetChannels()
					fmt.Printf("Reading channels: %v\n", channels)
					for _, ch := range channels {
						fmt.Printf("Reading from %v\n", ch)
						<-ch
					}
					fmt.Println("go viponlnyatsa")
					v.RunFunction(ctx)
					trigger.UnlockMutex()
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

func (e *eventLoop) On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- int) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}

	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- newEvent.GetId()
		}
		fmt.Println("Can't attach listener, On disabled!")
		return
	}
	e.mx.Lock()
	defer e.mx.Unlock()

	e.events[eventName] = append(e.events[eventName], newEvent)
	if newEvent.GetPriority() > 0 {
		sort.Slice(e.events[eventName], func(i, j int) bool {
			return e.events[eventName][i].GetPriority() < e.events[eventName][j].GetPriority()
		})
	}
	fmt.Println(eventName, e.events[eventName])

	if out != nil {
		out <- newEvent.GetId()
	}

}

func (e *eventLoop) Trigger(ctx context.Context, eventName string, out chan<- string) {

	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		fmt.Println("Context canceled")
		return
	}
	if slices.Contains(e.disabled, TRIGGER) {
		fmt.Println("Can't subscriber, subscriber disabled!")
		return
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	for i := len(e.events[eventName]) - 1; i >= 0; i-- {
		curEvent := e.events[eventName][i]
		go func(ev event.Interface) {

			result := ev.RunFunction(ctx)
			if out != nil {
				out <- result
			}
			listener := ev.GetSubscriber()

			if listener == nil {
				return
			}
			if triggerChannels := listener.GetChannels(); len(triggerChannels) > 0 {
				evTrigger := ev.GetSubscriber()
				evTrigger.LockMutex()
				//evTrigger.mx.Lock()
				fmt.Println("Sending messages...")
				fmt.Println(triggerChannels)
				for _, ch := range triggerChannels {
					fmt.Printf("Writing to %v\n", ch)
					ch <- 1
				}
				fmt.Println("All messages send")
				evTrigger.UnlockMutex()
				//ev.subscriber.mx.Unlock()
			}
		}(curEvent)

		if curEvent.IsOnce() {
			//TODO пофиксить, иногда вылетает с ошибонькой
			e.events[eventName] = helpers.RemoveIndex(e.events[eventName], i)
		}
	}
}

func (e *eventLoop) Toggle(eventFuncs ...EventFunction) {
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

func (e *eventLoop) runScheduledEvent(ctx context.Context, event event.Interface) {
	evntInterval := event.GetSchedule().GetInterval()
	fmt.Printf("Scheduled event starting with interval %v\n", evntInterval)
	ticker := time.NewTicker(evntInterval)
	defer ticker.Stop()
	fmt.Println("Run infinite cycle")
	for {
		select {
		//case <-ticker.C:
		case <-ticker.C:
			//TODO подумоть, нужно ли запускать функцию интервального ивента как горутину
			go event.RunFunction(ctx)
		case <-done(event.GetSchedule().GetQuitChannel(), e.stopScheduler, ctx):
			fmt.Println("Scheduled event stopped")
			return
			//case <-event.quit:
			//	fmt.Println("Interface quited")
			//case <-e.stopScheduler:
			//	fmt.Println("Scheduler stopped")
			//case <-ctx.Done():
			//	fmt.Println("Context stopped")
		}
	}
	//fmt.Printf("Scheduled event finished")
}

// ScheduleEvent добавляет ивент в список ивентов-таймеров. Если шедулер запущен - запускает этот ивент.
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- int) {
	if isContextDone(ctx) {
		//TODO выводить в логи пердупреждение, что контекст закрыт
		return
	}

	e.intervalEvents = append(e.intervalEvents, newEvent)

	if e.isSchedulerRunning {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- newEvent.GetId()
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
			if events[i].GetId() == id {
				e.events[key] = helpers.RemoveIndex(e.events[key], i)
				return true
			}
		}
	}

	for i := len(e.intervalEvents) - 1; i >= 0; i-- {
		if e.intervalEvents[i].GetId() == id {
			e.intervalEvents[i].GetSchedule().GetQuitChannel() <- true
			e.intervalEvents = helpers.RemoveIndex(e.intervalEvents, i)
			return true
		}
	}
	return false
}

func (e *eventLoop) LockMutex() {
	e.mx.Lock()
}

func (e *eventLoop) UnlockMutex() {
	e.mx.Unlock()
}
