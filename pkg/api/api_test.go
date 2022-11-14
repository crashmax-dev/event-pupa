package api

import (
	"go.uber.org/zap/zapcore"
	"net/http"
	"testing"
	"time"
)

func Test_Test(t *testing.T) {
	err := StartServer(zapcore.DebugLevel, nil)
	if err != nil {
		t.Log(err)
		return
	}

	time.Sleep(time.Second)
	resp, err := http.Post("https://localhost:8090/events/2", "", nil)

	t.Log(err)
	t.Log(resp)
}
