package handlers

import (
	"context"
	"encoding/json"
	"eventloop/internal/httpApi/eventpreset"
	"eventloop/internal/httpApi/helper"
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
	Status string
	Result []string
}

func (sh *schedulerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var (
		JSON ScheduleResponse
	)
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		helper.NoMethodResponse(writer, "POST")
		sh.baseHandler.logger.Infof(helper.ApiMessage("[Toggle] No such method: %s"), request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	//Получаем ID ивента из URL, и создаём
	if param != "" {
		if id, err := strconv.Atoi(param); err == nil {
			if newEvent, errC := eventpreset.CreateEvent(id, eventpreset.INTERVALED); errC == nil {
				if errSE := sh.baseHandler.evLoop.ScheduleEvent(ctx, newEvent, nil); errSE == nil {
					JSON.Status = "Event is scheduled succesfully"
				} else {
					writer.WriteHeader(500)
					sh.baseHandler.logger.Errorf(helper.ApiMessage("schedule event fail: %v"), errSE)
				}
			} else {
				helper.ServerLogErr(writer, "error while creating event: %v", sh.baseHandler.logger, 500, errC)
			}
		} else {
			helper.ServerLogErr(writer, "no such event: %v", sh.baseHandler.logger, 400, param)
			sh.logger.Debugf(helper.ApiMessage("No such event details: %v"), err)
		}
	}

	if b, err := io.ReadAll(request.Body); err != nil {
		helper.ServerLogErr(writer, "bad request: %v", sh.baseHandler.logger, 400, err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			if errSS := sh.baseHandler.evLoop.StartScheduler(ctx); errSS != nil {
				if sh.baseHandler.evLoop.IsSchedulerRunning() {
					writer.WriteHeader(400)
					io.WriteString(writer, "Scheduler is already running")
				} else {
					writer.WriteHeader(500)
				}
				sh.baseHandler.logger.Errorf(helper.ApiMessage("scheduler start fail: %v"), errSS)
			} else {
				io.WriteString(writer, ". Scheduler started")
			}
		case "stop":
			sh.baseHandler.evLoop.StopScheduler()
			JSON = ScheduleResponse{Status: "Scheduler stopped", Result: sh.baseHandler.evLoop.GetSchedulerResults()}

		default:
			sh.baseHandler.logger.Errorf(helper.ApiMessage("No known method: %v"), sm)
		}
	}
	byteJson, _ := json.Marshal(JSON)
	writer.Write(byteJson)
}
