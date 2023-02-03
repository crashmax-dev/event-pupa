# EventLoop

Package for creating and managing event-based systems asynchronously.

## Features

- Different types of events:
  - Run events at trigger
  - Run events at interval
  - Run events with delay
  - Run a one-time event that deletes itself after execution
  - Run events that depend on the triggering of other events
  - And combine different types of events in any combination
- gRPC HTTP API (WIP)
  - For now there is old REST API, created with `net/http` standard library
- Logging:
  - Zap used, but can be easily switched to another logger, just need to implement interface)

## Installation

Simply run `go get gitlab.com/YSX/eventloop@latest` and you good to go.

There is an example of how to start EventLoop and an API in `cmd/make`.