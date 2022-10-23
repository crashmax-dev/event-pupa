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
	"github.com/google/uuid"
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
	http.HandleFunc("/scheduler/", schedulerHandler)

	atom.SetLevel(loggerInternal.NormalizeLevel(level))

	servErr := http.ListenAndServe(":8090", nil)
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn("Server closed")
	} else if servErr != nil {
		srvLogger.Errorf("Error starting server: %s\n", servErr)
		os.Exit(1)
	}
}

func serverlogErr(writer http.ResponseWriter, format string, a ...any) {
	errs := fmt.Sprintf(format, a...)
	srvLogger.Error(errs)
	_, err := io.WriteString(writer, errs)
	if err != nil {
		srvLogger.Errorf("error while responsing: %v", err)
	}
}

func schedulerHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		srvLogger.Infof("[Toggle] No such method: %s", request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	//Получаем ID ивента из URL, и создаём
	if id, err := strconv.Atoi(param); err != nil {
		serverlogErr(writer, "no such event: %v", err)
	} else {
		if newEvent, errС := CreateEvent(id, INTERVALED); errС != nil {
			serverlogErr(writer, "error while creating event: %v", errС)
		} else {
			evLoop.ScheduleEvent(ctx, newEvent, nil)
		}
	}

	if b, err := io.ReadAll(request.Body); err != nil {
		serverlogErr(writer, "bad request: %v", err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			evLoop.StartScheduler(ctx)
		case "stop":
			evLoop.StopScheduler()
		default:
			srvLogger.Errorf("No known method: %v", sm)
		}
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
		srvLogger.Errorf("JSON decode error: %v", err)
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
		newEvent, err := CreateEvent(v, REGULAR)
		if err != nil {
			srvLogger.Errorf("Error while creating trigger event: %v", err)
		} else {
			triggers = append(triggers, newEvent)
			evLoop.On(ctx, param, newEvent, nil)
		}

	}
	for _, v := range sInfo.Listeners {
		newEvent, err := CreateEvent(v, REGULAR)
		if err != nil {
			srvLogger.Errorf("Error while creating listener event: %v", err)
		} else {
			listeners = append(listeners, newEvent)
		}
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
	params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

	switch request.Method {
	case "GET":
		srvLogger.Infof("GET request")
		if evnts, err := evLoop.GetEventsByName(params[0]); err == nil {
			if codedMessage, errJson := json.Marshal(evnts); err == nil {
				_, errW := writer.Write(codedMessage)
				if errW != nil {
					srvLogger.Errorf("error responding: %v", errW)
				}
			} else {
				serverlogErr(writer, errJson.Error(), nil)
			}
		} else {
			serverlogErr(writer, err.Error(), nil)
		}
	case "POST", "PUT":
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
		newEvent, err := CreateEvent(id, REGULAR)
		if err != nil {
			srvLogger.Errorf("Error while creating event: %v", err)
			return
		}
		evLoop.On(ctx, eventName, newEvent, nil)

		srvLogger.Infof("Event type %v created for %v", id, eventName)
		_, err = io.WriteString(writer, "OK")
		if err != nil {
			srvLogger.Errorf("error responding: %v", err)
		}

	case "PATCH":
		if b, err := io.ReadAll(request.Body); err == nil {
			var sl []uuid.UUID
			errJson := json.Unmarshal(b, &sl)
			if errJson != nil {
				serverlogErr(writer, "wrong json: %v", errJson)
				return
			}
			srvLogger.Infof("Removing events %v", sl)
			for _, v := range sl {
				evLoop.RemoveEvent(v)
			}
			_, errRespond := io.WriteString(writer, "OK")
			if errRespond != nil {
				srvLogger.Errorf("error responding: %v", errRespond)
			}

		} else {
			serverlogErr(writer, "Error reading request: %v", err.Error())
		}
	default:
		internal.NoMethodResponse(writer, "GET, POST, PUT, PATCH")
	}
}

func StopServer() {
	defer cancel()
}

func init() {
	srvLogger, atom = loggerInternal.Initialize(zapcore.DebugLevel, "logs", "api")
}
