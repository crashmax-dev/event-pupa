package handler

import (
	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/eventloop"
	"io"
	"net/http"
	"strings"
)

// toggleHandler включает и выключает ивенты. Для срабатывания отправляется POST запрос на /toggle/
// Принимает в параметрах запроса строку с перечислением функций через
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

	var ef []eventloop.EventFunction

	for _, v := range s {
		elem := eventloop.EventFunctionMapping[v]
		if elem > 0 {
			ef = append(ef, elem)
		}
	}
	outputStr := tg.baseHandler.evLoop.Toggle(ef...)
	tg.baseHandler.logger.Infof(outputStr)
	_, ioerr := io.WriteString(writer, outputStr)
	if ioerr != nil {
		helper.ServerLogErr(writer, "error while responding: %v", tg.logger, 500, err)
	}
}
