package hw

import (
	"errors"
	"os"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"
)

var envI2cBase string

const i2cSlaveReq = 0x0703

type Device interface {
	GetBus() int
	GetAddr() int
	WriteBytes(buf []byte) (int, error)
	ReadBytes(buf []byte) (int, error)
	Exit()
}

type deviceImpl struct {
	addr  int
	bus   int
	file  *os.File
	mutex *sync.Mutex
}

func NewDevice(addr int, bus int) (Device, error) {
	// get base-dir from environment
	if base, present := os.LookupEnv("i2cBase"); !present {
		return nil, errors.New("env 'i2cBase' not set")
	} else {
		envI2cBase = base + strconv.FormatInt(int64(bus), 10)
	}

	if file, err := os.OpenFile(envI2cBase, os.O_RDWR, 0600); err != nil {
		return nil, err
	} else if err := unix.IoctlSetInt(int(file.Fd()), i2cSlaveReq, int(addr)); err != nil {
		return nil, err
	} else {
		device := &deviceImpl{
			addr:  addr,
			bus:   bus,
			file:  file,
			mutex: &sync.Mutex{},
		}

		return device, nil
	}
}

func (d *deviceImpl) GetBus() int {
	return d.bus
}

func (d *deviceImpl) GetAddr() int {
	return d.addr
}

func (d *deviceImpl) WriteBytes(buf []byte) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.file.Write(buf)
}

func (d *deviceImpl) ReadBytes(buf []byte) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.file.Read(buf)
}

func (d *deviceImpl) Exit() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_ = d.file.Close()
}
