package api

import (
	"context"
	"errors"
	loggerInternal "eventloop/internal/logger"
	"eventloop/pkg/api/internal/handlers"
	"eventloop/pkg/eventloop"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

var (
	serv      http.Server
	srvLogger *zap.SugaredLogger
)

// StartServer стартует сервер, посылает сигнал в quit по завершению работы сервера. Канал при каждом вызове создаётся
// новый.
func StartServer(level zapcore.Level, quit *chan struct{}) error {
	if quit != nil {
		newChan := make(chan struct{})
		*quit = newChan
		defer close(*quit)
	}

	handlersMap := map[string]handlers.HandlerType{"/events/": handlers.EVENT,
		"/trigger/":   handlers.TRIGGER,
		"/subscribe/": handlers.SUBSCRIBE,
		"/toggle/":    handlers.TOGGLE,
		"/scheduler/": handlers.SCHEDULER}

	var (
		atom *zap.AtomicLevel
		err  error
	)
	srvLogger, atom, err = loggerInternal.Initialize(zapcore.DebugLevel, "logs", "api")
	if err != nil {
		fmt.Println("logger init failed: ", err)
	} else {
		srvLogger.Infof("Server starting...")
	}
	evLoop := eventloop.NewEventLoop(level)
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

	atom.SetLevel(loggerInternal.NormalizeLevel(level))

	servErr := serv.ListenAndServe()
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn("Server closed")
	} else if servErr != nil {
		return servErr
	}
	return nil
}

func StopServer(ctx context.Context) error {
	if err := serv.Shutdown(ctx); err != nil {
		return err
	} else {
		srvLogger.Infof("Server stopped.")
	}
	return nil
}
