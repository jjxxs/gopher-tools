package bus

import (
	"reflect"
	"testing"
	"time"
)

/**
 * Tests
 */

func TestGetNamedBus(t *testing.T) {
	myNames := []string{"", "t", "te", "tes", "test"}
	myBuses := []Bus{nil, nil, nil, nil, nil}
	for i, name := range myNames {
		myBuses[i] = GetNamedBus(name)
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
		b := GetNamedBus(name)
		ptr1 := reflect.ValueOf(b).Pointer()
		ptr2 := reflect.ValueOf(myBuses[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestBusSubscribePublishUnsubscribe(t *testing.T) {
	const iterations = 100
	const beforeSubscription = 1
	const whenSubscription = 2
	const afterSubscription = 3

	sut := GetNamedBus("busUnderTest")
	cnt := 0
	receiver := func(arg interface{}) {
		if i, ok := arg.(int); ok {
			// messages that were published before the subscription should not be delivered
			if i == beforeSubscription {
				t.Fail()
			}

			if i == whenSubscription {
				cnt++
			}

			// messages that were published after the subscription should not be delivered
			if i == afterSubscription {
				t.Fail()
			}
		} else {
			t.Fail()
		}
	}

	// publishing to a bus that has no subscribers does nothing
	for i := 0; i < iterations; i++ {
		sut.Publish(beforeSubscription)
	}
	// since messages are delivered asynchronously, wait a little
	time.Sleep(1 * time.Second)

	// messages published should be visible to subscribers
	sut.Subscribe(receiver)
	for i := 0; i < iterations; i++ {
		sut.Publish(whenSubscription)
	}
	time.Sleep(1 * time.Second)
	if cnt != iterations {
		t.Fail()
	}

	// no messages should be received after unsubscribing
	sut.Unsubscribe(receiver)
	for i := 0; i < iterations; i++ {
		sut.Publish(afterSubscription)
	}
	time.Sleep(1 * time.Second)
}

/**
 * Benchmarks
 */

func BenchmarkPublishPrimitiveArgsOneSubscriber(b *testing.B) {
	bus := NewBus()

	fun1Ctr := 0
	fun1 := func(i interface{}) {
		fun1Ctr++
	}

	bus.Subscribe(fun1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Publish(i)
	}
}

// the type to be published
type myType struct {
	i int
	j int
	s string
	v []int
}

// the object to be published
var myArg = myType{
	i: 1,
	j: 2,
	s: "myArg1",
	v: []int{1, 2, 3},
}

func BenchmarkPublishStructByValueOneSubscriber(b *testing.B) {
	const subscribers = 1
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(myArg)
	}
}

func BenchmarkPublishStructByValueOneHundredSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(myArg)
	}
}

func BenchmarkPublishStructByValueOneThousandSubscriber(b *testing.B) {
	const subscribers = 1000
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(myArg)
	}
}

func BenchmarkPublishReferenceOneSubscriber(b *testing.B) {
	const subscribers = 1
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(&myArg)
	}
}

func BenchmarkPublishReferenceOneHundredSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(&myArg)
	}
}

func BenchmarkPublishReferenceOneThousandSubscriber(b *testing.B) {
	const subscribers = 100
	sut := NewBus()

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
		sut.Subscribe(subs[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sut.Publish(&myArg)
	}
}
