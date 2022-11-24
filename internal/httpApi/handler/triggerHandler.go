package handler

import (
	"context"
	"eventloop/internal/httpApi/helper"
	"eventloop/pkg/channelEx"
	"io"
	"net/http"
	"strings"
	"time"
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
	if err := th.baseHandler.evLoop.Trigger(triggerCtx, param, ch); err != nil {
		writer.WriteHeader(500)
		th.baseHandler.logger.Errorf(internal.ApiMessage("event trigger fail: %v"), err)
	}

	var (
		output []string
		chnl   = ch.Channel()
	)
	for elem := range chnl {
		output = append(output, elem)
	}
	_, err := io.WriteString(writer, strings.Join(output, ","))
	if err != nil {
		helper.ServerLogErr(writer, "error while sending trigger results: %v", th.logger, 500, err)
	}
}
