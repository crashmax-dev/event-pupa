package handlers

import (
	"eventloop/internal/logger"
	"eventloop/pkg/eventloop"
)

type baseHandler struct {
	logger logger.Interface
	evLoop eventloop.Interface
}
