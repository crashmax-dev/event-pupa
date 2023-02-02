package eventloop

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"eventloop/internal/loggerImplementation"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/event/subscriber"
	"eventloop/pkg/eventloop/internal"
	"eventloop/pkg/eventloop/internal/eventsContainer"
	loggerEventLoop "eventloop/pkg/logger"
	"golang.org/x/exp/slices"
)

type EventFunction string

const (
	TRIGGER  EventFunction = "TRIGGER"
	REGISTER EventFunction = "REGISTER"
)

// eventLoop представляет собой менеджер событий. Позволяет использовать как классические события с названиями для
// каждого, так и одноразовые, выполняющиеся с определённым интервалом. Также можно задавать приоритет обычным событиям.
// Для использования нужно создавать event.
type eventLoop struct {
	events eventsContainer.Interface
	mx     *sync.RWMutex

	disabled []EventFunction

	logger loggerEventLoop.Interface
}

// NewEventLoop - конструктор для менеджера событий. Инициализирует новый Event Loop.
// Для level рекомендуются DebugLevel для Dev, и ErrorLevel для Prod. Можно указать любой уровень, он нормализуется в
// Debug и Error, в зависимости от велчины уровня.
func NewEventLoop(level string) Interface {
	elLogger, err := loggerImplementation.NewLogger(level, "logs", "")
	if err != nil {
		fmt.Printf("logger init error: %v", err)
	}
	return &eventLoop{
		mx:     &sync.RWMutex{},
		events: eventsContainer.New(),
		logger: elLogger,
	}
}

func (e *eventLoop) RegisterEvent(
	ctx context.Context,
	newEvents ...event.Interface,
) (errReturn error) {
	e.mx.Lock()
	defer e.mx.Unlock()
	for _, evnt := range newEvents {
		if ctxErr := e.checkContext(ctx, "can't register event, context is done",
			"events", evnt.GetUUID(),
			"trigger", evnt.GetTriggerName(),
		); ctxErr != nil {
			internal.WriteToExecCh(ctx, "")
			errReturn = internal.WrapError(errReturn, ctxErr)
			continue
		}
		// Если выключено добавление - не добавляем
		if slices.Contains(e.disabled, REGISTER) {
			errStr := "register disabled, can't register event"
			e.logger.Warnw(
				errStr,
				"event", evnt.GetUUID(),
			)
			internal.WriteToExecCh(ctx, "")
			errReturn = internal.WrapError(errReturn, errors.New(errStr))
			continue
		}

		// ON
		if triggerName := evnt.GetTriggerName(); triggerName != "" {
			e.events.AddEvent(evnt)
			e.logger.Debugw(
				"Event added", "triggerName", evnt.GetTriggerName(), "eventId",
				evnt.GetUUID(),
			)
		} else if intervalComp, intervalErr := evnt.Interval(); intervalErr == nil { // INTERVAL
			e.events.AddEvent(evnt)
			e.logger.Debugw("Event added", "interval", intervalComp.GetDuration())
		} else if afterComp, afterErr := evnt.After(); afterErr == nil { // AFTER
			e.events.AddEvent(evnt)
			e.logger.Debugw(
				"Event added", "start_time", afterComp.GetDuration(),
				"eventId",
				evnt.GetUUID(),
			)
		} else {
			errStr := "event must be at least ON, INTERVAL or AFTER"
			errNew := fmt.Errorf(errStr)
			e.logger.Debugw(errStr, "eventId", evnt.GetUUID())
			errReturn = internal.WrapError(errReturn, errNew)
			continue
		}

		internal.WriteToExecCh(ctx, "")
	}
	return errReturn
}

// Subscribe подписывает список событий listeners на список событий triggers. Само событие триггерится с помощью Trigger/
// В случае передачи контекста с дедлайном или таймаутом, если контекст ещё живой, подписанные события всё равно
// выполнятся один раз в случае триггера.
func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error {
	subCtx := loggerEventLoop.WithLogger(ctx, e.logger)

	defer internal.WriteToExecCh(ctx, "")

	if isContextDone(subCtx) {
		errStr := "can't subscribe, context is done"
		e.logger.Warnw(errStr,
			"triggers", triggers,
			"listeners", listeners,
		)
		return errors.New(errStr)
	}
	for _, listener := range listeners {
		listenerSubComponent, _ := listener.Subscriber()
		for _, t := range triggers {
			ch := make(chan subscriber.SubChInfo, 1)
			generalClosedInfo := false
			listenerSubComponent.AddChannel(t.GetUUID(), ch, &generalClosedInfo)
			e.logger.Infow("Event subscribed", "listenerSubComponent", t.GetUUID(), "listener", listener.GetUUID())
			tSub, _ := t.Subscriber()
			tSub.AddChannel(listener.GetUUID(), ch, &generalClosedInfo)
		}
		e.addEvent("", listener)
		// Запскаем ждуна для слушателя, когда триггеры сработают, и срабатываем сами
		go e.runnerListener(subCtx, listener)
	}
	for _, t := range triggers {
		go e.runnerTrigger(subCtx, t)
	}
	return nil
}

func (e *eventLoop) addEvent(triggerName string, newEvent event.Interface) {
	e.events.TriggerName(triggerName).Priority(newEvent.GetPriority()).AddEvent(newEvent)
	for _, t := range newEvent.GetTypes() {
		if e.eventsByType[t] == nil {
			e.eventsByType[t] = []event.Interface{}
		}
		e.eventsByType[t] = append(e.eventsByType[t], newEvent)
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

// Горутина события-слушателя
func (e *eventLoop) runnerListener(ctx context.Context, v event.Interface) {
	var (
		subComponent, _ = v.Subscriber()
		channels        = subComponent.Channels()
	)

	if subComponent.IsRunning() {
		return
	}
	subComponent.SetIsRunning(true)

	exitChan := isEventDone(ctx, subComponent.Exit(), e.logger)
	for {
		i := len(channels)
		if i == 0 {
			subComponent.SetIsRunning(false)
			return
		}
		subComponent.LockMutex()
		for id, ch := range channels {
			select {
			case <-exitChan:
				for _, closeCh := range channels {
					closeCh.SetIsClosed()
				}
				subComponent.SetIsRunning(false)
				return
			case <-ch.GetInfoCh():
				if ch.IsClosed() {
					delete(channels, id)
				}
				logTxt := fmt.Sprintf(
					"Reading channel from %v [%v/%v]", id, i,
					len(channels),
				)
				e.logger.Debugw(logTxt, "event", v.GetUUID())
				i--
			}
			if i <= 0 {
				if i < 0 {
					panic("too much channels waited")
				}
				e.logger.Infow("Subscriber event fired", "event", v.GetUUID())
				v.RunFunction(ctx)
			}
		}
		subComponent.UnlockMutex()
	}
}

// Горутина события-триггера
func (e *eventLoop) runnerTrigger(ctx context.Context, v event.Interface) {
	var (
		subComponent, _ = v.Subscriber()
		channels        = subComponent.Channels()
	)

	if subComponent.IsRunning() {
		return
	}

	subComponent.SetIsRunning(true)

	exitChan := isEventDone(ctx, subComponent.Exit(), e.logger)
	e.logger.Debugw("Runner trigger started", "eventId", v.GetUUID())
	for {
		select {
		case <-exitChan:
			for _, closeCh := range channels {
				closeCh.SetIsClosed()
			}
			subComponent.SetIsRunning(false)
			return
		case <-subComponent.ChanTrigger():
			e.logger.Debugw("TriggerEvent activated", "eventId", v.GetUUID())
			subComponent.LockMutex()
			i := 1
			for id, chnl := range channels {
				if chnl.IsClosed() {
					delete(channels, id)
				}
				logTxt := fmt.Sprintf(
					"Writing channel for %v [%v/%v]", id, i,
					len(channels),
				)
				e.logger.Debugw(logTxt, "event", v.GetUUID())
				chnl.GetInfoCh() <- subscriber.TriggerListener
				i++
			}
			subComponent.UnlockMutex()
		}
	}
}

// Trigger вызывает событие с определённым triggerName. Функция ждёт выполнения всех добавленных на событие функций,
// поэтому синхронный вызов заблокирует родительский цикл выполнения программы.
// В Ch пишется резульат выполнения каждого триггера, после использования канал закрывается. Поэтому для каждого вызова
// нужно создавать новый channelEx
func (e *eventLoop) Trigger(ctx context.Context, triggerName string) error {
	errFunc := func(msg string) error {
		e.logger.Warnw(
			msg,
			"eventname", triggerName,
		)
		internal.WriteToExecCh(ctx, "")
		return errors.New(msg)
	}

	triggerCtx := loggerEventLoop.WithLogger(ctx, e.logger)

	var deferErr error

	if ctxErr := e.checkContext(triggerCtx,
		"can't trigger event, context is done",
		"triggerName", triggerName,
	); ctxErr != nil {
		internal.WriteToExecCh(ctx, "")
		return ctxErr
	}

	// Выключен ли Триггер
	if slices.Contains(e.disabled, TRIGGER) {
		return errFunc("can't trigger event, trigger function is disabled")
	}

	if !e.events.IsTriggerEnabled(triggerName) {
		return errFunc("can't trigger event, trigger name is disabled")
	}

	e.logger.Debugw("Trying to get mutex", "triggerName", triggerName)
	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("ChanTrigger event", "triggerName", triggerName)

	// Run before global events
	e.triggerEventFuncList(triggerCtx, e.events.EventsByTrigger(string(BEFORE_TRIGGER))...)

	eventsByPriority := e.events.GetPrioritySortedEventsByTrigger(triggerName)
	for _, ev := range eventsByPriority {
		e.logger.Debugw("Start runFunc goroutine", "eventId", ev.GetUUID())
		go e.triggerEventFunc(triggerCtx, ev)

		if once, err := ev.Once(); err == nil {
			once.Do(
				func() {
					e.RemoveEventByUUIDs(ev.GetUUID())
				},
			)
		}
	}
	if len(eventsByPriority) == 0 {
		internal.WriteToExecCh(triggerCtx, "")
	}

	// Run after global events
	e.triggerEventFuncList(triggerCtx, e.events.EventsByTrigger(string(AFTER_TRIGGER))...)

	return deferErr
}

func (e *eventLoop) triggerEventFuncList(ctx context.Context, list ...event.Interface) {
	for _, listItem := range list {
		listItem.RunFunction(ctx)
	}
}

func (e *eventLoop) triggerEventFunc(ctx context.Context, ev event.Interface) {
	if after, afterErr := ev.After(); afterErr == nil {
		e.logger.Debugw("Waiting for start", "eventId", ev.GetUUID(), "time", after.GetDuration())
		after.Wait()
	}

	if interval, err := ev.Interval(); err == nil {
		if interval.IsRunning() {
			interval.GetQuitChannel() <- true
			internal.WriteToExecCh(ctx, "")
		} else {
			e.logger.Debugw("Run scheduled", "eventId", ev.GetUUID())
			e.runScheduledEvent(ctx, ev)
		}
	} else {
		ev.RunFunction(ctx)
	}
}

// isEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - всё, гг (либо канал самого ивента, канал ивентлупа и context.Done()
func isEventDone[T any](ctx context.Context,
	eventCh <-chan T,
	logger loggerEventLoop.Interface,
) <-chan struct{} {
	result := make(chan struct{})
	go func(eventCh <-chan T) {
		select {
		case <-ctx.Done():
			if logger != nil {
				logger.Warnw("Event stopped because of context")
			}
			result <- struct{}{}
		case <-eventCh:
			if logger != nil {
				logger.Infow("Event stopped because of event want to stop")
			}
			result <- struct{}{}
		}
	}(eventCh)
	return result
}

func (e *eventLoop) runScheduledEvent(ctx context.Context, ev event.Interface) {
	schedCtx, cancel := context.WithCancel(ctx)
	intervalComponent, _ := ev.Interval()
	evntInterval := intervalComponent.GetDuration()
	e.logger.Infow("Scheduled ev starting with interval",
		"ev", ev.GetUUID(),
		"interval", evntInterval,
	)
	ticker := time.NewTicker(evntInterval)
	intervalComponent.SetRunning(true)

	defer cancel()
	defer intervalComponent.SetRunning(false)
	defer ticker.Stop()

	exitChan := isEventDone(schedCtx, intervalComponent.GetQuitChannel(), e.logger)
	for {
		select {
		case <-ticker.C:
			go func(ev event.Interface) {
				ev.RunFunction(schedCtx)
				if once, onceErr := ev.Once(); onceErr == nil {
					once.Do(
						func() {
							cancel()
						},
					)
					return
				}
			}(ev)

		case <-exitChan:
			return
		}
	}
}

func (e *eventLoop) RemoveEventByUUIDs(uUIDs ...string) []string {
	return e.events.RemoveEventByUUIDs(uUIDs...)
}

func (e *eventLoop) RemoveTriggers(triggers ...string) []string {
	return e.events.RemoveTriggers(triggers...)
}

// GetAttachedEvents возвращает все события, прикреплённые к triggerName
func (e *eventLoop) GetAttachedEvents(triggerName string) (result []event.Interface) {
	return e.events.EventsByTrigger(triggerName)
}

func (e *eventLoop) GetEventsByType(eventType string) (result []string, errReturn error) {
	t, err := event.AsType(eventType)
	if err != nil {
		return nil, err
	}
	result = make([]string, 0, len(e.eventsByType[t]))
	for _, v := range e.eventsByType[t] {
		result = append(result, v.GetUUID())
	}
	return result, nil
}

func (e *eventLoop) GetTriggerNames() AllTriggers {
	ReturnIriggers.userTriggers = e.events.GetAll()
	return ReturnIriggers
}

func (e *eventLoop) Sync() error {
	return e.logger.Sync()
}

func (e *eventLoop) checkContext(ctx context.Context, message string, loggerArgs ...string) error {
	if isContextDone(ctx) {
		errStr := fmt.Sprintf("%v (%v)", message, ctx.Err())
		e.logger.Warnw(
			errStr,
			loggerArgs,
		)
		return errors.New(errStr)
	}
	return nil
}
