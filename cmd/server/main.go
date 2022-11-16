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
		_, _ = fmt.Scanln(&input)
	}
	sc <- syscall.SIGTERM
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := api.StartServer(zapcore.InfoLevel)
		if err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Server started...")
	go inputMonitor(sc)
	<-sc
	fmt.Println("Server stopping...")
	if err := api.StopServer(ctx); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Server stopped.")
}
