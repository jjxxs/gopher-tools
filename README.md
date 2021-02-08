<p align="center">
  <img width="270" height="294" src="https://github.com/jjxxs/gopher-tools/blob/media/.github/media/gopher_tools_small.png">
</p>

# gopher-tools
[![Build Status](https://travis-ci.org/jjxxs/gopher-tools.svg?branch=develop)](https://travis-ci.org/jjxxs/gopher-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjxxs/gopher-tools)](https://goreportcard.com/report/github.com/jjxxs/gopher-tools)
[![Release](https://img.shields.io/github/v/release/jjxxs/gopher-tools.svg)](https://github.com/jjxxs/gopher-tools/releases/latest)
[![License](https://img.shields.io/github/license/jjxxs/gopher-tools)](/LICENSE)

A collection of tools and components mostly for use in backend-applications. The design goal is to provide bare-bone
functionality with fitting abstractions to easily extend and adapt to common use-cases.

## Packages
All packages provide tests, and some provide benchmarks. Run  ```go test gopher-tools``` or ```gobench gopher-tools``` 
to run the tests or to benchmark the packages on your system.

## Bus
A ```Bus``` provides an implementation of a loosely-coupled publish-subscriber pattern. ```Receiver``` can
subscribe to the ```Bus``` and are called whenever a Message is published. By default the ```Message```- and 
```Receiver```-type is ```interface{}```. If necessary, this can be changed in ```bus.go```.
```go
type Message = interface{}
type Receiver = func(msg Message)
```

##### Subscribe & Publish
Calls to ```Publish``` will place the ```Message``` in a queue. ```Publish``` returns immediately unless the queue is full,
in which case it will block. ```Message``` are delivered to subscribed ```Receiver``` from within a single go-routine. The
```Receiver``` should be slim for ideal performance.
```go
bus.MsgQueueSize = 2000 // defaults to 1000 if not set
myBus := bus.NewBus()
myBus.Subscribe(func(msg string) {
    fmt.Println(msg)
})
myBus.Publish("busMessage")
```

##### NamedBus
The ```GetNamedBus```-function acts as a factory for ```Bus```-singletons. Repeated calls with the same
name always return the same ```Bus```.
```go
myNamedBus := bus.GetNamedBus("webEventBus")
```

#### Performance
```
CPU:    Intel i7-8700k
goos:   linux
goarch: amd64

Message-Type         Subscribers     Msgs/s
-------------------------------------------------------------------------------
Primitive                      1   14886969   76.6 ns/op   8 B/op   1 allocs/op
Primitive                   1000     737824   1655 ns/op   8 B/op   1 allocs/op
Struct by Value                1   11013112    106 ns/op  64 B/op   1 allocs/op
Struct by Value             1000     725788   1723 ns/op  64 B/op   1 allocs/op
Struct by Reference            1   18971588   66.0 ns/op   0 B/op   0 allocs/op
Struct by Reference         1000     763729   1665 ns/op   0 B/op   0 allocs/op
```

## Environment
Access environment-variables. Use ```GetEnvironmentOrDefault``` to provide a default which will be 
returned, if the given variable is not set. Use ```GetEnvironmentOrPanic``` to ```panic()``` when the
given variable is not set. Example:
```go
// get 'myEnvVar' or use a default if not set
varOrDefault := config.GetEnvironmentOrDefault("myEnvVar", "myEnvVarDefault")

// panic if 'myEnvVar' is not set
varOrPanic := config.GetEnvironmentOrPanic("myEnvVar")
```

## Errors
Trivial ```string```-constants for frequently needed error-messages. This includes messages for database- and json-
related errors. Example:
```go
if err := myDatabaseProvider.Connect(); err != nil {
    fmt.Println(errors.DatabaseFailedToConnect)
}
```

## Execution
Add documentation here.

## Server
Add documentation here.

## Signal
Add documentation here.