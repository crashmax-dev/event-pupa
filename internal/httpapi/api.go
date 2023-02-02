package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eventloop/cmd/server/docs"
	"eventloop/internal/httpapi/handler"
	"eventloop/internal/httpapi/helper"
	"eventloop/pkg/eventloop"
	"eventloop/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "eventloop/cmd/server/docs"
)

const _APIPREFIX = "[API] "

var (
	serv http.Server
)

// StartServer стартует API сервер для доступа к Event Loop. Функция блокирующая
func StartServer(port int, evLoop eventloop.Interface, srvLogger logger.Interface) error {
	helper.APIMessageSetPrefix(_APIPREFIX)

	handlersMap := map[string]handler.Type{"/events/": handler.EVENT,
		"/trigger/":   handler.TRIGGER,
		"/subscribe/": handler.SUBSCRIBE,
		"/toggle/":    handler.TOGGLE,
		"/scheduler/": handler.SCHEDULER}

	mux := http.NewServeMux()
	for k, v := range handlersMap {
		mux.Handle(k, handler.NewHandler(v, srvLogger, evLoop))
	}

	// Swagger
	docs.SwaggerInfo.Host = fmt.Sprintf(docs.SwaggerInfo.Host, port)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL(
		fmt.Sprintf("http://localhost:%v/swagger/doc.json", port))))

	serv = http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	servErr := serv.ListenAndServe()
	if errors.Is(servErr, http.ErrServerClosed) {
		srvLogger.Warn(helper.APIMessage("Server closed"))
	} else if servErr != nil {
		return servErr
	}

	return nil
}

func StopServer(ctx context.Context, srvLogger logger.Interface) error {
	err := serv.Shutdown(ctx)
	if err != nil {
		return err
	}
	srvLogger.Infof(helper.APIMessage("Server stopped."))
	return nil
}
