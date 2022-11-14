package handlers

import (
	"eventloop/pkg/eventloop"
	"go.uber.org/zap"
	"net/http"
)

type HandlerType uint8

const (
	EVENT HandlerType = iota
	TRIGGER
	SUBSCRIBE
	TOGGLE
	SCHEDULER
)

// NewHandler создаёт новое событие типа ht, logger и evloop для всех хэндлеров одного сервера должны быть одни и те же
func NewHandler(ht HandlerType, logger *zap.SugaredLogger, evLoop eventloop.Interface) http.Handler {
	bh := baseHandler{logger: logger, evLoop: evLoop}
	var handlerMap = map[HandlerType]http.Handler{
		EVENT:     &eventHandler{bh},
		TRIGGER:   &triggerHandler{bh},
		SUBSCRIBE: &subscribeHandler{baseHandler: bh},
		TOGGLE:    &toggleHandler{bh},
		SCHEDULER: &schedulerHandler{bh},
	}

	return handlerMap[ht]
}
