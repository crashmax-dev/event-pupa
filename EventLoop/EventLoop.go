package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

type EventLoop interface {
	On(eventName string, eventFunc func())
	Trigger(eventName string)
}

type Event interface {
	GetName() string
}

type eventLoop struct {
	//events []event
	events map[string][]event
	mx     *sync.RWMutex
	test   int
}

type event struct {
	//name string
	priority int
	fun      func()
}

//func (e *event) GetName() string {
//	return e.name
//}

func (e *eventLoop) On(eventName string, eventFunc func(), priority int) {
	e.mx.Lock()
	defer e.mx.Unlock()
	//e.events = append(e.events, event{name: eventName, fun: eventFunc})
	newEvent := event{fun: eventFunc, priority: priority}
	e.events[eventName] = append(e.events[eventName], newEvent)
	if priority > 0 {
		sort.Slice(e.events[eventName], func(i, j int) bool {
			return e.events[eventName][i].priority > e.events[eventName][j].priority
		})
	}
}

func (e *eventLoop) Trigger(eventName string) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	// there is no particular need to sync event callbacks
	// and there is potential lock
	// just one of callable can freeze any actions with e,
	// cause of wg.Wait and сallable execution (long, in high loaded system even 1 second can be `forever`)
	for _, ev := range e.events[eventName] {
		go func(ev event) {
			ev.fun()
		}(ev)
	}
}

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal)

	evLoop := eventLoop{
		//events: make([]event, 0),
		events: make(map[string][]event, 0),
		mx:     &sync.RWMutex{},
	}

	// must be called async
	go evLoop.On("keke", func() {
		fmt.Printf("%s\n", "lol")
	}, 0)

	// must be called async
	go evLoop.On("keke", func() {
		fmt.Printf("%s\n", "lol2")
	}, 1)

	// must be called async
	go evLoop.Trigger("keke")

	// easiest way to loop main forever
	<-quit
}
