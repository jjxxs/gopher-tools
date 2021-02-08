package signal

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHandlerShouldCallCallbacks(t *testing.T) {
	handler := NewHandler()
	countSigUsr1 := 0
	countSigUsr2 := 0
	other := 0
	handler.Register(func(signal os.Signal) {
		switch signal {
		case syscall.SIGUSR1:
			countSigUsr1++
		case syscall.SIGUSR2:
			countSigUsr2++
		default:
			other++
		}
	}, syscall.SIGUSR1, syscall.SIGUSR2)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	_ = p.Signal(syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if countSigUsr1 != 1 {
		t.Fail()
	}
	if countSigUsr2 != 1 {
		t.Fail()
	}
	if other != 0 {
		t.Fail()
	}
}

func TestHandlerShouldNotCallAfterUnregister(t *testing.T) {
	handler := NewHandler()
	countSigUsr1 := 0
	countSigUsr2 := 0
	other := 0
	cb := func(signal os.Signal) {
		switch signal {
		case syscall.SIGUSR1:
			countSigUsr1++
		case syscall.SIGUSR2:
			countSigUsr2++
		default:
			other++
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
	if countSigUsr1 != 1 {
		t.Fail()
	}
	if countSigUsr2 != 1 {
		t.Fail()
	}
	if other != 0 {
		t.Fail()
	}
}
