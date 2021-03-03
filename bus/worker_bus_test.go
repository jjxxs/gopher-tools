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
func TestGetNamedMultiWorkerBus(t *testing.T) {
	names := []string{"t1", "t2", "t3"}
	bs := []Bus{nil, nil, nil}
	for i, name := range names {
		bs[i] = GetNamedWorkerBus(name)
	}

	// same name should always return same reference
	for i, name := range names {
		b := GetNamedWorkerBus(name)
		ptr1 := reflect.ValueOf(b).Pointer()
		ptr2 := reflect.ValueOf(bs[i]).Pointer()
		if ptr1 != ptr2 {
			t.Fail()
		}
	}
}

func TestGetMultiWorkerBusShouldReturnSingleton(t *testing.T) {
	bs := []Bus{nil, nil, nil}
	for i := 0; i < 3; i++ {
		bs[i] = GetWorkerBus()
	}

	if bs[0] == nil {
		t.Fail()
	} else if bs[0] != bs[1] {
		t.Fail()
	} else if bs[1] != bs[2] {
		t.Fail()
	}
}

type workerTestSub struct {
	c chan struct{}
}

func (s *workerTestSub) HandleMessage(_ Message) {
	s.c <- struct{}{}
}

func TestMultiWorkerBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	b := NewWorkerBus(100)
	c := make(chan struct{}, 100)
	s := &workerTestSub{c}
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

func TestMultiWorkerBusReceiverShouldNotReceivePublishMessageAfterUnsubscribe(t *testing.T) {
	b := NewWorkerBus(100)
	c := make(chan struct{}, 100)
	s := &workerTestSub{c}
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
func BenchmarkMultiWorkerBusPublishPrimitive__1_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N)
	s := &workerTestSub{c}
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

func BenchmarkMultiWorkerBusPublishPrimitive__1000_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub{c}
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
		bu.Publish(i)
	}
	wg.Wait()
}

// the msg to be published
var workerBenchObj = struct {
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

func BenchmarkMultiWorkerBusPublishStructByValue__1_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N)
	s := &workerTestSub{c}
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
		bu.Publish(workerBenchObj)
	}
	wg.Wait()
}

func BenchmarkMultiWorkerBusPublishStructByValue__1000_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub{c}
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
		bu.Publish(workerBenchObj)
	}
	wg.Wait()
}

func BenchmarkMultiWorkerBusPublishReference__1_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N)
	s := &workerTestSub{c}
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
		bu.Publish(&workerBenchObj)
	}
	wg.Wait()
}

func BenchmarkMultiWorkerBusPublishReference__1000_Subs(b *testing.B) {
	bu := NewWorkerBus(b.N)
	c := make(chan struct{}, b.N*1000)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub{c}
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
		bu.Publish(&workerBenchObj)
	}
	wg.Wait()
}
