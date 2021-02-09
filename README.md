<p align="center">
  <img width="270" height="294" src="https://github.com/jjxxs/gopher-tools/blob/media/.github/media/gopher_tools_small.png">
</p>

# gopher-tools
[![Build Status](https://travis-ci.org/jjxxs/gopher-tools.svg?branch=develop)](https://travis-ci.org/jjxxs/gopher-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjxxs/gopher-tools)](https://goreportcard.com/report/github.com/jjxxs/gopher-tools)
[![Release](https://img.shields.io/github/v/release/jjxxs/gopher-tools.svg)](https://github.com/jjxxs/gopher-tools/releases/latest)
[![License](https://img.shields.io/github/license/jjxxs/gopher-tools)](/LICENSE)

A small collection of components mostly for backend-applications. 
 
The design goal is to keep it simple and straightforward with fitting abstractions to easily adapt to common use-cases.

## Packages
All packages provide tests which can be run with ```go test gopher-tools```. Some packages provide benchmarks which can
be run with ```gobench gopher-tools```.

## Websocket

### Connection
Represents a websocket-connection. Wraps the ```RFC-6455```-compliant but relatively low-level ```gorilla/websocket```
library. Supports registration of handlers for ```error```- and ```close```-events. Gracefully closes the connection via
the ```Shutdown```-method. Will call the provided ```onMessage```-handler whenever a message is received.
```go
c := getYourGorillaWebsocketConnection() // get your gorilla/websocket-connection

onMessage := func(this Connection, msgType int, data []byte) {
    fmt.Println("received message", string(data))
    this.Send(msgType, data) // echo the message
}
onError := func(this Connection, err error) {
    fmt.Println("error occurred", err.Error())
}
onClose := func(this Connection, err websocket.CloseError) {
    fmt.Println("connection closed", err.Code, err.Text)
}

conn := NewConnection(c, onMessage)
conn.OnError(onError) // optional
conn.OnClose(onClose) // optional
fmt.Println("new connection with address", conn) // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
conn.Shutdown() // gracefully closes the connection within 1 second or kills it otherwise
```

##### Performance
```TestConcurrentConnections``` within ```connection_test.go``` can be used for performance-testing. During the test, a server is started which
accepts incoming websocket-connections. The test then proceeds to connect a specified amount of websockets to
the server. Both server- and client-sided connection-endpoints are handled via ```Connection```-structs. Once all connections
are established, a message is exchanged repeatedly between server- and client-endpoints. This emulates a high-load situation
in which a server must handle many active connections sending messages at the same time.

The following results show the test with 20.000 concurrent connections, which makes for 40.000 ```Connection```-endpoints 
being handled by a single ```Intel i7-8700k```. During the test, each connection round-tripped a message 100 times.
```
=== RUN   TestConcurrentConnections
--- PASS: TestConcurrentConnections (24.47s)
    connection_test.go:134: 
        concurrent connections:       20000
        round-trips per connection:     100
        total messages exchanged:   4000000
        total time:                    8.56s
        avg time per round-trip:       4.28µs
        avg time per message:          2.14µs
```

### Broadcaster
Aggregates ```Connection``` for message-broadcasting. ```Connection``` can be added with ```Add``` and removed with 
```Remove```. By calling ```Broadcast``` a given message is sent to all connections that are currently added to the 
```Broadcast```. 

### DemilitarizedUpgrader
Upgrades a ```http``` into a ```ws```-connection. Offers no protection against cross site request forgery (```csrf```). Only
use this ```Upgrader``` during development or within demilitarized zones.

## Bus
A ```Bus``` provides an implementation of a loosely-coupled publish-subscriber pattern. Subscribers subscribe to the
```Bus``` via ```Subscribe```. Messages published via ```Publish``` are forwarded to all subscribers. Default types of
subscriber-functions and messages are ```func(interface{})``` and ```interface{}``` but can be changed in ```bus.go```
if necessary.

##### Subscribe & Publish
```Publish``` places the provided message in a queue which is picked up by a single go-routine and forwarded to
every subscriber. ```Publish``` returns immediately unless the queue is full, in which case it will block until the
message can be placed. Use ```PublishTimeout``` to only block for a given duration in case the queue is full. 
Subscriber-functions should be kept slim for ideal performance.
```go
bus.MsgQueueSize = 2000 // defaults to 1000 if not set
myBus := bus.NewBus()
myBus.Subscribe(func(msg string) {
    fmt.Println(msg) // keep subscribers slim
})
myBus.Publish("busMessage") // blocks if the queue is full
success := myBus.PublishTimeout("busMessage", 100 * time.Millisecond) // only blocks for 100ms if the queue is full
```

##### NamedBus
The ```GetNamedBus```-function acts as a factory for ```Bus```-singletons. Repeated calls with the same
name always return the same ```Bus```.
```go
myNamedBus := bus.GetNamedBus("webEventBus")
```

##### Performance
```
CPU:    Intel i7-8700k
goos:   linux
goarch: amd64

Message-Type         Subscribers       Msgs/s
-------------------------------------------------------------------------------
Primitive                      1   15,524,827   70.1 ns/op   8 B/op   1 allocs/op
Primitive                   1000      770,786   1656 ns/op   8 B/op   1 allocs/op
Struct by Value                1   12,343,693    111 ns/op  64 B/op   1 allocs/op
Struct by Value             1000      668,594   1780 ns/op  64 B/op   1 allocs/op
Struct by Reference            1   20,156,332   62.8 ns/op   0 B/op   0 allocs/op
Struct by Reference         1000      755,595   1691 ns/op   0 B/op   0 allocs/op
```

## Environment
Use ```GetEnvironmentOrDefault``` to retrieve an environment-variable or return a default 
if the given variable is not set. Use ```GetEnvironmentOrPanic``` to retrieve an environment-variable or ```panic()``` 
if the variable is not set.
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

## Signal

### Handler
```Handler``` provides functionality to ```Register``` callbacks that are invoked when the application receives a given ```os.Signal```.
Use ```Register``` to register a callback and ```Unregister``` to unregister it. ```Exit``` makes the handler stop listening
for signals. Once ```Exit``` was called, the ```Handler``` becomes useless.
```go
handler := NewHandler()
handler.Register(func(sig os.Signal) {
    switch sig {
    case syscall.SIGTERM:
        fmt.Println("received SIGTERM")
    case syscall.USR1:
        fmt.Println("received USR1")
}
}, syscall.SIGTERM, syscall.USR1)
```

##### Handle
```Handle``` registers a callback with a ```Handler```-singleton. The singleton is created when the function is executed 
for the first time.
```go
Handle(func(sig os.Signal) {
    fmt.Println("received SIGTERM in handler-singleton")
}, syscall.SIGTERM)
```
  
### ShutdownContext
```GetShutdownContext``` decorates a given ```context.Context``` and returns it together with a cancel-function. 
```RegisterOnShutdownCallback``` can be used on a previously decorated context to register callbacks that are called,
when the cancel-function is executed. Use ```WaitForShutdownContext``` to wait for all registered callbacks to finish
their execution.
```go
ctx, cancel := GetShutdownContext(context.Background)
RegisterOnShutdownCallback(ctx, yourDatabaseShutdownFunction)
RegisterOnShutdownCallback(ctx, yourWebserverShutdownFunction)

Handle(func(sig os.Signal) { // use handle to react to SIGTERM
    fmt.Println("received shutdown-signal, gracefully shutting down...")
    cancel() // cancelling the context will invoke all previously registered functions  
    WaitForShutdownContext(ctx) // wait for the functions to finish (e.g. closing connections to database)
}, syscall.SIGTERM)
```
