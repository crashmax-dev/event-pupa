package api

import (
	"context"
	"encoding/json"
	"errors"
	loggerInternal "eventloop/internal/logger"
	"eventloop/pkg/api/internal"
	"eventloop/pkg/eventloop"
	"eventloop/pkg/eventloop/event"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type subscribeInfo struct {
	Listeners []int `json:"listeners"`
	Triggers  []int `json:"triggers"`
}

var (
	evLoop    eventloop.Interface
	ctx       context.Context
	cancel    context.CancelFunc
	srvLogger *zap.SugaredLogger
	atom      *zap.AtomicLevel
)

func StartServer(level zapcore.Level) {
	ctx, cancel = context.WithCancel(context.Background())
	evLoop = eventloop.NewEventLoop(level)

	http.HandleFunc("/events/", eventHandler)
	http.HandleFunc("/trigger/", triggerHandler)
	http.HandleFunc("/subscribe/", subscribeHandler)
	http.HandleFunc("/toggle/", toggleHandler)

	atom.SetLevel(loggerInternal.NormalizeLevel(level))

	servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn("Server closed")
	} else if servErr != nil {
		srvLogger.Errorf("Error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func toggleHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		srvLogger.Infof("[Toggle] No such method: %s", request.Method)
		return
	}

	b, err := io.ReadAll(request.Body)
	if err != nil {
		srvLogger.Errorf("Bad request: %v", err)
	}

	s := strings.Split(string(b), ",")
	srvLogger.Infof("Toggle: %v", s)
	for _, v := range s {
		elem := eventloop.EventFunctionMapping[v]
		if elem > 0 {
			evLoop.Toggle(elem)
		}
	}
}

func subscribeHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		return
	}

	sInfo := subscribeInfo{}
	if err := json.NewDecoder(request.Body).Decode(&sInfo); err != nil {
		fmt.Println("ooopsss! an error occurred, please try again ", err)
	}

	param := strings.TrimPrefix(request.URL.Path, "/subscribe/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	var (
		triggers, listeners []event.Interface
	)
	for _, v := range sInfo.Triggers {
		newEvent := events[v-1]()
		triggers = append(triggers, newEvent)
		evLoop.On(ctx, param, newEvent, nil)
	}
	for _, v := range sInfo.Listeners {
		listeners = append(listeners, events[v-1]())
	}

	evLoop.Subscribe(ctx, triggers, listeners)
}

func triggerHandler(writer http.ResponseWriter, request *http.Request) {
	triggerCtx, triggerCancel := context.WithTimeout(ctx, time.Second*5)
	defer triggerCancel()
	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/trigger/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	ch := eventloop.Channel[string]{Ch: make(chan string, 1)}
	go evLoop.Trigger(triggerCtx, param, &ch)

	var output []string
	for elem := range ch.Ch {
		output = append(output, elem)
	}
	_, err := io.WriteString(writer, strings.Join(output, ","))
	if err != nil {
		writer.WriteHeader(500)
		srvLogger.Errorf("error while sending trigger results: %v", err)
	}
}

func eventHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		//io.WriteString(writer, strconv.Itoa(number))
		fmt.Println("get good")
	case "POST", "PUT":
		params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

		id, err := strconv.Atoi(params[0])
		if err != nil || id > len(events) || len(params) != 2 {
			writer.WriteHeader(404)
			return
		}
		//if events[id-1].evnt == nil {
		//	fn := events[id-1].eventFunc()
		//	events[id-1].evnt = fn()
		//}
		eventName := params[1]
		evLoop.On(ctx, eventName, events[id-1](), nil)

		fmt.Println("all good")
	default:
		internal.NoMethodResponse(writer, "GET, POST, PUT")
	}
}

func StopServer() {
	defer cancel()
}

func init() {
	srvLogger, atom = loggerInternal.Initialize(zapcore.DebugLevel, "logs", "api")
}
