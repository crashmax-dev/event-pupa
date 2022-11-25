package handler

import (
	"context"
	"encoding/json"
	"eventloop/internal/httpApi/eventpreset"
	"eventloop/internal/httpApi/helper"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// eventHandler для обработки запросов по получению событий, по созданию и аттачу событий, удалению.
// Событие создаётся из пресетов, по числу после "/events/"
type eventHandler struct {
	baseHandler
}

func (eh *eventHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

	switch request.Method {
	case "GET":
		eh.get(writer, params[0])
	case "POST", "PUT":
		eh.postput(ctx, writer, params)
	case "DELETE": // Удаление ивента
		eh.delete(writer, request)
	default:
		helper.NoMethodResponse(writer, "GET, POST, PUT, PATCH")
	}
}

func (eh *eventHandler) get(writer http.ResponseWriter, eventName string) {
	eh.baseHandler.logger.Debugf(helper.ApiMessage("GET request"))
	if evnts, err := eh.baseHandler.evLoop.GetAttachedEvents(eventName); err == nil {
		if codedMessage, errJson := json.Marshal(evnts); errJson == nil {
			_, errW := writer.Write(codedMessage)
			if errW != nil {
				eh.baseHandler.logger.Errorf(helper.ApiMessage("error responding: %v"), errW)
			}
		} else {
			helper.ServerLogErr(writer, errJson.Error(), eh.baseHandler.logger, 400)
		}
	} else {
		helper.ServerLogErr(writer, err.Error(), eh.baseHandler.logger, 200)
	}
}

func (eh *eventHandler) postput(ctx context.Context, writer http.ResponseWriter, params []string) {
	id, err := strconv.Atoi(params[0])
	if err != nil || id > len(eventpreset.Events) || len(params) != 2 {
		writer.WriteHeader(404)
		return
	}

	eventName := params[1]
	newEvent, err := eventpreset.CreateEvent(id, eventpreset.REGULAR)
	if err != nil {
		eh.baseHandler.logger.Errorf(helper.ApiMessage("Error while creating event: %v"), err)
		return
	}

	errOn := eh.baseHandler.evLoop.On(ctx, eventName, newEvent, nil)
	if errOn != nil {
		eh.logger.Errorf(errOn.Error())
		writer.WriteHeader(400)
		io.WriteString(writer, "Event is not created")
		return
	}

	eh.baseHandler.logger.Infof(helper.ApiMessage("Event type %v created for %v"), id, eventName)
	_, err = io.WriteString(writer, newEvent.GetID().String())
	if err != nil {
		eh.baseHandler.logger.Errorf(helper.ApiMessage("error responding: %v"), err)
	}
}

func (eh *eventHandler) delete(writer http.ResponseWriter, request *http.Request) {
	if b, err := io.ReadAll(request.Body); err == nil {
		var sl []uuid.UUID
		errJson := json.Unmarshal(b, &sl)
		if errJson != nil {
			helper.ServerLogErr(writer, "wrong json: %v", eh.baseHandler.logger, 400, errJson)
			return
		}
		eh.baseHandler.logger.Infof(helper.ApiMessage("Removing events %v"), sl)
		for _, v := range sl {
			eh.baseHandler.evLoop.RemoveEvent(v)
		}
		_, errRespond := io.WriteString(writer, "OK")
		if errRespond != nil {
			eh.baseHandler.logger.Errorf(helper.ApiMessage("error responding: %v"), errRespond)
		}

	} else {
		helper.ServerLogErr(writer, "Error reading request: %v", eh.baseHandler.logger, 400, err.Error())
	}
}
