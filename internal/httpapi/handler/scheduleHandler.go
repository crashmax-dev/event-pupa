package handler

import (
	"context"
	"encoding/json"
	"eventloop/internal/httpapi/eventpreset"
	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/eventloop/event"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// schedulerHandler запускает и останавливает выполнение интервальных событий, создаёт новые из пресетов. При стопе
// возвращает JSON вида:
/* {Result: [...], Status: "..."}*/
type schedulerHandler struct {
	baseHandler
}

type ScheduleResponse struct {
	SchedulerStatus string
	EventStatus     string
	Result          []string
}

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

	sh.scheduleEvent(ctx, writer, &JSON, param)

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
	param string) {
	var (
		id       int
		newEvent event.Interface
		err      error
	)

	if param == "" {
		return
	}

	if id, err = strconv.Atoi(param); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(writer, "no such event: %v", sh.baseHandler.logger, 400, param)
		sh.logger.Debugf(helper.APIMessage("No such event details: %v"), err)
		return
	}

	if newEvent, err = eventpreset.CreateEvent(id, eventpreset.INTERVALED); err != nil {
		jSON.EventStatus = helper.ServerJSONLogErr(writer, "error while creating event: %v", sh.baseHandler.logger, 500, err)
		return
	}

	if err = sh.baseHandler.evLoop.ScheduleEvent(ctx, newEvent, nil); err != nil {
		writer.WriteHeader(500)
		jSON.EventStatus = "schedule event fail"
		sh.baseHandler.logger.Errorf(helper.APIMessage("schedule event fail: %v"), err)
		return
	}

	jSON.EventStatus = "Event is scheduled succesfully"
}
