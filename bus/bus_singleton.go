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
