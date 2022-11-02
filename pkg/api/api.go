package api

import (
	"context"
	"errors"
	loggerInternal "eventloop/internal/logger"
	"eventloop/pkg/api/internal/handlers"
	"eventloop/pkg/eventloop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

var (
	serv      http.Server
	srvLogger *zap.SugaredLogger
)

func StartServer(level zapcore.Level) {
	handlersMap := map[string]handlers.HandlerType{"/events/": handlers.EVENT,
		"/trigger/":   handlers.TRIGGER,
		"/subscribe/": handlers.SUBSCRIBE,
		"/toggle/":    handlers.TOGGLE,
		"/scheduler/": handlers.SCHEDULER}

	//servCtx, servCancel := context.WithCancel(outerContext)
	var atom *zap.AtomicLevel
	srvLogger, atom = loggerInternal.Initialize(zapcore.DebugLevel, "logs", "api")
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
	//servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn("Server closed")
	} else if servErr != nil {
		srvLogger.Errorf("Error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func StopServer(ctx context.Context) {
	if err := serv.Shutdown(ctx); err != nil {
		srvLogger.Errorf("Server shutdown error: %v", err)
	} else {
		srvLogger.Infof("Server stopped.")
	}
}
