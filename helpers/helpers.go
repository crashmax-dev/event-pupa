package helpers

import "fmt"

func RemoveIndex[T any](s []T, index int) []T {
	fmt.Println(s, index)
	return append(s[:index], s[index+1:]...)
}
