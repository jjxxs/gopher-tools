package system

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

type Filesystem struct {
	Filesystem string `json:"filesystem"`
	Total      uint64 `json:"total"` // total capacity in bytes
	Used       uint64 `json:"used"`  // usage in bytes
	MountedOn  string `json:"mountedOn"`
}

// Filesystems retrieves mounted filesystems via 'df'.
func Filesystems() (fss []Filesystem, err error) {
	var bs []byte
	if bs, err = exec.Command("df", "-P").Output(); err == nil {
		if ss := strings.Split(string(bs), "\n"); len(ss) == 0 {
			err = fmt.Errorf("invalid file format")
		} else {
			for i := 1; i < len(ss); i++ { // skip header-line
				if len(ss[i]) == 0 {
					continue // ignore empty lines (trailing)
				}
				var j, skip1 int
				var skip2 string
				fs := Filesystem{}
				if j, err = fmt.Sscanf(ss[i], "%s %d %d %d %s %s", &fs.Filesystem, &fs.Total, &fs.Used, &skip1, &skip2, &fs.MountedOn); err != nil || j != 6 {
					err = fmt.Errorf("failed to parse output")
				}
				fss = append(fss, fs)
			}
		}
	}
	return
}

type CpuLoadAverages struct {
	Avg1  float32 `json:"avg1"`
	Avg5  float32 `json:"avg5"`
	Avg15 float32 `json:"avg15"`
}

// CpuLoad retrieves the average cpu load over past 1, 5, 15 minutes via /proc/loadavg.
func CpuLoad() (load CpuLoadAverages, err error) {
	var bs []byte
	if bs, err = ioutil.ReadFile("/proc/loadavg"); err == nil {
		var j int
		if j, err = fmt.Sscanf(string(bs), "%f %f %f", &load.Avg1, &load.Avg5, &load.Avg15); err != nil || j != 3 {
			err = fmt.Errorf("failed to parse output")
		}
	}
	return
}

// Uptime retrieves the system uptime in seconds via /proc/uptime.
func Uptime() (up float64, err error) {
	var bs []byte
	if bs, err = ioutil.ReadFile("/proc/uptime"); err == nil {
		var j int
		if j, err = fmt.Sscanf(string(bs), "%f", &up); err != nil || j != 1 {
			err = fmt.Errorf("failed to parse output")
		}
	}
	return
}
