package internal

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}

func ServerlogErr(writer http.ResponseWriter, format string, logger *zap.SugaredLogger, statusCode int, a ...any) {
	errs := fmt.Sprintf(format, a...)
	logger.Error(errs)

	_, err := io.WriteString(writer, errs)
	writer.WriteHeader(statusCode)

	if err != nil {
		logger.Errorf("error while responsing: %v", err)
	}
}
