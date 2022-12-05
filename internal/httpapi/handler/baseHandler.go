package handler

import (
	"eventloop/pkg/eventloop"
	"eventloop/pkg/logger"
)

type baseHandler struct {
	logger logger.Interface
	evLoop eventloop.Interface
}
