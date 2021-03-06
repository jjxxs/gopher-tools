package system

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

type Filesystem struct {
	Filesystem string `json:"filesystem"`
	Capacity   uint64 `json:"capacity"` // total capacity in bytes
	Used       uint64 `json:"used"`     // usage in bytes
	MountedOn  string `json:"mountedOn"`
}

// Filesystems - Retrieves mounted filesystems as reported by the 'df'-
// utility. Uses Posix-Format for compatibility and ignored 'tmpfs'- and
// 'devtmpfs'-filesystems.
func Filesystems() (fss []Filesystem, err error) {
	var bs []byte
	if bs, err = exec.Command("df", "-P", "-xtmpfs", "-xdevtmpfs").Output(); err == nil {
		if ss := strings.Split(string(bs), "\n"); len(ss) < 2 {
			err = fmt.Errorf("invalid file format")
		} else {
			for i := 1; i < len(ss); i++ { // disregard header-line
				if len(ss[i]) == 0 {
					continue // ignore empty (trailing) lines
				}
				fs := Filesystem{}
				if f := strings.Fields(ss[i]); len(f) < 6 {
					err = fmt.Errorf("invalid file format")
				} else if fs.Capacity, err = strconv.ParseUint(f[1], 10, 64); err != nil {
					err = fmt.Errorf("invalid file format")
				} else if fs.Used, err = strconv.ParseUint(f[2], 10, 64); err != nil {
					err = fmt.Errorf("invalid file format")
				} else {
					fs.Filesystem = f[0]
					fs.MountedOn = f[5]
					fss = append(fss, fs)
				}
			}
		}
	}
	return
}

// CpuLoad - Retrieves the average cpu load over past 1, 5, 15 minutes.
// Uses /proc/loadavg to access this information. Returns error, if the
// file couldn't be found or had an unexpected format.
func CpuLoad() (avg1 float32, avg5 float32, avg15 float32, err error) {
	var bs []byte
	if bs, err = ioutil.ReadFile("/proc/loadavg"); err == nil {
		if avgs := strings.Fields(string(bs)); len(avgs) < 3 {
			err = fmt.Errorf("invalid file format")
		} else if avg1, err = parseFloat32(avgs[0]); err != nil {
			err = fmt.Errorf("invalid file format")
		} else if avg5, err = parseFloat32(avgs[1]); err != nil {
			err = fmt.Errorf("invalid file format")
		} else if avg15, err = parseFloat32(avgs[2]); err != nil {
			err = fmt.Errorf("invalid file format")
		}
	}
	return
}

func parseFloat32(s string) (f float32, err error) {
	var f64 float64
	if f64, err = strconv.ParseFloat(s, 64); err == nil {
		f = float32(f64)
	}
	return
}

// Uptime - Retrieves the systems uptime in seconds. Uses '/proc/uptime'
// to access this information. Returns error, if the file couldn't be
// or had an unexpected format.
func Uptime() (up int, err error) {
	var bs []byte
	if bs, err = ioutil.ReadFile("/proc/uptime"); err == nil {
		if ups := strings.Fields(string(bs)); len(ups) < 2 {
			err = fmt.Errorf("invalid file format")
		} else if ups = strings.Split(ups[0], "."); len(ups) < 2 {
			err = fmt.Errorf("invalid file format")
		} else if up, err = strconv.Atoi(ups[0]); err != nil {
			err = fmt.Errorf("invalid file format")
		}
	}
	return
}
