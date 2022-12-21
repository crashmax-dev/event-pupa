package internal

import (
	"context"
	"sync"
)

type eventLoopContextKey string

const (
	LOGGER_CTX_KEY  eventLoopContextKey = "logger"
	EXEC_CH_CTX_KEY eventLoopContextKey = "execCh"
)

var riMx sync.Mutex

func RemoveSliceItemByIndex[T any](s []T, index int) []T {
	riMx.Lock()
	defer riMx.Unlock()
	if len(s) > 1 {
		return append(s[:index], s[index+1:]...)
	}
	return nil
}

func WriteToExecCh(ctx context.Context, result string) {
	if ch := ctx.Value(EXEC_CH_CTX_KEY); ch != nil {
		ch.(chan string) <- result
	}
}
