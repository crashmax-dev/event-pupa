package internal

import (
	"context"
	"encoding/json"
	"eventloop/pkg/eventloop"
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BaseHandler struct {
	logger *zap.SugaredLogger
	evLoop eventloop.Interface
}

type EventHandler struct {
	Base BaseHandler
}

type TriggerHandler struct {
	Base BaseHandler
}

type SubscribeHandler struct {
	Base      BaseHandler
	Listeners []int `json:"listeners"`
	Triggers  []int `json:"triggers"`
}

type ToggleHandler struct {
	Base BaseHandler
}

type SchedulerHandler struct {
	Base BaseHandler
}

func NewBaseHandler(logger *zap.SugaredLogger, evLoop eventloop.Interface) BaseHandler {
	return BaseHandler{logger: logger, evLoop: evLoop}
}

func (eh *EventHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())

	params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

	switch request.Method {
	case "GET":
		eh.Base.logger.Infof("GET request")
		if evnts, err := eh.Base.evLoop.GetEventsByName(params[0]); err == nil {
			if codedMessage, errJson := json.Marshal(evnts); err == nil {
				_, errW := writer.Write(codedMessage)
				if errW != nil {
					eh.Base.logger.Errorf("error responding: %v", errW)
				}
			} else {
				ServerlogErr(writer, errJson.Error(), nil)
			}
		} else {
			ServerlogErr(writer, err.Error(), eh.Base.logger)
		}
	case "POST", "PUT":
		id, err := strconv.Atoi(params[0])
		if err != nil || id > len(Events) || len(params) != 2 {
			writer.WriteHeader(404)
			return
		}

		eventName := params[1]
		newEvent, err := CreateEvent(id, REGULAR)
		if err != nil {
			eh.Base.logger.Errorf("Error while creating event: %v", err)
			return
		}
		ch := make(chan uuid.UUID)
		eh.Base.evLoop.On(ctx, eventName, newEvent, ch)

		eh.Base.logger.Infof("Event type %v created for %v", id, eventName)
		_, err = io.WriteString(writer, "OK")
		if err != nil {
			eh.Base.logger.Errorf("error responding: %v", err)
		}
		<-ch
		cancel()
	case "PATCH":
		if b, err := io.ReadAll(request.Body); err == nil {
			var sl []uuid.UUID
			errJson := json.Unmarshal(b, &sl)
			if errJson != nil {
				ServerlogErr(writer, "wrong json: %v", eh.Base.logger, errJson)
				return
			}
			eh.Base.logger.Infof("Removing events %v", sl)
			for _, v := range sl {
				eh.Base.evLoop.RemoveEvent(v)
			}
			_, errRespond := io.WriteString(writer, "OK")
			if errRespond != nil {
				eh.Base.logger.Errorf("error responding: %v", errRespond)
			}

		} else {
			ServerlogErr(writer, "Error reading request: %v", eh.Base.logger, err.Error())
		}
	default:
		NoMethodResponse(writer, "GET, POST, PUT, PATCH")
	}
}

func (th *TriggerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	triggerCtx, triggerCancel := context.WithTimeout(context.Background(), time.Second*5)
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

	ch := eventloop.Channel[string]{Ch: make(chan string, 1)}
	go th.Base.evLoop.Trigger(triggerCtx, param, &ch)

	var output []string
	for elem := range ch.Ch {
		output = append(output, elem)
	}
	_, err := io.WriteString(writer, strings.Join(output, ","))
	if err != nil {
		writer.WriteHeader(500)
		th.Base.logger.Errorf("error while sending trigger results: %v", err)
	}
}

func (sh *SubscribeHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		NoMethodResponse(writer, "POST")
		return
	}

	sInfo := struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&sInfo); err != nil {
		sh.Base.logger.Errorf("JSON decode error: %v", err)
	}

	param := strings.TrimPrefix(request.URL.Path, "/subscribe/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	var (
		triggers, listeners []event.Interface
	)
	for _, v := range sh.Triggers {
		newEvent, err := CreateEvent(v, REGULAR)
		if err != nil {
			sh.Base.logger.Errorf("Error while creating trigger event: %v", err)
		} else {
			triggers = append(triggers, newEvent)
			sh.Base.evLoop.On(ctx, param, newEvent, nil)
		}

	}
	for _, v := range sh.Listeners {
		newEvent, err := CreateEvent(v, REGULAR)
		if err != nil {
			sh.Base.logger.Errorf("Error while creating listener event: %v", err)
		} else {
			listeners = append(listeners, newEvent)
		}
	}

	sh.Base.evLoop.Subscribe(ctx, triggers, listeners)
}

func (tg *ToggleHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		NoMethodResponse(writer, "POST")
		tg.Base.logger.Infof("[Toggle] No such method: %s", request.Method)
		return
	}

	b, err := io.ReadAll(request.Body)
	if err != nil {
		tg.Base.logger.Errorf("Bad request: %v", err)
	}

	s := strings.Split(string(b), ",")
	tg.Base.logger.Infof("Toggle: %v", s)
	for _, v := range s {
		elem := eventloop.EventFunctionMapping[v]
		if elem > 0 {
			tg.Base.evLoop.Toggle(elem)
		}
	}
}

func (sh *SchedulerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		NoMethodResponse(writer, "POST")
		sh.Base.logger.Infof("[Toggle] No such method: %s", request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	//Получаем ID ивента из URL, и создаём
	if id, err := strconv.Atoi(param); err != nil {
		ServerlogErr(writer, "no such event: %v", sh.Base.logger, err)
	} else {
		if newEvent, errС := CreateEvent(id, INTERVALED); errС != nil {
			ServerlogErr(writer, "error while creating event: %v", sh.Base.logger, errС)
		} else {
			sh.Base.evLoop.ScheduleEvent(ctx, newEvent, nil)
		}
	}

	if b, err := io.ReadAll(request.Body); err != nil {
		ServerlogErr(writer, "bad request: %v", sh.Base.logger, err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			sh.Base.evLoop.StartScheduler(ctx)
		case "stop":
			sh.Base.evLoop.StopScheduler()
		default:
			sh.Base.logger.Errorf("No known method: %v", sm)
		}
	}
}
