package signal

import (
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestHandlerShouldCallCallbacks(t *testing.T) {
	handler := NewHandler()
	var countSigUsr1, countSigUsr2, other atomic.Int32
	handler.Register(func(signal os.Signal) {
		switch signal {
		case syscall.SIGUSR1:
			countSigUsr1.Add(1)
		case syscall.SIGUSR2:
			countSigUsr2.Add(1)
		default:
			other.Add(1)
		}
	}, syscall.SIGUSR1, syscall.SIGUSR2)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	_ = p.Signal(syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if countSigUsr1.Load() != 1 {
		t.Fail()
	}
	if countSigUsr2.Load() != 1 {
		t.Fail()
	}
	if other.Load() != 0 {
		t.Fail()
	}
}

func TestHandlerShouldNotCallAfterUnregister(t *testing.T) {
	handler := NewHandler()
	var countSigUsr1, countSigUsr2, other atomic.Int32
	cb := func(signal os.Signal) {
		switch signal {
		case syscall.SIGUSR1:
			countSigUsr1.Add(1)
		case syscall.SIGUSR2:
			countSigUsr2.Add(1)
		default:
			other.Add(1)
		}
	}
	handler.Register(cb, syscall.SIGUSR1, syscall.SIGUSR2)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1) // these should be received
	_ = p.Signal(syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	handler.Unregister(cb)
	_ = p.Signal(syscall.SIGUSR1) // these shouldn't be received
	_ = p.Signal(syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if countSigUsr1.Load() != 1 {
		t.Fail()
	}
	if countSigUsr2.Load() != 1 {
		t.Fail()
	}
	if other.Load() != 0 {
		t.Fail()
	}
}

func TestHandlerShouldNotCallAfterExit(t *testing.T) {
	handler := NewHandler()
	var count atomic.Int32
	handler.Register(func(signal os.Signal) {
		count.Add(1)
	}, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)      // should be received
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	handler.Exit()
	_ = p.Signal(syscall.SIGUSR1)      // should not be received
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if count.Load() != 1 {
		t.Fail()
	}
}
