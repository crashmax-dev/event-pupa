package handlers

import (
	"context"
	internal2 "eventloop/internal/http_api/internal"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// schedulerHandler запускает и останавливает выполнение интервальных событий, создаёт новые из пресетов/
type schedulerHandler struct {
	baseHandler
}

func (sh *schedulerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		internal2.NoMethodResponse(writer, "POST")
		sh.baseHandler.logger.Infof(internal2.ApiMessage("[Toggle] No such method: %s"), request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	//Получаем ID ивента из URL, и создаём
	if id, err := strconv.Atoi(param); err != nil {
		internal2.ServerLogErr(writer, "no such event: %v", sh.baseHandler.logger, 400, param)
		sh.logger.Debugf(internal2.ApiMessage("No such event details: %v"), err)
	} else {
		if newEvent, errC := internal2.CreateEvent(id, internal2.INTERVALED); errC != nil {
			internal2.ServerLogErr(writer, "error while creating event: %v", sh.baseHandler.logger, 500, errC)
		} else {
			if errSE := sh.baseHandler.evLoop.ScheduleEvent(ctx, newEvent, nil); errSE != nil {
				writer.WriteHeader(500)
				sh.baseHandler.logger.Errorf(internal2.ApiMessage("schedule event fail: %v"), errSE)
			}
		}
	}

	if b, err := io.ReadAll(request.Body); err != nil {
		internal2.ServerLogErr(writer, "bad request: %v", sh.baseHandler.logger, 400, err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			if errSS := sh.baseHandler.evLoop.StartScheduler(ctx); errSS != nil {
				writer.WriteHeader(500)
				sh.baseHandler.logger.Errorf(internal2.ApiMessage("scheduler start fail: %v"), errSS)
			}
		case "stop":
			sh.baseHandler.evLoop.StopScheduler()
		default:
			sh.baseHandler.logger.Errorf(internal2.ApiMessage("No known method: %v"), sm)
		}
	}
}
