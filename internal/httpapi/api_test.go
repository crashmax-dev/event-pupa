package httpapi

import (
	"bytes"
	"context"
	loggerImplement "eventloop/internal/logger"
	"eventloop/pkg/eventloop"
	"eventloop/pkg/logger"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const OK = "200 OK"

var testLogger logger.Interface

func TestEventCreate(t *testing.T) {
	const EVENTNAME = "test_create"

	resp, _ := createEvent(t, EVENTNAME)

	if resp.Status != OK {
		t.Errorf(resp.Status)
	}
}

func TestEventGet(t *testing.T) {
	const (
		EVENTNAME     = "test_get"
		WANT_STATUS   = "OK"
		WANT_RESPONSE = ""
	)

	createEvent(t, EVENTNAME)

	respGet, result := getEvents(t, EVENTNAME)

	if respGet.Status != WANT_STATUS && result == WANT_RESPONSE {
		t.Errorf(respGet.Status)
	}
}

func TestEventTrigger(t *testing.T) {
	const (
		EVENTNAME = "test_trigger"
		WANT      = "1"
	)

	createEvent(t, EVENTNAME)
	result := triggerEvents(t, EVENTNAME)

	if result != WANT {
		t.Errorf("RESULT: %v | WANT: %v", result, WANT)
	}
}

func TestEventToggle(t *testing.T) {
	const (
		EVENTNAME = "test_toggle"
		WANT      = "fail"
		EVENTS    = "TRIGGER,ON"
	)

	defer toggleEvents(t, bytes.NewBufferString(EVENTS))

	toggleEvents(t, bytes.NewBufferString(EVENTS))

	_, createResp := createEvent(t, EVENTNAME)

	triggerResp := triggerEvents(t, EVENTNAME)

	if !strings.Contains(createResp, WANT) && !strings.Contains(triggerResp, WANT) {
		t.Errorf("Create response: %v | Trigger response: %v | Required word: %v", createResp, triggerResp, WANT)
	}

}

func TestEventRemove(t *testing.T) {
	const (
		EVENTNAME = "test_remove"
		WANT      = "no such event name: " + EVENTNAME
	)

	_, result := createEvent(t, EVENTNAME)

	removeEvent(t, []string{result})

	_, resultGet := getEvents(t, EVENTNAME)

	if resultGet != WANT {
		t.Errorf("Result: %v, WANT: %v", resultGet, WANT)
	}
}

func TestEventSubscribe(t *testing.T) {
	const (
		EVENTNAME = "test_subscribe"
		WANT      = 2
	)

	var (
		result int
	)

	JSON := struct {
		Listeners []int `json:"listeners"`
		Triggers  []int `json:"triggers"`
	}{Listeners: []int{1, 1}, Triggers: []int{1, 1}}

	t.Log(subscribeEvents(t, EVENTNAME, JSON))

	nums := strings.Split(triggerEvents(t, EVENTNAME), ",")

	for _, v := range nums {
		temp, _ := strconv.Atoi(v)
		result += temp
	}

	if result != WANT {
		t.Errorf("Result: %v, WANT: %v", result, WANT)
	}
}

func TestEventSchedule(t *testing.T) {
	const (
		WANT = "1"
	)

	if respStart, result := scheduleEvent(t, bytes.NewBufferString("START")); respStart.StatusCode != 200 {
		t.Errorf("Server not started or event is not created: %v", result)
	}
	time.Sleep(time.Millisecond * 600)

	resp, JSON := scheduleStop(t)

	if resp.StatusCode != 200 {
		t.Errorf("Response status: %v", resp.StatusCode)
	}
	if JSON.Result[0] != WANT {
		t.Errorf("WANT: %v; Result: %v", WANT, JSON.Result[0])
	}
}

func TestMain(m *testing.M) {
	var (
		err error
	)
	testLogger, err = loggerImplement.NewLogger("debug", "logs", "test")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	evLoop := eventloop.NewEventLoop(testLogger.Level())

	go func() {
		errServ := StartServer(8090, evLoop, testLogger)
		if errServ != nil {
			fmt.Println(errServ)
			os.Exit(1)
		}
	}()

	time.Sleep(time.Millisecond * 20)

	errCode := m.Run()

	StopServer(context.TODO(), testLogger)

	os.Exit(errCode)
}
