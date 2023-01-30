package logger

import (
	"context"
)

type ctxValue struct{}

var loggerKey = ctxValue{}

func FromContext(ctx context.Context) Interface {
	return ctx.Value(loggerKey).(Interface)
}

func WithLogger(ctx context.Context, l Interface) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}
