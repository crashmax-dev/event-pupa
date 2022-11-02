package handlers

import (
	"context"
	"encoding/json"
	"eventloop/pkg/api/internal"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type eventHandler struct {
	baseHandler
}

func (eh *eventHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	params := strings.Split(strings.TrimPrefix(request.URL.Path, "/events/"), "/")

	switch request.Method {
	case "GET":
		eh.baseHandler.logger.Infof("GET request")
		if evnts, err := eh.baseHandler.evLoop.GetEventsByName(params[0]); err == nil {
			if codedMessage, errJson := json.Marshal(evnts); err == nil {
				_, errW := writer.Write(codedMessage)
				if errW != nil {
					eh.baseHandler.logger.Errorf("error responding: %v", errW)
				}
			} else {
				internal.ServerlogErr(writer, errJson.Error(), eh.baseHandler.logger, 400)
			}
		} else {
			internal.ServerlogErr(writer, err.Error(), eh.baseHandler.logger, 200)
		}
	case "POST", "PUT":
		id, err := strconv.Atoi(params[0])
		if err != nil || id > len(internal.Events) || len(params) != 2 {
			writer.WriteHeader(404)
			return
		}

		eventName := params[1]
		newEvent, err := internal.CreateEvent(id, internal.REGULAR)
		if err != nil {
			eh.baseHandler.logger.Errorf("Error while creating event: %v", err)
			return
		}
		ch := make(chan uuid.UUID)
		go eh.baseHandler.evLoop.On(ctx, eventName, newEvent, ch)

		eh.baseHandler.logger.Infof("Event type %v created for %v", id, eventName)
		_, err = io.WriteString(writer, "OK")
		if err != nil {
			eh.baseHandler.logger.Errorf("error responding: %v", err)
		}
		<-ch
	case "PATCH":
		if b, err := io.ReadAll(request.Body); err == nil {
			var sl []uuid.UUID
			errJson := json.Unmarshal(b, &sl)
			if errJson != nil {
				internal.ServerlogErr(writer, "wrong json: %v", eh.baseHandler.logger, 400, errJson)
				return
			}
			eh.baseHandler.logger.Infof("Removing events %v", sl)
			for _, v := range sl {
				eh.baseHandler.evLoop.RemoveEvent(v)
			}
			_, errRespond := io.WriteString(writer, "OK")
			if errRespond != nil {
				eh.baseHandler.logger.Errorf("error responding: %v", errRespond)
			}

		} else {
			internal.ServerlogErr(writer, "Error reading request: %v", eh.baseHandler.logger, 400, err.Error())
		}
	default:
		internal.NoMethodResponse(writer, "GET, POST, PUT, PATCH")
	}
}
