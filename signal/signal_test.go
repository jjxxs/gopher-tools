package signal

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHandlerShouldCallCallbacks(t *testing.T) {
	handler := NewHandler()
	countSigAlrm := 0
	countSigHup := 0
	other := 0
	handler.Register(func(signal os.Signal) {
		switch signal {
		case syscall.SIGALRM:
			countSigAlrm++
		case syscall.SIGHUP:
			countSigHup++
		default:
			other++
		}
	}, syscall.SIGALRM, syscall.SIGHUP)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGALRM)
	_ = p.Signal(syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if countSigAlrm != 1 {
		t.Fail()
	}
	if countSigHup != 1 {
		t.Fail()
	}
	if other != 0 {
		t.Fail()
	}
}

func TestHandlerShouldNotCallAfterUnregister(t *testing.T) {
	handler := NewHandler()
	countSigAlrm := 0
	countSigHup := 0
	other := 0
	cb := func(signal os.Signal) {
		switch signal {
		case syscall.SIGALRM:
			countSigAlrm++
		case syscall.SIGHUP:
			countSigHup++
		default:
			other++
		}
	}
	handler.Register(cb, syscall.SIGALRM, syscall.SIGHUP)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGALRM) // these should be received
	_ = p.Signal(syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	handler.Unregister(cb)
	_ = p.Signal(syscall.SIGALRM) // these shouldn't be received
	_ = p.Signal(syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond) // signals are delivered async, wait a little
	if countSigAlrm != 1 {
		t.Fail()
	}
	if countSigHup != 1 {
		t.Fail()
	}
	if other != 0 {
		t.Fail()
	}
}
