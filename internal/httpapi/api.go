package httpapi

import (
	"context"
	"errors"
	"eventloop/internal/httpapi/handler"
	"eventloop/internal/httpapi/helper"
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
func StartServer(port int, evLoop eventloop.Interface, srvLogger logger.Interface) error {
	helper.APIMessageSetPrefix(_APIPREFIX)

	handlersMap := map[string]handler.Type{"/events/": handler.EVENT,
		"/trigger/":   handler.TRIGGER,
		"/subscribe/": handler.SUBSCRIBE,
		"/toggle/":    handler.TOGGLE,
		"/scheduler/": handler.SCHEDULER}

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
		srvLogger.Warn(helper.APIMessage("Server closed"))
	} else if servErr != nil {
		return servErr
	}

	return nil
}

func StopServer(ctx context.Context, srvLogger logger.Interface) error {
	err := serv.Shutdown(ctx)
	if err != nil {
		return err
	}
	srvLogger.Infof(helper.APIMessage("Server stopped."))
	return nil
}
