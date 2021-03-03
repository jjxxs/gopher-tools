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
func TestGetNamedBus(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	bs := []Bus{nil, nil, nil}
	for i, name := range names {
		bs[i] = GetNamedBus(name)
	}

	// same name should always return same reference
	for i, name := range names {
		b := GetNamedBus(name)
		ptr1 := reflect.ValueOf(b).Pointer()
		ptr2 := reflect.ValueOf(bs[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestGetBusShouldReturnSingleton(t *testing.T) {
	bs := []Bus{nil, nil, nil}
	for i := 0; i < 3; i++ {
		bs[i] = GetBus()
	}

	if bs[0] == nil {
		t.Fail()
	} else if bs[0] != bs[1] {
		t.Fail()
	} else if bs[1] != bs[2] {
		t.Fail()
	}
}

type busTestSub struct {
	c chan struct{}
}

func (s *busTestSub) HandleMessage(_ Message) {
	s.c <- struct{}{}
}

func TestBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	b := NewBus()
	c := make(chan struct{}, 100)
	s := &busTestSub{c}
	b.Subscribe(s.HandleMessage)
	for i := 0; i < 100; i++ {
		b.Publish(i)
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

func TestBusReceiverShouldNotReceivePublishMessageAfterUnsubscribe(t *testing.T) {
	b := NewBus()
	c := make(chan struct{}, 100)
	s := &busTestSub{c}
	unsubscribe := b.Subscribe(s.HandleMessage)
	for i := 0; i < 100; i++ {
		b.Publish(i)
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
	unsubscribe()
	for i := 0; i < 100; i++ {
		b.Publish(i) // these should not be delivered
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
func BenchmarkBusPublishPrimitive__1_Subs(b *testing.B) {
	bu := NewBus()
	c := make(chan struct{}, b.N)
	s := busTestSub{c}
	bu.Subscribe(s.HandleMessage)
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
		bu.Publish(i)
	}
	wg.Wait()
}

func BenchmarkBusPublishPrimitive__1000_Subs(b *testing.B) {
	bus := NewBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		s := &busTestSub{c}
		bus.Subscribe(s.HandleMessage)
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

func BenchmarkBusPublishStructByValue__1_Subs(b *testing.B) {
	bu := NewBus()
	c := make(chan struct{}, b.N)
	s := &busTestSub{c}
	bu.Subscribe(s.HandleMessage)
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
		bu.Publish(simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkBusPublishStructByValue__1000_Subs(b *testing.B) {
	bu := NewBus()
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		s := &busTestSub{c}
		bu.Subscribe(s.HandleMessage)
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
		bu.Publish(simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkBusPublishReference__1_Subs(b *testing.B) {
	bu := NewBus()
	c := make(chan struct{}, b.N)
	s := &busTestSub{c}
	bu.Subscribe(s.HandleMessage)
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
		bu.Publish(&simpleBenchObj)
	}
	wg.Wait()
}

func BenchmarkBusPublishReference__1000_Subs(b *testing.B) {
	bu := NewBus()
	c := make(chan struct{}, b.N*1000)
	for i := 0; i < 1000; i++ {
		s := &busTestSub{c}
		bu.Subscribe(s.HandleMessage)
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
		bu.Publish(&simpleBenchObj)
	}
	wg.Wait()
}
