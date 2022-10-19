package api

import (
	"context"
	"errors"
	"eventloop/pkg/eventloop"
	"fmt"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	http.HandleFunc("/trigger/", triggerHandler)

	servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if servErr != nil {
		fmt.Printf("error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func triggerHandler(writer http.ResponseWriter, request *http.Request) {
	triggerCtx, triggerCancel := context.WithTimeout(ctx, time.Second*5)
	defer triggerCancel()
	if request.Method != "POST" {
		NoMethodResponse(writer, "POST")
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/trigger/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	ch := make(chan string)
	evLoop.Trigger(triggerCtx, param, ch)

	var output []string
	for elem := range ch {
		output = append(output, elem)
	}
	io.WriteString(writer, strings.Join(output, ","))

}

func eventHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		//io.WriteString(writer, strconv.Itoa(number))
		fmt.Println("get good")
	case "POST", "PUT":
		params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

		id, err := strconv.Atoi(params[0])
		if err != nil || id > len(events) {
			writer.WriteHeader(404)
			return
		}
		if events[id-1].evnt == nil {
			fn := events[id-1].eventFunc()
			events[id-1].evnt = fn()
		}

		if len(params) > 1 {
			eventName := params[1]
			evLoop.On(ctx, eventName, events[id-1].evnt, nil)
		}

		fmt.Println("all good")
	default:
		NoMethodResponse(writer, "GET, POST, PUT")
	}
}

func StopServer() {
	defer cancel()
}

func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	fmt.Println("errorka")
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}
