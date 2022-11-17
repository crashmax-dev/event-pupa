package handlers

import (
	"context"
	"encoding/json"
	internal2 "eventloop/internal/http_api/internal"
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
		eh.baseHandler.logger.Debugf(internal2.ApiMessage("GET request"))
		if evnts, err := eh.baseHandler.evLoop.GetAttachedEvents(params[0]); err == nil {
			if codedMessage, errJson := json.Marshal(evnts); errJson == nil {
				_, errW := writer.Write(codedMessage)
				if errW != nil {
					eh.baseHandler.logger.Errorf(internal2.ApiMessage("error responding: %v"), errW)
				}
			} else {
				internal2.ServerLogErr(writer, errJson.Error(), eh.baseHandler.logger, 400)
			}
		} else {
			internal2.ServerLogErr(writer, err.Error(), eh.baseHandler.logger, 200)
		}
	case "POST", "PUT":
		id, err := strconv.Atoi(params[0])
		if err != nil || id > len(internal2.Events) || len(params) != 2 {
			writer.WriteHeader(404)
			return
		}

		eventName := params[1]
		newEvent, err := internal2.CreateEvent(id, internal2.REGULAR)
		if err != nil {
			eh.baseHandler.logger.Errorf(internal2.ApiMessage("Error while creating event: %v"), err)
			return
		}

		go func() {
			errOn := eh.baseHandler.evLoop.On(ctx, eventName, newEvent, nil)
			if errOn != nil {
				eh.logger.Errorf(errOn.Error())
			}
		}()

		eh.baseHandler.logger.Infof(internal2.ApiMessage("Event type %v created for %v"), id, eventName)
		_, err = io.WriteString(writer, "OK")
		if err != nil {
			eh.baseHandler.logger.Errorf(internal2.ApiMessage("error responding: %v"), err)
		}
	case "PATCH":
		if b, err := io.ReadAll(request.Body); err == nil {
			var sl []uuid.UUID
			errJson := json.Unmarshal(b, &sl)
			if errJson != nil {
				internal2.ServerLogErr(writer, "wrong json: %v", eh.baseHandler.logger, 400, errJson)
				return
			}
			eh.baseHandler.logger.Infof(internal2.ApiMessage("Removing events %v"), sl)
			for _, v := range sl {
				eh.baseHandler.evLoop.RemoveEvent(v)
			}
			_, errRespond := io.WriteString(writer, "OK")
			if errRespond != nil {
				eh.baseHandler.logger.Errorf(internal2.ApiMessage("error responding: %v"), errRespond)
			}

		} else {
			internal2.ServerLogErr(writer, "Error reading request: %v", eh.baseHandler.logger, 400, err.Error())
		}
	default:
		internal2.NoMethodResponse(writer, "GET, POST, PUT, PATCH")
	}
}
