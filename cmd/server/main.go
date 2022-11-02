package main

import (
	"context"
	"eventloop/pkg/api"
	"fmt"
	"go.uber.org/zap/zapcore"
	"strings"
)

func main() {
	var quit chan struct{}
	go api.StartServer(zapcore.InfoLevel, quit)
	fmt.Println("Server started...")
	var input string
	for strings.ToLower(input) != "stop" {
		fmt.Scanln(&input)
	}
	api.StopServer(context.Background())
	//<-quit
	fmt.Println("Server stopped.")
}
