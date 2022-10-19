package internal

import (
	"fmt"
	"net/http"
)

func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	fmt.Println("errorka")
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}
