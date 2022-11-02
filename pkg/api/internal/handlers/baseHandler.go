package handlers

import (
	"eventloop/pkg/eventloop"
	"go.uber.org/zap"
)

type baseHandler struct {
	logger *zap.SugaredLogger
	evLoop eventloop.Interface
}
