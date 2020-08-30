package global

import (
	"context"
	"errors"
	"sync"
)

const shutdownGroupKey = "shutdownGroup"

// Provides a context with a cancel-func that carries a WaitGroup.
// Callbacks can be registered to this context and will be called
// when the context is cancelled.
func GetShutdownContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	if val := ctx.Value(shutdownGroupKey); val == nil {
		ctx = context.WithValue(ctx, shutdownGroupKey, &sync.WaitGroup{})
	}
	return ctx, cancel
}

// Blocks until all callbacks were called in the WaitGroup carried
// by the given context.
func WaitForShutdownContext(ctx context.Context) error {
	if val := ctx.Value(shutdownGroupKey); val != nil {
		if shutdownGroup, ok := val.(*sync.WaitGroup); ok {
			shutdownGroup.Wait()
		} else {
			return errors.New("value has invalid type")
		}
	} else {
		return errors.New("value not found")
	}
	return nil
}

// When the given context is cancelled through its cancel-func, the given
// callback will be called.
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
			return errors.New("value has invalid type")
		}
	} else {
		return errors.New("value not found")
	}
	return nil
}
