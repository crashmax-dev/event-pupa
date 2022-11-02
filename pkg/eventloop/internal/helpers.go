package internal

import (
	"sync"
)

var riMx sync.Mutex

func RemoveIndex[T any](s []T, index int) []T {
	riMx.Lock()
	defer riMx.Unlock()
	if len(s) > 1 {
		return append(s[:index], s[index+1:]...)
	} else {
		return nil
	}
}
