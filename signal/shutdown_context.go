package signal

import (
	"context"
	"errors"
	"sync"
)

// Provides means to initiate an application-wide shutdown. Since usually applications are
// composed of various independently acting modules, e.g. database-access, web-handling etc. it is desirable
// to have a mechanism to shut all those modules down gracefully.
//
// Create a new ShutdownContext with GetShutdownContext. Pass this context to your modules, they should register
// their shutdown-routines with RegisterOnShutdownCallback. If the cancel-function initially provided by
// GetShutdownContext  is called, all registered callbacks are invoked. Use WaitForShutdownContext to wait until
// all registered callbacks have finished.

type key string

const shutdownGroupKey key = "github.com/jjxxs/gopher-tools/signal/shutdownGroupKey"

var invalidWaitGroup = errors.New("context has invalid wait-group")
var hasNoWaitGroup = errors.New("context has no wait-group")

// GetShutdownContext decorates a given context and returns it together
// with a cancel-function. RegisterOnShutdownCallback can now be used
// to register callbacks, that are called when the cancel-function is called.
// Use WaitForShutdownContext to wait for all registered callbacks to finish
// their execution.
func GetShutdownContext(ctx context.Context) (shutdownCtx context.Context, cancel func()) {
	shutdownCtx, cancel = context.WithCancel(ctx)
	if val := shutdownCtx.Value(shutdownGroupKey); val == nil {
		shutdownCtx = context.WithValue(shutdownCtx, shutdownGroupKey, &sync.WaitGroup{})
	}
	return shutdownCtx, cancel
}

// WaitForShutdownContext blocks until all registered callbacks were called.
func WaitForShutdownContext(ctx context.Context) error {
	if val := ctx.Value(shutdownGroupKey); val != nil {
		if shutdownGroup, ok := val.(*sync.WaitGroup); ok {
			shutdownGroup.Wait()
		} else {
			return invalidWaitGroup
		}
	} else {
		return hasNoWaitGroup
	}
	return nil
}

// RegisterOnShutdownCallback registers a function that is called when the
// contexts cancel-function is called.
func RegisterOnShutdownCallback(ctx context.Context, callback func()) error {
	if val := ctx.Value(shutdownGroupKey); val != nil {
		if shutdownGroup, ok := val.(*sync.WaitGroup); ok {
			shutdownGroup.Add(1)
			go func() {
				<-ctx.Done()
				callback()
				shutdownGroup.Done()
			}()
		} else {
			return invalidWaitGroup
		}
	} else {
		return hasNoWaitGroup
	}
	return nil
}
