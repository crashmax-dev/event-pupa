package handlers

import (
	"context"
	"encoding/json"
	"eventloop/pkg/api/internal"
	"eventloop/pkg/eventloop/event"
	"net/http"
	"strings"
)

type subscribeHandler struct {
	baseHandler
}

func (sh *subscribeHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		internal.NoMethodResponse(writer, "POST")
		return
	}

	sInfo := struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&sInfo); err != nil {
		internal.ServerlogErr(writer, "JSON decode error: %v", sh.logger, 400, err)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/subscribe/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		internal.ServerlogErr(writer, "Invalid params", sh.logger, 400)
		return
	}

	var (
		triggers, listeners []event.Interface
	)
	for _, v := range sInfo.Triggers {
		newEvent, err := internal.CreateEvent(v, internal.REGULAR)
		if err != nil {
			sh.baseHandler.logger.Errorf("Error while creating trigger event: %v", err)
		} else {
			triggers = append(triggers, newEvent)
			sh.baseHandler.evLoop.On(ctx, param, newEvent, nil)
		}
	}

	for _, v := range sInfo.Listeners {
		newEvent, err := internal.CreateEvent(v, internal.REGULAR)
		if err != nil {
			sh.baseHandler.logger.Errorf("Error while creating listener event: %v", err)
		} else {
			listeners = append(listeners, newEvent)
		}
	}

	sh.baseHandler.evLoop.Subscribe(ctx, triggers, listeners)
}
