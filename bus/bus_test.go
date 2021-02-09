package bus

import (
	"reflect"
	"testing"
	"time"
)

/**
 * Tests
 */

func TestGetNamedBusShouldReturnSameReferenceForSameName(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	buses := []Bus{nil, nil, nil}
	for i, name := range names {
		buses[i] = GetNamedBus(name)
	}

	// same name should always return same reference
	for i, name := range names {
		bus := GetNamedBus(name)
		ptr1 := reflect.ValueOf(bus).Pointer()
		ptr2 := reflect.ValueOf(buses[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	bus := NewBus()
	count := 0
	bus.Subscribe(func(msg Message) {
		count++
	})
	for i := 0; i < 100; i++ {
		bus.Publish(1)
	}
	time.Sleep(100 * time.Millisecond) // message delivered asynchronously, wait a little
	if count != 100 {
		t.Fail()
	}
}

func TestBusReceiverShouldNotReceivePublishMessageAfterUnsubscribe(t *testing.T) {
	bus := NewBus()
	count := 0
	receiver := func(msg Message) {
		count++
	}
	bus.Subscribe(receiver)
	for i := 0; i < 100; i++ {
		bus.Publish(1)
	}
	time.Sleep(100 * time.Millisecond) // message delivered asynchronously, wait a little
	bus.Unsubscribe(receiver)
	for i := 0; i < 100; i++ {
		bus.Publish(1) // these should not be delivered
	}
	time.Sleep(100 * time.Millisecond) // message delivered asynchronously, wait a little
	if count != 100 {
		t.Fail()
	}
}

/**
 * Benchmarks
 */

func BenchmarkPublishPrimitive__1_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	bus.Subscribe(func(msg Message) {
		count++
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(i)
	}
}

func BenchmarkPublishPrimitive__1000_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	for i := 0; i < 1000; i++ {
		bus.Subscribe(func(msg Message) {
			count++
		})
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(i)
	}
}

// the type to be published
type benchType struct {
	i int
	j int
	s string
	v []int
}

// the object to be published
var benchObj = benchType{
	i: 1,
	j: 2,
	s: "myArg1",
	v: []int{1, 2, 3},
}

func BenchmarkPublishStructByValue__1_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	bus.Subscribe(func(msg Message) {
		count++
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(benchObj)
	}
}

func BenchmarkPublishStructByValue__1000_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	for i := 0; i < 1000; i++ {
		bus.Subscribe(func(msg Message) {
			count++
		})
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(benchObj)
	}
}

func BenchmarkPublishReference__1_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	bus.Subscribe(func(msg Message) {
		count++
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(&benchObj)
	}
}

func BenchmarkPublishReference__1000_Subs(b *testing.B) {
	bus := NewBus()
	count := 0
	for i := 0; i < 1000; i++ {
		bus.Subscribe(func(msg Message) {
			count++
		})
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(&benchObj)
	}
}
