package handlers

import (
	"eventloop/internal/http_api/internal"
	"eventloop/pkg/eventloop"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// toggleHandler включает и выключает ивенты. Принимает в параметрах запроса строку с перечислением функций через
// запятую, которые надо включить или выключить.
type toggleHandler struct {
	baseHandler
}

func (tg *toggleHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		tg.baseHandler.logger.Infof(internal.ApiMessage("[Toggle] No such method: %s"), request.Method)
		return
	}

	b, err := io.ReadAll(request.Body)
	if err != nil {
		internal.ServerLogErr(writer, "Bad request: %v", tg.logger, 400, err)
		return
	}

	s := strings.Split(string(b), ",")
	outputStr := fmt.Sprintf("Toggle: %v", s)
	tg.baseHandler.logger.Infof(outputStr)
	_, ioerr := io.WriteString(writer, outputStr)
	if ioerr != nil {
		internal.ServerLogErr(writer, "error while responding: %v", tg.logger, 500, err)
	}
	for _, v := range s {
		elem := eventloop.EventFunctionMapping[v]
		if elem > 0 {
			tg.baseHandler.evLoop.Toggle(elem)
		}
	}
}
