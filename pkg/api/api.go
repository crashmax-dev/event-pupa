package api

import (
	"context"
	"errors"
	"eventloop/pkg/eventloop"
	"fmt"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
)

var (
	evLoop eventloop.Interface
	ctx    context.Context
	cancel context.CancelFunc
	events []ApiEvents
)

func StartServer(level zapcore.Level) {
	ctx, cancel = context.WithCancel(context.Background())
	evLoop = eventloop.NewEventLoop(level)

	http.HandleFunc("/events/", eventHandler)

	servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if servErr != nil {
		fmt.Printf("error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func StopServer() {
	defer cancel()
}
