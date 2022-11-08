package main

import (
	"context"
	"eventloop/pkg/api"
	"fmt"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func inputMonitor(sc chan<- os.Signal) {
	var input string
	for strings.ToLower(input) != "stop" {
		fmt.Scanln(&input)
	}
	sc <- syscall.SIGTERM
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan struct{})
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	go api.StartServer(zapcore.InfoLevel, &quit)
	fmt.Println("Server started...")
	go inputMonitor(sc)
	<-sc
	fmt.Println("Server stopping...")
	api.StopServer(ctx)
	<-quit
	fmt.Println("Server stopped.")
}
