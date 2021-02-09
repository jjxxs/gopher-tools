package bus

import (
	"reflect"
	"testing"
)

func TestGetNamedBusShouldReturnSingletonForName(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	buses := []Bus{nil, nil, nil}
	for i, name := range names {
		buses[i] = GetNamedBus(name)
	}

	// different name should return different reference, same name returns same reference
	for i, bus1 := range buses {
		for j, bus2 := range buses {
			ptr1 := reflect.ValueOf(bus1).Pointer()
			ptr2 := reflect.ValueOf(bus2).Pointer()
			if i != j {
				if ptr1 == ptr2 {
					t.Fail()
				}
			} else {
				if ptr1 != ptr2 {
					t.Fail()
				}
			}
		}
	}
}
