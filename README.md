## gopher-tools
[![Build Status](https://travis-ci.org/jjxxs/gopher-tools.svg?branch=develop)](https://travis-ci.org/jjxxs/gopher-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjxxs/gopher-tools)](https://goreportcard.com/report/github.com/jjxxs/gopher-tools)
[![Release](https://img.shields.io/github/v/release/jjxxs/gopher-tools.svg)](https://github.com/jjxxs/gopher-tools/releases/latest)
[![License](https://img.shields.io/github/license/jjxxs/gopher-tools)](/LICENSE)

A collection of tools and components for use in the backend. The design goal is to provide bare-bone functionality with fitting abstractions to easily extend and adapt to the most common use-cases.

## Overview

### Bus
A Bus provides an implementation of a loosely-coupled publish-subscriber pattern. Interested consumers can subscribe with a function that will be called with messaged published on the bus. For every consumer a dedicated go-routine is employed to asynchronously deliver the message. The implementation uses a Queue to buffer pending messages for every consumer.

The benchmarks show the performance for the most commonly used messages, e.g. primitive types, structs by value and structs by reference.
```
BenchmarkPublishPrimitiveArgsOneSubscriber-12              	11194382	        99 ns/op	       8 B/op	       1 allocs/op
BenchmarkPublishStructByValueOneSubscriber-12            	 8781723	       134 ns/op	      64 B/op	       1 allocs/op
BenchmarkPublishStructByValueOneHundredSubscriber-12     	   79690	     14770 ns/op	      64 B/op	       1 allocs/op
BenchmarkPublishReferenceOneSubscriber-12                	14176437	        89 ns/op	       0 B/op	       0 allocs/op
BenchmarkPublishReferenceOneHundredSubscriber-12         	   79422	     14813 ns/op	       0 B/op	       0 allocs/op
```

Please note that a separate go-routine and queue is maintained for every subscribed consumer. This allows for some processing to happen in the subscribed function but slows down the bus for lightweight message-delivery.

### Config

##### config.Config
A config provides means to access a configuration-object

##### config.jsonConfig
A json-config represents a json-formatted file-based configuration.

### Server

##### server.Server
A server serves static files via http and can also accept incoming websocket-connections.

##### server.Connection
A connection represents a connection made by the server, e.g. a connected websocket.

##### server.bufferedWebsocketConnection
A buffered websocket-connection buffers incoming/outgoing messages before they are processed.

##### server.ConnectionHandler
A connection-handler handles connections made by the server.

##### server.multiplexConnectionHandler
A multiplex connection-handler aggregates multiple connection-handlers into a single connection-handler.

### Signal

##### signal.Handle()
Handles a specified signal with the given function. This is a comfortable interface for the most common use-case. Uses a signal-Handler singleton.

##### signal.Handler
A handler provides means to register functions that are called when the application receives a specified signal.