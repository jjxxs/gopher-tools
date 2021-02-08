package hw

/*
import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type pinMode []byte
type pinState []byte

var (
	instance Gpio = nil
	once     sync.Once
)

// envGpioBase is set by the environment-variable 'gpioBase'
var envGpioBase, gpioExport, gpioUnexport string

var (
	pinInput  pinMode  = []byte("in")  // 0x69, 0x6e
	pinOutput pinMode  = []byte("out") // 0x6f, 0x75, 0x74
	pinHigh   pinState = []byte("1")   // 0x31
	pinLow    pinState = []byte("0")   // 0x30
)

type Pin interface {
	GetPinId() model.PinId
	SetMode(mode model.IOMode) error
	GetMode() (model.IOMode, error)
	SetState(state model.State) error
	GetState() (model.State, error)
	Toggle() error
	getBaseDir() string
	getValueFile() *os.File
	getModeFile() *os.File
}

type pinImpl struct {
	Id        model.PinId
	baseDir   string
	mutex     *sync.Mutex
	modeFile  *os.File
	valueFile *os.File
}

type Gpio interface {
	OpenPin(pinId model.PinId) (Pin, error)
	ClosePin(pinId model.PinId) error
}

type gpioImpl struct {
	pins         map[model.PinId]Pin
	mutex        *sync.Mutex
	exportFile   *os.File
	unexportFile *os.File
}

func GetInstance() Gpio {
	initInstance := func() {
		// get base-dir from environment
		if base, present := os.LookupEnv("gpioBase"); !present {
			// TODO: how to get log.WithHistory here?
			panic("env 'gpioBase' ist not set")
		} else {
			envGpioBase = base
			gpioExport = envGpioBase + "/export"
			gpioUnexport = envGpioBase + "/unexport"
		}

		// open file-handles
		if exportFile, err := os.OpenFile(gpioExport, os.O_RDWR, 0770); err != nil {
			panic(err)
		} else if unexportFile, err := os.OpenFile(gpioUnexport, os.O_RDWR, 0770); err != nil {
			panic(err)
		} else {
			instance = &gpioImpl{
				mutex:        &sync.Mutex{},
				pins:         make(map[model.PinId]Pin),
				exportFile:   exportFile,
				unexportFile: unexportFile,
			}
		}
	}

	once.Do(initInstance)

	return instance
}

func (g *gpioImpl) OpenPin(pinId model.PinId) (Pin, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// pin is already open
	if pin, ok := g.pins[pinId]; ok {
		return pin, nil
	}

	// create new pin
	pin := &pinImpl{
		mutex:   &sync.Mutex{},
		Id:      pinId,
		baseDir: path.Join(envGpioBase, fmt.Sprintf("gpio%d", pinId)),
	}

	// try to export pin
	if _, err := g.exportFile.Write([]byte(strconv.Itoa(int(pinId)))); err != nil {
		switch err := err.(type) {
		case *os.PathError:
			if err.Err.Error() != "device or resource busy" {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	// open /sys/class/gpio/gpio<pinId>/direction
	for tries := 0; tries <= 10; tries++ {
		if tries == 10 {
			return nil, fmt.Errorf("failed to export pin#%d, can't open %s", pin.Id, path.Join(pin.baseDir, "direction"))
		}

		if modeFile, err := os.OpenFile(path.Join(pin.baseDir, "direction"), os.O_RDWR, 0770); err != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			pin.modeFile = modeFile
			break
		}
	}

	// open /sys/class/gpio/gpio<pinId>/value
	for tries := 0; tries <= 10; tries++ {
		if tries == 10 {
			return nil, fmt.Errorf("failed to export pin#%d, can't open %s", pin.Id, path.Join(pin.baseDir, "value"))
		}

		if valueFile, err := os.OpenFile(path.Join(pin.baseDir, "value"), os.O_RDWR, 0770); err != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			pin.valueFile = valueFile
			break
		}
	}

	// add pin to map and return
	g.pins[pinId] = pin
	return pin, nil
}

func (g *gpioImpl) ClosePin(pinId model.PinId) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// pin doesn't exist
	if pin, ok := g.pins[pinId]; ok {
		// try to unexport
		if _, err := g.unexportFile.Write([]byte(strconv.Itoa(int(pinId)))); err != nil {
			switch err := err.(type) {
			case *os.PathError:
				if err.Err.Error() != "invalid argument" {
					return err
				}
			default:
				return err
			}
		}

		// wait for /sys/class/gpio<pinId> to disappear
		tries := 0
		for _, err := os.Stat(pin.getBaseDir()); err == nil && tries < 10; tries++ {
			time.Sleep(200 * time.Millisecond)
		}
		if tries == 10 {
			return fmt.Errorf("failed to unexport %d, can stat %s", pin, pin.getBaseDir())
		}

		// close value- & mode-files
		if err1, err2 := pin.getValueFile().Close(), pin.getModeFile().Close(); err1 != nil || err2 != nil {
			return fmt.Errorf("failed to close value- or mode-file for pin#%d, err1=%s, err2=%s", pin.GetPinId(), err1, err2)
		}

		// delete pin mapping
		delete(g.pins, pinId)
	}

	return nil
}

func (g *gpioImpl) Exit() []error {
	retErrs := make([]error, 0)

	// close all pins
	for _, pin := range g.pins {
		if err := g.ClosePin(pin.GetPinId()); err != nil {
			retErrs = append(retErrs, err)
		}
	}

	// close export- & unexport files
	if err1, err2 := g.exportFile.Close(), g.unexportFile.Close(); err1 != nil || err2 != nil {
		if err1 != nil {
			retErrs = append(retErrs, err1)
		} else if err2 != nil {
			retErrs = append(retErrs, err2)
		}
	}

	return retErrs
}

func (p *pinImpl) GetPinId() model.PinId {
	return p.Id
}

func (p *pinImpl) SetMode(mode model.IOMode) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.setMode(mode)
}

func (p *pinImpl) setMode(mode model.IOMode) error {
	// reset stream to start of file
	if o, err := p.modeFile.Seek(0, 0); err != nil || o != 0 {
		return fmt.Errorf("error seeking to start of mode-file, pin=%d, offset=%d, err=%s", p.Id, o, err)
	}

	bMode := pinInput
	if mode == model.Output {
		bMode = pinOutput
	}

	_, err := p.modeFile.Write(bMode)
	return err
}

func (p *pinImpl) GetMode() (model.IOMode, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.getMode()
}

func (p *pinImpl) getMode() (model.IOMode, error) {
	// reset stream to start of file
	if o, err := p.modeFile.Seek(0, 0); err != nil || o != 0 {
		return model.Output, fmt.Errorf("error seeking to start of module-file, pin=%d, offset=%d, err=%s", p.Id, o, err)
	}

	// read the file
	b := make([]byte, 1)
	if l, err := p.modeFile.Read(b); err != nil || l != 1 {
		return model.Output, fmt.Errorf("error reading mode from %d, l=%d, err=%s", p.Id, l, err)
	}

	switch b[0] {
	case pinInput[0]:
		return model.Input, nil
	case pinOutput[0]:
		return model.Output, nil
	default:
		return model.Output, fmt.Errorf("unrecognized mode pin=%d, b=%v", p.Id, b)
	}
}

func (p *pinImpl) SetState(state model.State) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.setState(state)
}

func (p *pinImpl) setState(state model.State) error {
	// reset stream to start of file
	if o, err := p.valueFile.Seek(0, 0); err != nil || o != 0 {
		return fmt.Errorf("error seeking to start of value-file, pin=%d, offset=%d, err=%s", p.Id, o, err)
	}

	bState := pinLow
	if state == model.On {
		bState = pinHigh
	}
	_, err := p.valueFile.Write(bState)
	return err
}

func (p *pinImpl) GetState() (model.State, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.getState()
}

func (p *pinImpl) getState() (model.State, error) {
	// reset stream to start of file
	if o, err := p.valueFile.Seek(0, 0); err != nil || o != 0 {
		return model.UndefinedState, fmt.Errorf("error seeking to start of value-file, pin=%d, offset=%d, err=%s", p.Id, o, err)
	}

	b := make([]byte, 1)
	if l, err := p.valueFile.Read(b); err != nil || l != 1 {
		return model.UndefinedState, fmt.Errorf("error reading state from %d, l=%d, err=%s", p.Id, l, err)
	}

	if b[0] == pinLow[0] {
		return model.Off, nil
	} else {
		return model.On, nil
	}
}

func (p *pinImpl) Toggle() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	v, err := p.getState()
	if err != nil {
		return err
	}

	if v == model.On {
		return p.setState(model.Off)
	} else {
		return p.setState(model.On)
	}
}

func (p *pinImpl) getBaseDir() string {
	return p.baseDir
}

func (p *pinImpl) getValueFile() *os.File {
	return p.valueFile
}

func (p *pinImpl) getModeFile() *os.File {
	return p.modeFile
}*/
