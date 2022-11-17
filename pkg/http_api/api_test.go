package http_api

import (
	"net/http"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	//err := StartServer()
	//if err != nil {
	//	t.Log(err)
	//	return
	//}

	time.Sleep(time.Second)
	resp, err := http.Get("localhost:8090/events/event1")

	t.Log(err)
	t.Log(resp)
}
