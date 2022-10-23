package helpers

import (
	"runtime"
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

//func GenerateIdFromNow() string {
//	t := time.Now()
//	//t = t.Add(time.Millisecond)
//	id := strings.Replace(t.Format("20060102150405.00000000000000000"), ".", "", -1)
//	return id
//}

func GetOSFilePath(filePath string) string {
	os := runtime.GOOS
	if os == "windows" {
		return "winfile:///" + filePath
	} else {
		return filePath
	}
}
