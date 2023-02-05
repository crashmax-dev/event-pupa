package eventloop

import (
	"fmt"

	"gitlab.com/YSX/eventloop/pkg/eventloop/internal"
	"gitlab.com/YSX/eventloop/pkg/logger"
	"golang.org/x/exp/slices"
)

func toggle[T comparable](container *[]T, logger logger.Interface, items ...T) (result string) {
	c := *container
	for _, v := range items {
		if result != "" {
			result += " | "
		}
		// Включение
		if x := slices.Index(c, v); x != -1 {
			result += fmt.Sprintf("Enabling %v", v)
			logger.Info(result)
			c = internal.RemoveSliceItemByIndex(c, x)
		} else { // Выключение
			result += fmt.Sprintf("Disabling %v", v)
			logger.Info(result)
			c = append(c, v)
		}
	}
	*container = c
	return
}

// ToggleEventLoopFunc включает/выключает функции менеджера событий (REGISTER и TRIGGER).
// При попытке использования этих функций возвращается ошибка.
// Функции можно включить обратно простым прокидыванием тех же параметров, в зависимости от того что надо включить.
func (e *eventLoop) ToggleEventLoopFuncs(eventFuncs ...EventFunction) string {
	return toggle(&e.disabled, e.logger, eventFuncs...)
}

// ToggleTriggers включает/выключает отдельные триггеры.
// При попытке использования этих триггеров возвращается ошибка.
// Триггеры можно включить обратно простым прокидыванием тех же параметров, в зависимости от того что надо включить.
func (e *eventLoop) ToggleTriggers(triggerNames ...string) (result string) {
	for _, name := range triggerNames {
		if result != "" {
			result += " | "
		}
		// Включение
		if !e.events.IsTriggerEnabled(name) {
			result += fmt.Sprintf("Enabling %v", name)
			e.logger.Info(result)
			e.events.ToggleTrigger(name, true)
		} else { // Выключение
			result += fmt.Sprintf("Disabling %v", name)
			e.logger.Info(result)
			e.events.ToggleTrigger(name, false)
		}
	}
	return
}
