package signal

import (
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {
	var count atomic.Int32
	Handle(func(signal os.Signal) {
		count.Add(1)
	}, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if count.Load() != 2 {
		t.Fail()
	}
}

func TestHandleOneShot(t *testing.T) {
	var count atomic.Int32
	handlerFunc := func(sig os.Signal) {
		count.Add(1)
	}
	HandleOneShot(handlerFunc, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if count.Load() != 1 {
		t.Fail()
	}
}
