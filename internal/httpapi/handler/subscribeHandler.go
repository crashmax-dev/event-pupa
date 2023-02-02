package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"eventloop/internal/httpapi/eventpreset"
	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/eventloop/event"
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

// ServeHTTP принимает JSON с массивами listeners и triggers, в каждом из которых пресеты ивентов, которые будут созданы
func (sh *subscribeHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx, _ := context.WithCancel(context.Background())

	if request.Method != "POST" {
		helper.NoMethodResponse(writer, "POST")
		return
	}

	sInfo := struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&sInfo); err != nil {
		helper.ServerLogErr(writer, "JSON decode error: %v", sh.logger, 400, err)
		return
	}

	triggerName := strings.TrimPrefix(request.URL.Path, "/subscribe/")
	if len(strings.SplitAfter(triggerName, "/")) > 1 {
		helper.ServerLogErr(writer, "Invalid params", sh.logger, 400)
		return
	}

	var (
		triggers, listeners []event.Interface
	)
	for _, v := range sInfo.Triggers {
		newEvent, err := eventpreset.CreateEvent(v, eventpreset.REGULAR, triggerName)
		if err != nil {
			sh.baseHandler.logger.Errorf(helper.APIMessage("Error while creating trigger event: %v"), err)
		} else {
			triggers = append(triggers, newEvent)
			if errRegister := sh.baseHandler.evLoop.RegisterEvent(ctx, newEvent); errRegister != nil {
				sh.baseHandler.logger.Errorf(helper.APIMessage("schedule event fail: %v"), errRegister)
			}
		}
	}

	for _, v := range sInfo.Listeners {
		newEvent, err := eventpreset.CreateEvent(v, eventpreset.REGULAR, "")
		if err != nil {
			sh.baseHandler.logger.Errorf(helper.APIMessage("Error while creating listener event: %v"), err)
		} else {
			listeners = append(listeners, newEvent)
		}
	}

	if errSubscribe := sh.baseHandler.evLoop.Subscribe(ctx, triggers, listeners); errSubscribe != nil {
		writer.WriteHeader(500)
		sh.baseHandler.logger.Errorf(helper.APIMessage("event subscribe fail: %v"), errSubscribe)
	}
}
