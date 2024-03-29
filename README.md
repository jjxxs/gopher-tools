<p align="center">
  <img width="270" height="294" src="https://github.com/jjxxs/gopher-tools/blob/master/.github/media/gopher_tools_small-transparent.png">
</p>

# gopher-tools
[![Build Status](https://travis-ci.org/jjxxs/gopher-tools.svg?branch=develop)](https://travis-ci.org/jjxxs/gopher-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjxxs/gopher-tools)](https://goreportcard.com/report/github.com/jjxxs/gopher-tools)
[![Release](https://img.shields.io/github/v/release/jjxxs/gopher-tools.svg)](https://github.com/jjxxs/gopher-tools/releases/latest)
[![License](https://img.shields.io/github/license/jjxxs/gopher-tools)](/LICENSE)

A small collection of components mostly for backend-applications. The design goal is to keep it simple and straightforward with fitting abstractions to easily adapt to common use-cases. All packages provide tests which can be run with ```go test gopher-tools```. Some packages provide benchmarks which can
be run with ```gobench gopher-tools```.

## Websocket

### Connection
Wraps the ```RFC-6455```-compliant but relatively low-level ```gorilla/websocket``` websocket-library. Supports registration of ```onmessage```-,  ```error```- and ```close```-handlers. Connections can be closed via ```Shutdown```.
```go
onMessage := func(this Connection, msgType int, data []byte) { 
	this.Send(gorilla.TextMessage, []byte{"Hello, Gopher!"})
}
onClose := func(this Connection, code int, text string) { }
onError := func(this Connection, err error) { }

conn := NewConnection(getGorillaConnection(), onMessage, onClose, onError)
conn.Close() // gracefully closes connection within 1 second, otherwise kills it
```

##### Performance
```TestConcurrentConnections``` in ```connection_test.go``` can be used for performance-testing. It emulates a high-load situation
in which server- and client-websockets rapidly exchange messages. Server- and client-sides are both handled via the ```Connection```-type. 
Results for 20.000 concurrent connections (40.000 ```Connection```-objects) being handled by an ```Intel i7-8700k``` where each connection
round-trips a message for 100 times:
```
=== RUN   TestConcurrentConnections
    connection_test.go:126: 
        concurrent connections:     20000
        round-trips per connection: 100
        total messages exchanged:   4000000
        total time:                 6.7504401s
        avg time per round-trip:    3.375µs
        avg time per message:       1.687µs
--- PASS: TestConcurrentConnections (24.31s)
```

## Bus
A ```Bus``` provides an implementation of a loosely-coupled publish-subscriber
pattern. Subscribers can subscribe to the Bus and are called whenever a
Message is published. A ```WorkerBus``` starts a dedicated go-routine for every
subscriber while ```Bus``` delivers with the routine that called ```Publish```.

##### Singletons
There are simple singletons and named singletons. Using these singletons is optional.
```go
myBusSingleton := GetBus()
myNamedBusSingleton := GetNamedBus("wsRequests")
myWorkerBusSingleton := GetWorkerBus()
myNamedWorkerBusSingleton := GetNamedWorkerBus("wsLongRequests")
```

##### Performance
```
goos: linux
goarch: amd64
pkg: github.com/jjxxs/gopher-tools/bus
cpu: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
BenchmarkBusPublishPrimitive__1_Subs
BenchmarkBusPublishPrimitive__1_Subs-12           	62088909	        19.21 ns/op	       0 B/op	       0 allocs/op
BenchmarkBusPublishPrimitive__1000_Subs
BenchmarkBusPublishPrimitive__1000_Subs-12        	  251073	      4880 ns/op	       0 B/op	       0 allocs/op
BenchmarkBusPublishStructByValue__1_Subs
BenchmarkBusPublishStructByValue__1_Subs-12       	48561636	        24.38 ns/op	       0 B/op	       0 allocs/op
BenchmarkBusPublishStructByValue__1000_Subs
BenchmarkBusPublishStructByValue__1000_Subs-12    	  169450	      7065 ns/op	       0 B/op	       0 allocs/op
BenchmarkBusPublishReference__1_Subs
BenchmarkBusPublishReference__1_Subs-12           	63486026	        18.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkBusPublishReference__1000_Subs
BenchmarkBusPublishReference__1000_Subs-12        	  254914	      4681 ns/op	       0 B/op	       0 allocs/op
```

## Signal

### Handler
A ```Handler``` invokes registered callbacks when a matching ```os.Signal``` is received. It is easily accessed via
the provided Singleton which can be used by calling the ```Handle```- or the ```HandleOneShot```function.
```go
Handle(func(sig os.Signal) {
	switch sig {
	case syscall.SIGTERM:
		os.Exit(0)
	case syscall.SIGALRM:
		os.Exit(1)
    }
}, syscall.SIGTERM, syscall.SIGALRM)
```