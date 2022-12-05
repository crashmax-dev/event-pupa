package helper

import (
	"eventloop/pkg/logger"
	"fmt"
	"io"
	"net/http"
)

var (
	APIPrefix string
)

// NoMethodResponse возвращает клиенту 405 и пишет, какие методы он может использовать для запроса
func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}

// ServerLogErr пишет ошибку с форматируемым текстом format с параметрами a, в лог logger и клиенту в writer
func ServerLogErr(writer http.ResponseWriter, format string, logger logger.Interface, statusCode int, a ...any) {
	errs := fmt.Sprintf(format, a...)
	logger.Error(APIPrefix + errs)

	writer.WriteHeader(statusCode)
	if _, err := io.WriteString(writer, errs); err != nil {
		logger.Error(err)
	}
}

func ServerJSONLogErr(writer http.ResponseWriter,
	format string,
	logger logger.Interface,
	statusCode int,
	a ...any) string {
	errs := fmt.Sprintf(format, a...)
	logger.Error(APIPrefix + errs)

	writer.WriteHeader(statusCode)
	return errs
}

func APIMessageSetPrefix(inputApiPrefix string) {
	APIPrefix = inputApiPrefix
}

func APIMessage(message string) string {
	return APIPrefix + message
}
