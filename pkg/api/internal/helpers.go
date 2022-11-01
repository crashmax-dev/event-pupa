package internal

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	fmt.Println("server respond 405")
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}

func ServerlogErr(writer http.ResponseWriter, format string, logger *zap.SugaredLogger, a ...any) {
	errs := fmt.Sprintf(format, a...)
	logger.Error(errs)
	_, err := io.WriteString(writer, errs)
	if err != nil {
		logger.Errorf("error while responsing: %v", err)
	}
}
