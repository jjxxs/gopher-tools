package signal

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {
	count := 0
	Handle(func(signal os.Signal) {
		count++
	}, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if count != 2 {
		t.Fail()
	}
}

func TestHandleOneShot(t *testing.T) {
	count := 0
	handlerFunc := func(sig os.Signal) {
		count++
	}
	HandleOneShot(handlerFunc, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	_ = p.Signal(syscall.SIGUSR1)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if count != 1 {
		t.Fail()
	}
}
