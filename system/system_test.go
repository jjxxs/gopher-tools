package system

import (
	"fmt"
	"testing"
)

// these are just smoke-tests since the results
// are highly dependent on the executing platform

func TestFilesystems(t *testing.T) {
	if _, err := Filesystems(); err != nil {
		t.Fail()
	}
}

func TestCpuLoad(t *testing.T) {
	if _, err := CpuLoad(); err != nil {
		t.Fail()
	}
}

func TestUptime(t *testing.T) {
	if x, err := Uptime(); err != nil {
		t.Fail()
	} else {
		fmt.Println(x)
	}
}

func TestUptimeMs(t *testing.T) {
	if x, err := UptimeMs(); err != nil {
		t.Fail()
	} else {
		fmt.Println(x)
	}
}
