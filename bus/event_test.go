package bus

import (
	"reflect"
	"testing"
	"time"
)

const event1 Event = "event1"
const event2 Event = "event2"

func TestGetNamedEventBus(t *testing.T) {
	myNames := []string{"", "t", "te", "tes", "test"}
	myBuses := []EventBus{nil, nil, nil, nil, nil}
	for i, name := range myNames {
		myBuses[i] = GetNamedEventBus(name)
	}

	// busses with different names should return different references
	for i, bus1 := range myBuses {
		for j, bus2 := range myBuses {
			if i != j {
				ptr1 := reflect.ValueOf(bus1).Pointer()
				ptr2 := reflect.ValueOf(bus2).Pointer()
				if ptr1 == ptr2 {
					t.Fail()
				}
			}
		}
	}

	// busses with the same name should always return the same reference
	for i, name := range myNames {
		b := GetNamedEventBus(name)
		ptr1 := reflect.ValueOf(b).Pointer()
		ptr2 := reflect.ValueOf(myBuses[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestEventBusSubscribePublishUnsubscribe(t *testing.T) {
	const iterations = 100
	const arg = 42
	sut := GetNamedEventBus("busUnderTest")

	event1Count := 0
	event1Receiver := func(arg interface{}) {
		if _, ok := arg.(int); ok {
			event1Count++
		} else {
			t.Fail()
		}
	}

	event2Count := 0
	event2Receiver := func(arg interface{}) {
		if _, ok := arg.(int); ok {
			event2Count++
		}
	}

	// subscribe event2 but not event1
	sut.Subscribe(event2, event2Receiver)
	for i := 0; i < iterations; i++ {
		// publish to event1 should do nothing, there are no subscribers
		sut.Publish(event1, arg)
		// publish to event2 should increment event2Count
		sut.Publish(event2, arg)
	}
	time.Sleep(1 * time.Second)
	if event1Count != 0 {
		t.Fail()
	}
	if event2Count != iterations {
		t.Fail()
	}

	cancelSubscription := sut.Subscribe(event1, event1Receiver)
	for i := 0; i < iterations; i++ {
		// publish to event1 and event2 should be delivered to subscribers
		sut.Publish(event1, arg)
		sut.Publish(event2, arg)
	}
	time.Sleep(1 * time.Second)
	if event1Count != iterations {
		t.Fail()
	}
	if event2Count != iterations*2 {
		t.Fail()
	}

	// no messages should be received after unsubscribing
	cancelSubscription()
	for i := 0; i < iterations; i++ {
		sut.Publish(event1, arg)
		sut.Publish(event2, arg)
	}
	time.Sleep(1 * time.Second)
	if event1Count != iterations {
		t.Fail()
	}
	if event2Count != iterations*3 {
		t.Fail()
	}
}

/**
 * Benchmarks
 */

func BenchmarkEventBusPublishPrimitiveArgsOneSubscriber(b *testing.B) {
	eventBus := NewEventBus()

	fun1Ctr := 0
	fun1 := func(i interface{}) {
		fun1Ctr++
	}

	eventBus.Subscribe(event1, fun1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		eventBus.Publish(event1, i)
	}
}

type eventBusArgumentType struct {
	i int
	j int
	s string
	v []int
}

var eventBusArgument = myType{
	i: 1,
	j: 2,
	s: "myArg1",
	v: []int{1, 2, 3},
}

func BenchmarkEventBusPublishStructByValueOneSubscriber(b *testing.B) {
	const subscribers = 1
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, myArg)
	}
}

func BenchmarkEventBusPublishStructByValueOneHundredSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, myArg)
	}
}

func BenchmarkEventBusPublishStructByValueOneThousandSubscriber(b *testing.B) {
	const subscribers = 1000
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, myArg)
	}
}

func BenchmarkEventBusPublishReferenceOneSubscriber(b *testing.B) {
	const subscribers = 1
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, &myArg)
	}
}

func BenchmarkEventBusPublishReferenceOneHundredSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, &myArg)
	}
}

func BenchmarkEventBusPublishReferenceOneThousandSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewEventBus()

	// the subscribers/receivers - they just increment an integer
	subs := make([]func(arg interface{}), subscribers)
	for i := 0; i < subscribers; i++ {
		j := 0
		subs[i] = func(arg interface{}) {
			j++
		}
	}

	// subscribe
	for i := 0; i < subscribers; i++ {
		sut.Subscribe(event1, subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(event1, &myArg)
	}
}
