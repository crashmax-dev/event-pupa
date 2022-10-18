package api

import (
	"context"
	"eventloop/pkg/eventloop/event"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type EventFunction func() (func(writer http.ResponseWriter, request *http.Request), *event.Interface)

type ApiEvents struct {
	evnt      *event.Interface
	eventFunc EventFunction
}

func event1() (func(writer http.ResponseWriter, request *http.Request), *event.Interface) {

	var (
		thisEvent event.Interface
		number    int
	)

	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":
			io.WriteString(writer, strconv.Itoa(number))
		case "POST":
			thisEvent = event.NewEvent(func(ctx context.Context) string {
				number++
				return strconv.Itoa(number)
			})
		default:
			writer.Header().Add("Allow", "GET, POST")
			writer.WriteHeader(405)
		}
	}, &thisEvent
}

func eventHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		//io.WriteString(writer, strconv.Itoa(number))
	case "POST":
		id := strings.TrimPrefix(request.URL.Path, "/events/")
		events[id-1].eventFunc()
	})
default:
writer.Header().Add("Allow", "GET, POST")
writer.WriteHeader(405)
}
}

func init() {
	events = []ApiEvents{{eventFunc: event1}}
}
