package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eventloop/internal/httpapi/eventpreset"
	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/eventloop"
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

// schedulerHandler запускает и останавливает выполнение интервальных событий, создаёт новые из пресетов
type schedulerHandler struct {
	baseHandler
}

// Schedule response model info
// @Description User account information
// @Description with user id and username
type ScheduleResponse struct {
	// Id of new generated event
	UUID            uuid.UUID
	SchedulerStatus string   // Scheduler status after request
	EventStatus     string   // Status of newly generated event
	Result          []string // Results of scheduler execution after stop
}

// ServeHTTP godoc
//
// @Summary	Starts/stops event scheduler and/or creates scheduled event from preset
// @Tags		events,schedule
// @Accept	plain
// @Produce	json
// @Param		eventPresetId	path	number	false		"Predefined preset for new event" example(1)
// @Success	200	{object}	ScheduleResponse	"info about event, scheduler and scheduler execution results if scheduler stops"
// @Failure	400	{string}	string	"no event preset requested"
// @Failure	400	{string}	string	"no such event preset"
// @Failure	400	{string}	string	"scheduler is already running"
// @Failure	405	{string}	string	"only POST allowed"
// @Failure	500	{string}	string	"schedule event fail"
// @Failure	500	{string}	string	"scheduler start fail"
// @Router		/scheduler/{eventPresetId} [post]
func (sh *schedulerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var (
		JSON ScheduleResponse
	)
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		helper.NoMethodResponse(writer, "POST")
		sh.baseHandler.logger.Infof(helper.APIMessage("[Toggle] No such method: %s"), request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	JSON.UUID = sh.scheduleEvent(ctx, writer, &JSON, param)

	if errSS := sh.baseHandler.evLoop.Trigger(ctx, string(eventpreset.INTERVALED)); errSS != nil {
		writer.WriteHeader(500)
		JSON.SchedulerStatus = "Event trigger error"
		sh.baseHandler.logger.Errorf(helper.APIMessage("scheduler start fail: %v"), errSS)
	} else {
		JSON.SchedulerStatus = "Scheduler started"
	}

	byteJSON, _ := json.Marshal(JSON)
	if _, errWrite := writer.Write(byteJSON); errWrite != nil {
		sh.baseHandler.logger.Errorf(helper.APIMessage("Error responding: %v"), errWrite.Error())
	}
}

// scheduleEvent получаем ID ивента из URL, и создаём ивент
func (sh *schedulerHandler) scheduleEvent(
	ctx context.Context,
	writer http.ResponseWriter,
	jSON *ScheduleResponse,
	param string,
) uuid.UUID {
	var (
		id       int
		newEvent event.Interface
		err      error
	)

	if param == "" {
		return uuid.Nil
	}

	if id, err = strconv.Atoi(param); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(
			writer,
			"no event preset requested: %v",
			sh.baseHandler.logger,
			400,
			param,
		)
		sh.logger.Debugf(helper.APIMessage("no event preset requested details: %v"), err)
		return uuid.Nil
	}

	if newEvent, err = eventpreset.CreateEvent(id, eventpreset.INTERVALED, string(eventloop.INTERVALED)); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(
			writer,
			"error while creating event: %v",
			sh.baseHandler.logger,
			400,
			err,
		)
		return uuid.Nil
	}

	if err = sh.baseHandler.evLoop.RegisterEvent(ctx, newEvent); err != nil {
		writer.WriteHeader(500)
		jSON.EventStatus = "schedule event fail"
		sh.baseHandler.logger.Errorf(helper.APIMessage("schedule event fail: %v"), err)
		return uuid.Nil
	}

	jSON.EventStatus = "Event is scheduled succesfully"
	return uuid.MustParse(newEvent.GetUUID())
}
