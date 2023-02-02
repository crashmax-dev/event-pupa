package handler

import (
	"io"
	"net/http"
	"strings"

	"eventloop/internal/httpapi/helper"
)

// toggleHandler включает и выключает триггеры. Для срабатывания отправляется POST запрос на /toggle/
// Принимает в параметрах запроса строку с перечислением названий триггеров через
// запятую, которые надо включить или выключить.
type toggleHandler struct {
	baseHandler
}

func (tg *toggleHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		helper.NoMethodResponse(writer, "POST")
		tg.baseHandler.logger.Infof(helper.APIMessage("[Toggle] No such method: %s"), request.Method)
		return
	}

	b, err := io.ReadAll(request.Body)
	if err != nil {
		helper.ServerLogErr(writer, "Bad request: %v", tg.logger, 400, err)
		return
	}

	s := strings.Split(string(b), ",")

	outputStr := tg.baseHandler.evLoop.ToggleTriggers(s...)
	tg.baseHandler.logger.Infof(outputStr)
	_, ioerr := io.WriteString(writer, outputStr)
	if ioerr != nil {
		helper.ServerLogErr(writer, "error while responding: %v", tg.logger, 500, err)
	}
}
