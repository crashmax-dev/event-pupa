package httpApi

import (
	"context"
	"errors"
	"eventloop/internal/httpApi/handler"
	"eventloop/internal/httpApi/helper"
	"eventloop/internal/logger"
	"eventloop/pkg/eventloop"
	"net/http"
	"strconv"
	"time"
)

const _APIPREFIX = "[API] "

var (
	serv http.Server
)

// StartServer стартует API сервер для доступа к Event Loop. Функция блокирующая
func StartServer(port int, srvLogger logger.Interface) error {

	helper.ApiMessageSetPrefix(_APIPREFIX)

	handlersMap := map[string]handler.HandlerType{"/events/": handler.EVENT,
		"/trigger/":   handler.TRIGGER,
		"/subscribe/": handler.SUBSCRIBE,
		"/toggle/":    handler.TOGGLE,
		"/scheduler/": handler.SCHEDULER}

	evLoop := eventloop.NewEventLoop(srvLogger.Level())
	mux := http.NewServeMux()
	for k, v := range handlersMap {
		mux.Handle(k, handler.NewHandler(v, srvLogger, evLoop))
	}

	serv = http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	servErr := serv.ListenAndServe()
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn(helper.ApiMessage("Server closed"))
	} else if servErr != nil {
		return servErr
	}

	return nil
}

func StopServer(ctx context.Context, srvLogger logger.Interface) error {
	if err := serv.Shutdown(ctx); err != nil {
		return err
	} else {
		srvLogger.Infof(helper.ApiMessage("Server stopped."))
	}
	return nil
}
