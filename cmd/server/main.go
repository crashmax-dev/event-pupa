package main

import (
	"context"
	"eventloop/internal/httpapi"
	"eventloop/internal/logger"
	"eventloop/pkg/eventloop"
	loggerInterface "eventloop/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	_LOGLEVEL = "debug"
	_PORT     = 8090
)

// @title Event Loop API
// @version 1.0
// @description TODO
// @contact.url TODO
// @host localhost:%v
func inputMonitor(sc chan<- os.Signal) {
	var input string
	for strings.ToLower(input) != "stop" {
		_, _ = fmt.Scanln(&input)
	}
	sc <- syscall.SIGTERM
}

func main() {
	srvLogger, err := initLogger()
	evLoop := eventloop.NewEventLoop(srvLogger.Level())

	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		errServer := httpapi.StartServer(_PORT, evLoop, srvLogger)
		if errServer != nil {
			fmt.Println(err)
			return
		}
	}()

	fmt.Println("Server started...")

	go inputMonitor(sc)
	<-sc

	fmt.Println("Server stopping...")
	if errStop := httpapi.StopServer(ctx, srvLogger); err != nil {
		fmt.Println(errStop)
	}

	fmt.Println("Server stopped.")
}

func initLogger() (loggerInterface.Interface, error) {
	appLogger, err := logger.NewLogger(_LOGLEVEL, "logs", "")
	if err != nil {
		return nil, err
	}

	appLogger.Infof("Server starting...")
	return appLogger, nil
}
