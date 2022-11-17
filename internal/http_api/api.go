package http_api

import (
	"context"
	"errors"
	"eventloop/internal/http_api/internal"
	"eventloop/internal/http_api/internal/handlers"
	"eventloop/internal/logger"
	"eventloop/pkg/eventloop"
	"net/http"
	"time"
)

const _APIPREFIX = "[API] "

var (
	serv http.Server
)

// StartServer стартует API сервер для доступа к Event Loop
func StartServer(srvLogger logger.Interface) error {

	internal.ApiMessageSetPrefix(_APIPREFIX)

	handlersMap := map[string]handlers.HandlerType{"/events/": handlers.EVENT,
		"/trigger/":   handlers.TRIGGER,
		"/subscribe/": handlers.SUBSCRIBE,
		"/toggle/":    handlers.TOGGLE,
		"/scheduler/": handlers.SCHEDULER}

	evLoop := eventloop.NewEventLoop(srvLogger.Level())
	mux := http.NewServeMux()
	for k, v := range handlersMap {
		mux.Handle(k, handlers.NewHandler(v, srvLogger, evLoop))
	}

	serv = http.Server{
		Addr:         ":8090",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	servErr := serv.ListenAndServe()
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn(internal.ApiMessage("Server closed"))
	} else if servErr != nil {
		return servErr
	}

	return nil
}

func StopServer(ctx context.Context, srvLogger logger.Interface) error {
	if err := serv.Shutdown(ctx); err != nil {
		return err
	} else {
		srvLogger.Infof(internal.ApiMessage("Server stopped."))
	}
	return nil
}
