package handler

import (
	"gitlab.com/YSX/eventloop/pkg/eventloop"
	"gitlab.com/YSX/eventloop/pkg/logger"
)

type baseHandler struct {
	logger logger.Interface
	evLoop eventloop.Interface
}
