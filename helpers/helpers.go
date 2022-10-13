package helpers

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func RemoveIndex[T any](s []T, index int) []T {
	fmt.Println(s, index)
	return append(s[:index], s[index+1:]...)
}

func GenerateIdFromNow() int {
	id, _ := strconv.Atoi(time.Now().Format("20060102150405"))
	return id
}

func GetOSFilePath(filePath string) string {
	os := runtime.GOOS
	if os == "windows" {
		return "winfile:///" + filePath
	} else {
		return filePath
	}
}
