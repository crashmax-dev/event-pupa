package handler

import (
	"net/http"

	"eventloop/pkg/eventloop"
	"eventloop/pkg/logger"
)

type Type uint8

const (
	EVENT Type = iota
	TRIGGER
	SUBSCRIBE
	TOGGLE
	SCHEDULER
)

// NewHandler создаёт новое событие типа ht, logger и evloop для всех хэндлеров одного сервера должны быть одни и те же
func NewHandler(ht Type, logger logger.Interface, evLoop eventloop.Interface) http.Handler {
	bh := baseHandler{logger: logger, evLoop: evLoop}
	var handlerMap = map[Type]http.Handler{
		EVENT:     &eventHandler{bh},
		TRIGGER:   &triggerHandler{bh},
		SUBSCRIBE: &subscribeHandler{baseHandler: bh},
		TOGGLE:    &toggleHandler{bh},
		SCHEDULER: &schedulerHandler{bh},
	}

	return handlerMap[ht]
}
