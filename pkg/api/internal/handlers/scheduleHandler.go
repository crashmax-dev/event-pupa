package handlers

import (
	"context"
	"eventloop/pkg/api/internal"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type schedulerHandler struct {
	baseHandler
}

func (sh *schedulerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		sh.baseHandler.logger.Infof("[Toggle] No such method: %s", request.Method)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/scheduler/")

	//Получаем ID ивента из URL, и создаём
	if id, err := strconv.Atoi(param); err != nil {
		internal.ServerlogErr(writer, "no such event: %v", sh.baseHandler.logger, 400, param)
		sh.logger.Debugf("No such event details: %v", err)
	} else {
		if newEvent, errC := internal.CreateEvent(id, internal.INTERVALED); errC != nil {
			internal.ServerlogErr(writer, "error while creating event: %v", sh.baseHandler.logger, 500, errC)
		} else {
			sh.baseHandler.evLoop.ScheduleEvent(ctx, newEvent, nil)
		}
	}

	if b, err := io.ReadAll(request.Body); err != nil {
		internal.ServerlogErr(writer, "bad request: %v", sh.baseHandler.logger, 400, err)
	} else {
		switch sm := strings.ToLower(string(b)); sm {
		case "start":
			sh.baseHandler.evLoop.StartScheduler(ctx)
		case "stop":
			sh.baseHandler.evLoop.StopScheduler()
		default:
			sh.baseHandler.logger.Errorf("No known method: %v", sm)
		}
	}
}
