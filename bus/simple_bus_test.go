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
func TestGetNamedSimpleBus(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	buses := []Bus{nil, nil, nil}
	for i, name := range names {
		buses[i] = GetNamedSimpleBus(name)
	}

	// same name should always return same reference
	for i, name := range names {
		bus := GetNamedSimpleBus(name)
		ptr1 := reflect.ValueOf(bus).Pointer()
		ptr2 := reflect.ValueOf(buses[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestGetSimpleBusShouldReturnSingleton(t *testing.T) {
	buses := []Bus{nil, nil, nil}
	for i := 0; i < 3; i++ {
		buses[i] = GetSimpleBus()
	}

	if buses[0] == nil {
		t.Fail()
	} else if buses[0] != buses[1] {
		t.Fail()
	} else if buses[1] != buses[2] {
		t.Fail()
	}
}

type simpleTestSubscriber struct {
	c chan struct{}
}

func (s *simpleTestSubscriber) HandleMessage(_ Message) {
	s.c <- struct{}{}
}

func TestSimpleBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	bus := NewSimpleBus()
	c := make(chan struct{}, 100)
	bus.Subscribe(&simpleTestSubscriber{c})
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

func TestSimpleBusReceiverShouldNotReceivePublishMessageAfterUnsubscribe(t *testing.T) {
	bus := NewSimpleBus()
	c := make(chan struct{}, 100)
	s := &simpleTestSubscriber{c}
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
func BenchmarkSimpleBusPublishPrimitive__1_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&simpleTestSubscriber{c})
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

func BenchmarkSimpleBusPublishPrimitive__1000_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&simpleTestSubscriber{c})
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
var simpleBenchObj = struct {
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

func BenchmarkSimpleBusPublishStructByValue__1_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&simpleTestSubscriber{c})
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
		bus.Publish(simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkSimpleBusPublishStructByValue__1000_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&simpleTestSubscriber{c})
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
		bus.Publish(simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkSimpleBusPublishReference__1_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N)
	bus.Subscribe(&simpleTestSubscriber{c})
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
		bus.Publish(&simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkSimpleBusPublishReference__1000_Subs(b *testing.B) {
	bus := NewSimpleBus()
	c := make(chan struct{}, b.N*1000)
	for i := 0; i < 1000; i++ {
		bus.Subscribe(&simpleTestSubscriber{c})
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
		bus.Publish(&simpleBenchObj)
	}
	wg.Wait()
}
