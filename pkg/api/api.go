package api

import (
	"context"
	"errors"
	loggerInternal "eventloop/internal/logger"
	"eventloop/pkg/api/internal"
	"eventloop/pkg/eventloop"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

var (
	servCancel context.CancelFunc
)

func StartServer(outerContext context.Context, level zapcore.Level) {
	//servCtx, servCancel := context.WithCancel(outerContext)
	srvLogger, atom := loggerInternal.Initialize(zapcore.DebugLevel, "logs", "api")
	evLoop := eventloop.NewEventLoop(level)

	eh := new(internal.EventHandler)
	eh.Base = internal.NewBaseHandler(srvLogger, evLoop)
	th := new(internal.TriggerHandler)
	th.Base = internal.NewBaseHandler(srvLogger, evLoop)
	sh := new(internal.SubscribeHandler)
	sh.Base = internal.NewBaseHandler(srvLogger, evLoop)
	tgh := new(internal.ToggleHandler)
	tgh.Base = internal.NewBaseHandler(srvLogger, evLoop)
	sch := new(internal.SchedulerHandler)
	sch.Base = internal.NewBaseHandler(srvLogger, evLoop)

	mux := http.NewServeMux()
	mux.Handle("/events/", eh)
	mux.Handle("/trigger/", th)
	mux.Handle("/subscribe/", sh)
	mux.Handle("/toggle/", tgh)
	mux.Handle("/scheduler/", sch)

	s := &http.Server{
		Addr:         ":8090",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.Handle("/events/", eh)
	http.Handle("/trigger/", new(internal.TriggerHandler))
	http.Handle("/subscribe/", new(internal.SubscribeHandler))
	http.Handle("/toggle/", new(internal.ToggleHandler))
	http.Handle("/scheduler/", new(internal.SchedulerHandler))

	atom.SetLevel(loggerInternal.NormalizeLevel(level))

	servErr := s.ListenAndServe()
	//servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn("Server closed")
	} else if servErr != nil {
		srvLogger.Errorf("Error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func StopServer() {
	defer servCancel()
}
