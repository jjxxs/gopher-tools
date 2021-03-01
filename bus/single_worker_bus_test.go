package bus

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

/**
 * Tests
 */
func TestGetNamedSingleWorkerBus(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	buses := []Bus{nil, nil, nil}
	for i, name := range names {
		buses[i] = GetNamedSingleWorkerBus(name)
	}

	// same name should always return same reference
	for i, name := range names {
		bus := GetNamedSingleWorkerBus(name)
		ptr1 := reflect.ValueOf(bus).Pointer()
		ptr2 := reflect.ValueOf(buses[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestGetSingleWorkerBusShouldReturnSingleton(t *testing.T) {
	buses := []Bus{nil, nil, nil}
	for i := 0; i < 3; i++ {
		buses[i] = GetSingleWorkerBus()
	}

	if buses[0] == nil {
		t.Fail()
	} else if buses[0] != buses[1] {
		t.Fail()
	} else if buses[1] != buses[2] {
		t.Fail()
	}
}

type singleWorkerTestSubscriber struct {
	c chan struct{}
}

func (s *singleWorkerTestSubscriber) HandleMessage(_ Message) {
	s.c <- struct{}{}
}

func TestSingleWorkerBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, 100)
	bus.Subscribe(&singleWorkerTestSubscriber{c})
	for i := 0; i < 100; i++ {
		bus.Publish(i)
	}
	timer := time.NewTimer(100 * time.Millisecond)
	for count := 0; count < 100; {
		select {
		case <-timer.C:
			t.Fail()
		case <-c:
			count++
		}
	}
}

func TestSingleWorkerBusReceiverShouldNotReceivePublishMessageAfterUnsubscribe(t *testing.T) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, 100)
	s := &singleWorkerTestSubscriber{c}
	bus.Subscribe(s)
	for i := 0; i < 100; i++ {
		bus.Publish(i)
	}
	timer := time.NewTimer(100 * time.Millisecond)
	for count := 0; count < 100; {
		select {
		case <-timer.C:
			t.Fail()
		case <-c:
			count++
		}
	}
	bus.Unsubscribe(s)
	for i := 0; i < 100; i++ {
		bus.Publish(i) // these should not be delivered
	}
	select {
	case <-time.After(100 * time.Millisecond):
		break
	case <-c:
		t.Fail()
	}
}

/**
 * Benchmarks
 */
func BenchmarkSingleWorkerBusPublishPrimitive__1_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&singleWorkerTestSubscriber{c})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(i)
	}
	wg.Wait()
}

func BenchmarkSingleWorkerBusPublishPrimitive__1000_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&singleWorkerTestSubscriber{c})
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N*1000; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(i)
	}
	wg.Wait()
}

// the msg to be published
var singleWorkerBenchObj = struct {
	i int
	j int
	s string
	v []int
}{
	i: 1,
	j: 2,
	s: "myArg1",
	v: []int{1, 2, 3},
}

func BenchmarkSingleWorkerBusPublishStructByValue__1_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&singleWorkerTestSubscriber{c})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(singleWorkerBenchObj)
	}
	wg.Wait()
}

func BenchmarkSingleWorkerBusPublishStructByValue__1000_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&singleWorkerTestSubscriber{c})
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N*1000; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(singleWorkerBenchObj)
	}
	wg.Wait()
}

func BenchmarkSingleWorkerBusPublishReference__1_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&singleWorkerTestSubscriber{c})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(&singleWorkerBenchObj)
	}
	wg.Wait()
}

func BenchmarkSingleWorkerBusPublishReference__1000_Subs(b *testing.B) {
	bus := NewSingleWorkerBus()
	c := make(chan struct{}, b.N*1000)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&singleWorkerTestSubscriber{c})
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for count := 0; count < b.N*1000; count++ {
			<-c
		}
		wg.Done()
	}()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bus.Publish(&singleWorkerBenchObj)
	}
	wg.Wait()
}
