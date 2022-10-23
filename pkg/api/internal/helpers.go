package internal

import (
	"fmt"
	"net/http"
)

func NoMethodResponse(writer http.ResponseWriter, allowed string) {
	fmt.Println("server respond 405")
	writer.Header().Add("Allow", allowed)
	writer.WriteHeader(405)
}
