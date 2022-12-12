package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"eventloop/internal/httpapi/eventpreset"
	"eventloop/internal/httpapi/helper"
	"github.com/google/uuid"
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

// get godoc
//
//	@Summary 	Get all events by trigger name
//	@Tags		events,triggers
//	@Produce	json
//	@Param		{triggerName} 	path 	string 	true 	"Name of trigger"
//	@Success	200 {array} 	uuid.UUID 	"Array of Event UUIDs"
//	@Failure	404	{string} 	string 		"No events with that name"
//	@Failure	500	{string}	string
//	@Router 	/events/{triggerName} [get]
func (eh *eventHandler) get(writer http.ResponseWriter, eventName string) {
	eh.baseHandler.logger.Debugf(helper.APIMessage("GET request"))
	if evnts, err := eh.baseHandler.evLoop.GetAttachedEvents(eventName); err == nil {
		if codedMessage, errJSON := json.Marshal(evnts); errJSON == nil {
			_, errW := writer.Write(codedMessage)
			if errW != nil {
				eh.baseHandler.logger.Errorf(helper.APIMessage("error responding: %v"), errW)
			}
		} else {
			helper.ServerLogErr(writer, errJSON.Error(), eh.baseHandler.logger, 400)
		}
	} else {
		helper.ServerLogErr(writer, err.Error(), eh.baseHandler.logger, 200)
	}
}

// postput godoc
//
// @Summary 	Create preset event with given trigger name. Return UUID of freshly created event.
// @Tags		events,triggers
// @Produce	plain
// @Param		{eventPresetId}	path	number	true	"Predefined preset for new event"
// @Param		{triggerName} 	path 	string 	true 	"Name of trigger to attach"
// @Success	200
// @Failure	404
// @Failure	500	{string}	string
// @Router 	/events/{eventPresetId}/{triggerName} [put]
// @Router 	/events/{eventPresetId}/{triggerName} [post]
func (eh *eventHandler) postput(ctx context.Context, writer http.ResponseWriter, params []string) {
	id, err := strconv.Atoi(params[0])
	if err != nil || id > len(eventpreset.Events) || len(params) != 2 {
		writer.WriteHeader(404)
		return
	}

	eventName := params[1]
	newEvent, err := eventpreset.CreateEvent(id, eventpreset.REGULAR)
	if err != nil {
		eh.baseHandler.logger.Errorf(helper.APIMessage("Error while creating event: %v"), err)
		return
	}

	errOn := eh.baseHandler.evLoop.On(ctx, eventName, newEvent, nil)
	if errOn != nil {
		eh.logger.Errorf(errOn.Error())
		writer.WriteHeader(400)
		io.WriteString(writer, "Event is not created")
		return
	}

	eh.baseHandler.logger.Infof(helper.APIMessage("Event type %v created for %v"), id, eventName)
	_, err = io.WriteString(writer, newEvent.GetID().String())
	if err != nil {
		eh.baseHandler.logger.Errorf(helper.APIMessage("error responding: %v"), err)
	}
}

// delete godoc
//
// @Summary 	Delete events by UUIDs
// @Tags		events
// @Accept	json
// @Produce	json
// @Param		{eventPresetId}	body	[]string	true	"UUIDs of events to delete"
// @Success	200	{array}		[]string	"Array of remaining not deleted events from request"
// @Failure	400 {string}	string		"Something wrong with request"
// @Router 	/events/ [delete]
func (eh *eventHandler) delete(writer http.ResponseWriter, request *http.Request) {
	if b, err := io.ReadAll(request.Body); err == nil {
		var sl []uuid.UUID
		errJSON := json.Unmarshal(b, &sl)
		if errJSON != nil {
			helper.ServerLogErr(writer, "wrong json: %v", eh.baseHandler.logger, 400, errJSON)
			return
		}
		eh.baseHandler.logger.Infof(helper.APIMessage("Removing events %v"), sl)

		eh.baseHandler.evLoop.RemoveEventByUUIDs(sl)

		_, errRespond := io.WriteString(writer, "OK")
		if errRespond != nil {
			eh.baseHandler.logger.Errorf(helper.APIMessage("error responding: %v"), errRespond)
		}
	} else {
		helper.ServerLogErr(writer, "Error reading request: %v", eh.baseHandler.logger, 400, err.Error())
	}
}
