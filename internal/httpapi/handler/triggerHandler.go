package handler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/channelEx"
)

// triggerHandler триггерит ивенты по имени
type triggerHandler struct {
	baseHandler
}

func (th *triggerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	triggerCtx, triggerCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer triggerCancel()
	if request.Method != "POST" {
		helper.NoMethodResponse(writer, "POST")
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/trigger/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	ch := channelEx.NewChannel(1)

	var errTrig error
	go func() {
		if errTrig = th.baseHandler.evLoop.Trigger(triggerCtx, param, ch); errTrig != nil {
			io.WriteString(writer, "Event trigger fail")
			th.baseHandler.logger.Errorf(helper.APIMessage("event trigger fail: %v"), errTrig)
			return
		}
	}()

	var (
		output []string
	)
	for elem := range ch.Channel() {
		output = append(output, elem)
	}
	if len(output) == 0 {
		helper.ServerLogErr(writer, "nothing to trigger", th.logger, 204)
		return
	}

	_, err := io.WriteString(writer, strings.Join(output, ","))

	if err != nil {
		helper.ServerLogErr(writer, "error while sending trigger results: %v", th.logger, 500, err)
	}
}
