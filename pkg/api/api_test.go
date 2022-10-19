package api

import (
	"go.uber.org/zap/zapcore"
	"net/http"
	"testing"
	"time"
)

func Test_Test(t *testing.T) {
	StartServer(zapcore.DebugLevel)

	time.Sleep(time.Second)
	resp, err := http.Post("https://localhost:8090/events/2", "", nil)

	t.Log(err)
	t.Log(resp)
}
