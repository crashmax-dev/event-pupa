package eventloop

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"eventloop/internal/logger"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal"
	"eventloop/pkg/eventloop/internal/eventslist"
	loggerEventLoop "eventloop/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

// eventLoop представляет собой менеджер событий. Позволяет использовать как классические события с названиями для
// каждого, так и одноразовые, выполняющиеся с определённым интервалом. Также можно задавать приоритет обычным событиям.
// Для использования нужно создавать event.
type eventLoop struct {
	events eventslist.Interface
	mx     *sync.RWMutex

	disabled []EventFunction

	logger loggerEventLoop.Interface
}

// NewEventLoop - конструктор для менеджера событий. Инициализирует новый Event Loop.
// Для level рекомендуются DebugLevel для Dev, и ErrorLevel для Prod. Можно указать любой уровень, он нормализуется в
// Debug и Error, в зависимости от велчины уровня.
func NewEventLoop(level string) Interface {
	elLogger, err := logger.NewLogger(level, "logs", "")
	if err != nil {
		fmt.Printf("logger init error: %v", err)
	}

	return &eventLoop{
		mx:     &sync.RWMutex{},
		events: eventslist.New(),
		logger: elLogger,
	}
}

func (e *eventLoop) RegisterEvent(ctx context.Context, newEvent event.Interface) error {
	if ctxErr := e.checkContext(ctx, "can't register event, context is done",
		"event", newEvent.GetID().String(),
		"trigger", newEvent.GetTriggerName()); ctxErr != nil {
		internal.WriteToExecCh(ctx, "")
		return ctxErr
	}

	// Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, REGISTER) {
		errStr := "register disabled, can't register event"
		e.logger.Warnw(errStr,
			"event", newEvent.GetID())
		internal.WriteToExecCh(ctx, "")
		return errors.New(errStr)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	// ON
	if triggerName := newEvent.GetTriggerName(); triggerName != "" {
		if slices.Contains(restrictedEvents, eventLoopSystemEvent(triggerName)) {
			errStr := fmt.Sprintf("Trigger name %v is reserved", triggerName)
			e.logger.Warnf("Trigger name %v is reserved", triggerName)
			internal.WriteToExecCh(ctx, "")
			return errors.New(errStr)
		}
		e.addEvent(newEvent.GetTriggerName(), newEvent)

		e.logger.Debugw("Event added", "triggerName", newEvent.GetTriggerName())
	} else if intervalComp, intervalErr := newEvent.Interval(); intervalErr == nil { // INTERVAL
		e.addEvent(string(INTERVALED), newEvent)
		e.logger.Debugw("Event added", "interval", intervalComp.GetDuration())
	} else if afterComp, afterErr := newEvent.After(); afterErr == nil {
		e.addEvent(string(AFTER), newEvent)
		e.logger.Debugw("Event added", "start_time", afterComp.GetDuration())
	} else {
		return errors.New("event must be at least ON, INTERVAL or AFTER")
	}

	internal.WriteToExecCh(ctx, "")
	return nil
}

// Subscribe подписывает список событий listeners на список событий triggers. Само событие триггерится с помощью Trigger/
// В случае передачи контекста с дедлайном или таймаутом, если контекст ещё живой, подписанные события всё равно
// выполнятся один раз в случае триггера.
func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error {
	subCtx := context.WithValue(ctx, internal.LOGGER_CTX_KEY, e.logger)

	defer internal.WriteToExecCh(ctx, "")

	if isContextDone(subCtx) {
		errStr := "can't subscribe, context is done"
		e.logger.Warnw(errStr,
			"triggers", triggers,
			"listeners", listeners)
		return errors.New(errStr)
	}
	for _, listener := range listeners {
		listenerSubComponent := listener.Subscriber()
		for _, t := range triggers {
			ch := make(chan int, 1)
			listenerSubComponent.AddChannel(ch)
			e.logger.Infow("Event subscribed", "listenerSubComponent", t.GetID(), "listener", listener.GetID())
			tSub := t.Subscriber()
			tSub.SetIsTrigger(true)
			tSub.AddChannel(ch)
		}

		// Запскаем ждуна, когда триггеры сработают, и срабатываем сами
		go func(ctx context.Context, v event.Interface) {
			subComponent := v.Subscriber()
			for {
				select {
				case <-ctx.Done():
					e.logger.Infow("Stop listening because of context", "eventId", v.GetID())
					return
				default:
					subComponent.LockMutex()
					channels := subComponent.GetChannels()
					for i, ch := range channels {
						logTxt := fmt.Sprintf("Reading channel %v of %v", i+1, len(channels))
						e.logger.Debugw(logTxt, "event", v.GetID())
						<-ch
					}
					e.logger.Infow("Subscriber event fired", "event", v.GetID())
					v.RunFunction(ctx)
					subComponent.UnlockMutex()
				}
			}
		}(subCtx, listener)
	}
	return nil
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
	e.events.EventName(eventName).Priority(newEvent.GetPriority()).AddEvent(newEvent)
}

// Trigger вызывает событие с определённым eventName. Функция ждёт выполнения всех добавленных на событие функций,
// поэтому синхронный вызов заблокирует родительский цикл выполнения программы.
// В Ch пишется резульат выполнения каждого триггера, после использования канал закрывается. Поэтому для каждого вызова
// нужно создавать новый channelEx
func (e *eventLoop) Trigger(ctx context.Context, triggerName string) error {
	triggerCtx := context.WithValue(ctx, internal.LOGGER_CTX_KEY, e.logger)

	var deferErr error

	if ctxErr := e.checkContext(triggerCtx,
		"can't trigger event, context is done",
		"triggerName", triggerName); ctxErr != nil {
		internal.WriteToExecCh(ctx, "")
		return ctxErr
	}

	// Выключен ли Триггер
	if slices.Contains(e.disabled, TRIGGER) {
		str := "can't trigger event, trigger is disabled"
		e.logger.Warnw(str,
			"eventname", triggerName)
		internal.WriteToExecCh(ctx, "")
		return errors.New(str)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("Trigger event", "triggerName", triggerName)

	// Run before global events
	e.triggerEventFuncList(triggerCtx, e.events.EventName(string(BEFORE_TRIGGER)).Priority(BEFORE_PRIORITY).List())

	// Run before events
	e.triggerEventFuncList(triggerCtx, e.events.EventName(triggerName).Priority(BEFORE_PRIORITY).List())

	keys := e.events.EventName(triggerName).GetKeys()
	if priorIndex := len(keys) - 1; priorIndex >= 0 && keys[priorIndex] >= 0 {
		for priorIndex = len(keys) - 1; priorIndex >= 0 && keys[priorIndex] >= 0; priorIndex-- {
			priority := keys[priorIndex]
			for _, loopevent := range e.events.EventName(triggerName).Priority(priority).List() {
				go e.triggerEventFunc(triggerCtx, loopevent)

				if once, err := loopevent.Once(); err == nil {
					once.Do(func() {
						e.RemoveEventByUUIDs(loopevent.GetID())
					})
				}
			}
		}
	} else {
		internal.WriteToExecCh(triggerCtx, "")
	}

	// Run after global events
	e.triggerEventFuncList(triggerCtx, e.events.EventName(string(AFTER_TRIGGER)).Priority(AFTER_PRIORITY).List())

	// Run after events
	e.triggerEventFuncList(triggerCtx, e.events.EventName(triggerName).Priority(AFTER_PRIORITY).List())

	return deferErr
}

func (e *eventLoop) triggerEventFuncList(ctx context.Context, list eventslist.EventIdsList) {
	for _, listItem := range list {
		listItem.RunFunction(ctx)
	}
}

func (e *eventLoop) triggerEventFunc(ctx context.Context, ev event.Interface) {
	if after, afterErr := ev.After(); afterErr == nil {
		after.Wait()
	}

	if interval, err := ev.Interval(); err == nil {
		if interval.IsRunning() {
			interval.GetQuitChannel() <- true
			internal.WriteToExecCh(ctx, "")
		} else {
			e.runScheduledEvent(ctx, ev)
		}
	} else {
		ev.RunFunction(ctx)
	}
}

// Toggle выключает функции менеджера событий, ON и TRIGGER. При попытке использования этих функций выводится ошибка.
// Функции можно включить обратно простым прокидыванием тех же параметров, в зависимости от того что надо включить.
func (e *eventLoop) Toggle(eventFuncs ...EventFunction) (result string) {
	for _, v := range eventFuncs {
		if result != "" {
			result += " | "
		}
		// Включение
		if x := slices.Index(e.disabled, v); x != -1 {
			result += fmt.Sprintf("Enabling function %v", v)
			e.logger.Info(result)
			e.disabled = internal.RemoveSliceItemByIndex(e.disabled, x)
		} else { // Выключение
			result += fmt.Sprintf("Disabling function %v", v)
			e.logger.Info(result)
			e.disabled = append(e.disabled, v)
		}
	}
	return
}

// isScheduledEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - всё, гг (либо канал самого ивента, канал ивентлупа и context.Done()
func isScheduledEventDone(ctx context.Context, eventCh <-chan bool, logger loggerEventLoop.Interface) <-chan struct{} {
	waiter := make(chan struct{})
	result := make(chan struct{})
	go func(ch chan<- struct{}) {
		waiter <- struct{}{}
		for {
			select {
			case <-ctx.Done():
				if logger != nil {
					logger.Warnw("Scheduled event stopped because of context")
				}
				result <- struct{}{}
			case <-eventCh:
				if logger != nil {
					logger.Infow("Scheduled event stopped because of event want to stop")
				}
				result <- struct{}{}
			}
		}
	}(result)
	<-waiter
	return result
}

func (e *eventLoop) runScheduledEvent(ctx context.Context, ev event.Interface) {
	schedCtx, cancel := context.WithCancel(ctx)
	intervalComponent, _ := ev.Interval()
	evntInterval := intervalComponent.GetDuration()
	e.logger.Infow("Scheduled ev starting with interval",
		"ev", ev.GetID(),
		"interval", evntInterval)
	ticker := time.NewTicker(evntInterval)
	intervalComponent.SetRunning(true)

	defer cancel()
	defer intervalComponent.SetRunning(false)
	defer ticker.Stop()

	exitChan := isScheduledEventDone(schedCtx, intervalComponent.GetQuitChannel(), e.logger)
	for {
		select {
		case <-ticker.C:
			go func(ev event.Interface) {
				ev.RunFunction(schedCtx)
				if once, onceErr := ev.Once(); onceErr == nil {
					once.Do(func() {
						cancel()
					})
					return
				}
			}(ev)

		case <-exitChan:
			fmt.Println("Boom")
			return
		}
	}
}

func (e *eventLoop) RemoveEventByUUIDs(ids ...uuid.UUID) []uuid.UUID {
	return e.events.RemoveEventByUUIDs(ids...)
}

// GetAttachedEvents возвращает все события, прикреплённые к eventName
func (e *eventLoop) GetAttachedEvents(eventName string) (result []uuid.UUID, err error) {
	return e.events.GetEventIdsByName(eventName)
}

func (e *eventLoop) Sync() error {
	return e.logger.Sync()
}

func (e *eventLoop) checkContext(ctx context.Context, message string, loggerArgs ...string) error {
	if isContextDone(ctx) {
		errStr := fmt.Sprintf("%v (%v)", message, ctx.Err())
		e.logger.Warnw(errStr,
			loggerArgs)
		return errors.New(errStr)
	}
	return nil
}
