package handlers

import (
	"context"
	"eventloop/pkg/api/internal"
	"eventloop/pkg/eventloop"
	"io"
	"net/http"
	"strings"
	"time"
)

type triggerHandler struct {
	baseHandler
}

func (th *triggerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	triggerCtx, triggerCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer triggerCancel()
	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/trigger/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		writer.WriteHeader(404)
		return
	}

	ch := eventloop.Channel[string]{Ch: make(chan string, 1)}
	go th.baseHandler.evLoop.Trigger(triggerCtx, param, &ch)

	var output []string
	for elem := range ch.Ch {
		output = append(output, elem)
	}
	_, err := io.WriteString(writer, strings.Join(output, ","))
	if err != nil {
		internal.ServerlogErr(writer, "error while sending trigger results: %v", th.logger, 500, err)
	}
}
