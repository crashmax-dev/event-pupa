package internal

import (
	"eventloop/internal/logger"
	"fmt"
	"io"
	"net/http"
)

var (
	ApiPrefix string
)

// NoMethodResponse возвращает клиенту 405 и пишет, какие методы он может использовать для запроса
func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}

// ServerLogErr пишет ошибку с форматируемым текстом format с параметрами a, в лог logger и клиенту в writer
func ServerLogErr(writer http.ResponseWriter, format string, logger logger.Interface, statusCode int, a ...any) {
	errs := fmt.Sprintf(format, a...)
	logger.Error(ApiPrefix + errs)

	if _, err := io.WriteString(writer, errs); err != nil {
		logger.Error(err)
	}
	writer.WriteHeader(statusCode)
}

func ApiMessageSetPrefix(inputApiPrefix string) {
	ApiPrefix = inputApiPrefix
}

func ApiMessage(message string) string {
	return ApiPrefix + message
}
