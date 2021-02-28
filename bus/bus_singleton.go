package bus

import "sync"

// Holds 'named' Bus-singletons
var buss = sync.Map{}

// GetNamedBus - Provides thread-safe access to a Bus with a specified
// name. Repeated calls with the same name always return the same Bus.
func GetNamedBus(name string) Bus {
	var bus interface{}

	if bus, _ = buss.Load(name); bus == nil {
		bus = NewBus()
		buss.Store(name, bus)
	}

	return bus.(Bus)
}

var (
	once     = &sync.Once{}
	bus  Bus = nil
)

// GetBus - Provides thread-safe access to a Bus singleton which is
// initialized the first time GetBus is called. Repeated calls will
// return the same instance.
func GetBus() Bus {
	once.Do(func() {
		bus = NewBus()
	})
	return bus
}
