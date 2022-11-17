package handlers

import (
	"context"
	"encoding/json"
	internal2 "eventloop/internal/http_api/internal"
	"eventloop/pkg/eventloop/event"
	"net/http"
	"strings"
)

// subscribeHandler подписывает ивенты на события. Запрос должен быть JSON вида, числом обозначается пресет ивента:
/*
	{
    "listeners": [
        1,
        2
    ],
    "triggers": [
        2,
        1
    ]
}
*/
type subscribeHandler struct {
	baseHandler
}

func (sh *subscribeHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		internal2.NoMethodResponse(writer, "POST")
		return
	}

	sInfo := struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&sInfo); err != nil {
		internal2.ServerLogErr(writer, "JSON decode error: %v", sh.logger, 400, err)
		return
	}

	param := strings.TrimPrefix(request.URL.Path, "/subscribe/")
	if len(strings.SplitAfter(param, "/")) > 1 {
		internal2.ServerLogErr(writer, "Invalid params", sh.logger, 400)
		return
	}

	var (
		triggers, listeners []event.Interface
	)
	for _, v := range sInfo.Triggers {
		newEvent, err := internal2.CreateEvent(v, internal2.REGULAR)
		if err != nil {
			sh.baseHandler.logger.Errorf(internal2.ApiMessage("Error while creating trigger event: %v"), err)
		} else {
			triggers = append(triggers, newEvent)
			if errOn := sh.baseHandler.evLoop.On(ctx, param, newEvent, nil); errOn != nil {
				writer.WriteHeader(500)
				sh.baseHandler.logger.Errorf(internal2.ApiMessage("schedule event fail: %v"), errOn)
			}
		}
	}

	for _, v := range sInfo.Listeners {
		newEvent, err := internal2.CreateEvent(v, internal2.REGULAR)
		if err != nil {
			sh.baseHandler.logger.Errorf(internal2.ApiMessage("Error while creating listener event: %v"), err)
		} else {
			listeners = append(listeners, newEvent)
		}
	}

	if errSubscribe := sh.baseHandler.evLoop.Subscribe(ctx, triggers, listeners); errSubscribe != nil {
		writer.WriteHeader(500)
		sh.baseHandler.logger.Errorf(internal2.ApiMessage("event subscribe fail: %v"), errSubscribe)
	}
}
