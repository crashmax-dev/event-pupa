package helpers

import (
	"fmt"
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
