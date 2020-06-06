## gopher-tools
[![Build Status](https://travis-ci.org/jjxxs/gopher-tools.svg?branch=master)](https://travis-ci.org/jjxxs/gopher-tools)

A collection of tools and components for use in the backend. The design goal is to provide bare-bone functionality with fitting abstractions to easily extend and adapt to new use-cases.

## Overview

### Bus

##### bus.GetBusFromFactory()
Provides thread-safe access a bus with a specified name. This can be used as a store of named bus-singletons.

##### bus.Bus
A Bus provides a loosely coupled implementation of the publish-subscriber pattern.

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