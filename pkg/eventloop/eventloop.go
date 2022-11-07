package eventloop

import (
	"context"
	"errors"
	"eventloop/internal/logger"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
	"sync"
	"time"
)

const INTERVALED = "@INTERVALED"

// 1. String - EventName, 2. Int - приоритет, 3. String - Event Id
type eventsList map[string]map[int]map[string]event.Interface

type Channel[T any] struct {
	Ch       chan T
	isClosed bool
}

type eventLoop struct {
	events eventsList
	mx     *sync.RWMutex

	disabled           []EventFunction
	isSchedulerRunning bool
	stopScheduler      chan bool

	logger *zap.SugaredLogger
}

func NewEventLoop(level zapcore.Level) Interface {
	elLogger, _ := logger.Initialize(level, "logs", "")

	return &eventLoop{
		mx:            &sync.RWMutex{},
		events:        make(eventsList),
		stopScheduler: make(chan bool),
		logger:        elLogger,
	}
}

func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) {
	if isContextDone(ctx) {
		e.logger.Warnw("Can't subscribe, context is done",
			"triggers", triggers,
			"listeners", listeners)
		return
	}
	for _, v := range listeners {
		trigger := v.GetSubscriber()
		for _, t := range triggers {
			ch := make(chan int, 1)
			trigger.AddChannel(ch)
			e.logger.Infow("Event subscribed", "trigger", t.GetId(), "listener", v.GetId())
			t.GetSubscriber().AddChannel(ch)
		}

		//Запскаем ждуна, когда триггеры сработают, и срабатываем сами
		go func(ctx context.Context, v event.Interface) {
			for {
				select {
				case <-ctx.Done():
					e.logger.Infow("Stop listening because of context", "eventId", v.GetId())
					return
				default:
					trigger.LockMutex()
					e.logger.Debugw("Waiting for triggers...", "event", v.GetId())
					channels := trigger.GetChannels()
					e.logger.Debugw("Reading channels", "channels", channels, "event", v.GetId())
					for _, ch := range channels {
						e.logger.Debugw("Waiting for channel", "Ch", ch, "event", v.GetId())
						<-ch
					}
					e.logger.Infow("Subscriber event fired", "event", v.GetId())
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

func (e *eventLoop) addEvent(eventName string, newEvent event.Interface) {
	if e.events[eventName] == nil {
		e.events[eventName] = make(map[int]map[string]event.Interface)
	}

	if e.events[eventName][newEvent.GetPriority()] == nil {
		e.events[eventName][newEvent.GetPriority()] = make(map[string]event.Interface)
	}

	e.events[eventName][newEvent.GetPriority()][newEvent.GetId().String()] = newEvent
}

func (e *eventLoop) On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- uuid.UUID) {
	if isContextDone(ctx) {
		e.logger.Warnw("Can't add listener to event, context is done",
			"event", newEvent.GetId(),
			"eventname", eventName)
		return
	}

	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- newEvent.GetId()
		}
		e.logger.Warnw("Can't attach listener, On disabled",
			"event", newEvent.GetId(),
			"eventname", eventName)
		return
	}

	if eventName == INTERVALED {
		e.logger.Warnf("Event name %v is reserved", INTERVALED)
		return
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.addEvent(eventName, newEvent)

	e.logger.Debugw("Event added", "eventname", eventName, "eventlist", e.events[eventName])

	if out != nil {
		out <- newEvent.GetId()
	}
}

// Trigger must be executed as goroutine, or it will be blocked!!!
func (e *eventLoop) Trigger(ctx context.Context, eventName string, ch *Channel[string]) {

	defer closechannel(ch)

	if isContextDone(ctx) {
		e.logger.Warnw("Can't trigger event, context is done",
			"eventname", eventName)
		return
	}
	if slices.Contains(e.disabled, TRIGGER) {
		e.logger.Warnw("Can't trigger event, trigger is disabled",
			"eventname", eventName)
		return
	}

	if eventName == INTERVALED {
		e.logger.Warnf("Event name %v is reserved", INTERVALED)
		return
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("Events triggered", "eventname", eventName, "eventscount", len(e.events[eventName]))

	var (
		wg sync.WaitGroup
	)

	for priorIndex := len(e.events[eventName]) - 1; priorIndex >= 0; priorIndex-- {
		for _, loopevent := range e.events[eventName][priorIndex] {
			wg.Add(1)

			go func(ev event.Interface) {
				defer wg.Done()

				result := ev.RunFunction(ctx)
				if ch != nil {
					ch.Ch <- result
				}

				listener := ev.GetSubscriber()
				if listener == nil {
					return
				}
				if listenerChannels := listener.GetChannels(); len(listenerChannels) > 0 {
					evTrigger := ev.GetSubscriber()
					evTrigger.LockMutex()
					e.logger.Debugw("Sending messages...",
						"listener", listenerChannels,
						"trigger", ev.GetId())
					for _, chnl := range listenerChannels {
						e.logger.Debugw("Writing to channel", "channel", ch, "trigger", ev.GetId())
						chnl <- 1
					}
					e.logger.Infow("All messages send", "trigger", ev.GetId())
					evTrigger.UnlockMutex()
				}
			}(loopevent)

			if loopevent.IsOnce() {
				loopevent.GetOnce().Do(func() {
					e.RemoveEvent(loopevent.GetId())
				})
			}
		}
	}
	wg.Wait()
}

func closechannel(ch *Channel[string]) {
	if ch != nil && !ch.isClosed {
		close(ch.Ch)
		ch.isClosed = true
	}
}

func (e *eventLoop) Toggle(eventFuncs ...EventFunction) {
	for _, v := range eventFuncs {
		if x := slices.Index(e.disabled, v); x != -1 {
			e.logger.Infof("Enabling function %v", v)
			e.disabled = internal.RemoveIndex(e.disabled, x)
		} else {
			e.logger.Infof("Disabling function %v", v)
			e.disabled = append(e.disabled, v)
		}
	}
}

// isScheduledEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - гг (канал самого ивента, канал ивентлупа и context.Done()
func isScheduledEventDone(eventCh, eventLoopCh <-chan bool, ctx context.Context, logger *zap.SugaredLogger) <-chan struct{} {
	result := make(chan struct{}, 1)
	result <- struct{}{}
	select {
	case <-ctx.Done():
		if logger != nil {
			logger.Warnw("Scheduler stopped because of context")
		}
		return result
	case <-eventCh:
		if logger != nil {
			logger.Infow("Scheduler stopped because of event want to stop")
		}
		return result
	case <-eventLoopCh:
		if logger != nil {
			logger.Infow("Scheduler stopped because of event manager commands")
		}
		return result
	default:
		return make(chan struct{})
	}
}

func (e *eventLoop) runScheduledEvent(ctx context.Context, event event.Interface) {
	evntSchedule, _ := event.GetSchedule()
	evntInterval := evntSchedule.GetInterval()
	e.logger.Infow("Scheduled event starting with interval",
		"event", event.GetId(),
		"interval", evntInterval)
	ticker := time.NewTicker(evntInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go event.RunFunction(ctx)
		case <-isScheduledEventDone(evntSchedule.GetQuitChannel(), e.stopScheduler, ctx, e.logger):
			return
		}
	}
}

// ScheduleEvent добавляет ивент в список ивентов-таймеров. Если шедулер запущен - запускает этот ивент.
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID) {

	if _, err := newEvent.GetSchedule(); err != nil {
		e.logger.Errorw(err.Error(), "event", newEvent)
		return
	}

	if isContextDone(ctx) {
		e.logger.Warnw("Can't schedule, context done")
		return
	}

	e.addEvent(INTERVALED, newEvent)

	if e.isSchedulerRunning {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- newEvent.GetId()
	}
}

func (e *eventLoop) StartScheduler(ctx context.Context) {
	if isContextDone(ctx) {
		e.logger.Warnw("Scheduler can't start, context is done")
		return
	}

	if e.isSchedulerRunning {
		e.logger.Warnw("Scheduler is already running")
		return
	}

	for _, evts := range e.events[INTERVALED][0] {
		curEvts := evts
		go e.runScheduledEvent(ctx, curEvts)
	}

	e.isSchedulerRunning = true
	e.logger.Infow("Scheduler started")
}

func (e *eventLoop) StopScheduler() {
	e.logger.Infow("Scheduler stopping...")
	e.mx.Lock()
	defer e.mx.Unlock()
	if len(e.events[INTERVALED][0]) > 0 && e.isSchedulerRunning {
		e.logger.Infow("Send signal to stop")
		e.stopScheduler <- true
	}
	e.isSchedulerRunning = false
	e.logger.Infow("Send signal to stop")
}

func (e *eventLoop) RemoveEvent(id uuid.UUID) bool {

	for eventNameKey, eventNameValue := range e.events {
		for priorKey, priorValue := range eventNameValue {
			for eventIdKey, eventIdValue := range priorValue {
				if eventIdValue.GetId() == id {

					delete(priorValue, eventIdKey)
					if len(priorValue) == 0 {
						delete(eventNameValue, priorKey)
						if len(eventNameValue) == 0 {
							delete(e.events, eventNameKey)
						}
					}

					e.logger.Infow("Event removed from regular events", "event", eventIdValue.GetId())
					return true
				}
			}
		}
	}
	return false
}

func (e *eventLoop) GetEventsByName(eventName string) (result []uuid.UUID, err error) {
	if e.events[eventName] == nil {
		return nil, errors.New(fmt.Sprintf("no such event name: %v", eventName))
	}
	for _, priors := range e.events[eventName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetId())
		}
	}
	return result, nil
}

func (e *eventLoop) LockMutex() {
	e.mx.Lock()
}

func (e *eventLoop) UnlockMutex() {
	e.mx.Unlock()
}
