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
	uuid.UUID
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
// @Param		schedulerMethod	body	string	false	"Method for scheduler to start or stop" example(START)
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

	if b, err := io.ReadAll(request.Body); err != nil {
		JSON.SchedulerStatus = helper.ServerJSONLogErr(writer,
			"bad request: %v",
			sh.baseHandler.logger,
			400,
			err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			if errSS := sh.baseHandler.evLoop.StartScheduler(ctx); errSS != nil {
				if sh.baseHandler.evLoop.Scheduler().IsSchedulerRunning() {
					writer.WriteHeader(400)
					JSON.SchedulerStatus = "Scheduler is already running"
				} else {
					writer.WriteHeader(500)
				}
				sh.baseHandler.logger.Errorf(helper.APIMessage("scheduler start fail: %v"), errSS)
			} else {
				JSON.SchedulerStatus = "Scheduler started"
			}
		case "stop":
			sh.baseHandler.evLoop.StopScheduler()
			JSON = ScheduleResponse{EventStatus: JSON.EventStatus,
				SchedulerStatus: "Scheduler stopped",
				Result:          sh.baseHandler.evLoop.Scheduler().GetSchedulerResults()}

		default:
			sh.baseHandler.logger.Errorf(helper.APIMessage("No known method: %v"), sm)
			JSON.SchedulerStatus = "no know method for scheduler status change"
		}
	}
	byteJSON, _ := json.Marshal(JSON)
	if _, errWrite := writer.Write(byteJSON); errWrite != nil {
		sh.baseHandler.logger.Errorf(helper.APIMessage("Error responding: %v"), errWrite.Error())
	}
}

// scheduleEvent получаем ID ивента из URL, и создаём ивент
func (sh *schedulerHandler) scheduleEvent(ctx context.Context,
	writer http.ResponseWriter,
	jSON *ScheduleResponse,
	param string) uuid.UUID {
	var (
		id       int
		newEvent event.Interface
		err      error
	)

	if param == "" {
		return uuid.Nil
	}

	if id, err = strconv.Atoi(param); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(writer, "no event preset requested: %v", sh.baseHandler.logger, 400, param)
		sh.logger.Debugf(helper.APIMessage("no event preset requested details: %v"), err)
		return uuid.Nil
	}

	if newEvent, err = eventpreset.CreateEvent(id, eventpreset.INTERVALED); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(writer, "error while creating event: %v", sh.baseHandler.logger, 400, err)
		return uuid.Nil
	}

	if err = sh.baseHandler.evLoop.ScheduleEvent(ctx, newEvent, nil); err != nil {
		writer.WriteHeader(500)
		jSON.EventStatus = "schedule event fail"
		sh.baseHandler.logger.Errorf(helper.APIMessage("schedule event fail: %v"), err)
		return uuid.Nil
	}

	jSON.EventStatus = "Event is scheduled succesfully"
	return newEvent.GetID()
}
