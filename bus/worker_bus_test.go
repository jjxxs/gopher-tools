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
	bs := []Bus[any]{nil, nil, nil}
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
	bs := []Bus[any]{nil, nil, nil}
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

type workerTestSub[E any] struct {
	c chan E
}

func (s *workerTestSub[E]) HandleMessage(msg E) {
	s.c <- msg
}

func TestMultiWorkerBusReceiverShouldReceivePublishedMessages(t *testing.T) {
	b := NewWorkerBus[int](100)
	c := make(chan int, 100)
	s := &workerTestSub[int]{c}
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
	b := NewWorkerBus[int](100)
	c := make(chan int, 100)
	s := &workerTestSub[int]{c}
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
 * PublishTimeout edge cases (regression coverage for the Ticker→Timer fix).
 */
func TestPublishTimeoutSucceedsWhenSpaceAvailable(t *testing.T) {
	b := NewWorkerBus[int](10)
	if !b.PublishTimeout(42, 100*time.Millisecond) {
		t.Fatal("expected PublishTimeout to return true when queue has space")
	}
}

func TestPublishTimeoutReturnsFalseWhenQueueFull(t *testing.T) {
	b := NewWorkerBus[int](1)
	block := make(chan struct{})
	defer close(block)
	b.Subscribe(func(int) { <-block })

	// Pump until the queue saturates. With the subscriber blocked, the
	// 1-deep input queue and 1-deep sub queue will fill within a few sends.
	saturated := false
	for i := 0; i < 100; i++ {
		if !b.PublishTimeout(i, 10*time.Millisecond) {
			saturated = true
			break
		}
	}
	if !saturated {
		t.Fatal("PublishTimeout never returned false despite blocked subscriber")
	}

	// On a saturated queue, the timeout must actually elapse.
	start := time.Now()
	if b.PublishTimeout(99, 50*time.Millisecond) {
		t.Fatal("expected PublishTimeout to return false on a saturated queue")
	}
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("PublishTimeout returned too quickly on saturated queue: %v", elapsed)
	}
}

func TestPublishTimeoutNonPositiveReturnsFalseImmediately(t *testing.T) {
	b := NewWorkerBus[int](10)
	for _, timeout := range []time.Duration{0, -time.Second} {
		start := time.Now()
		ok := b.PublishTimeout(42, timeout)
		elapsed := time.Since(start)
		if ok {
			t.Errorf("timeout=%v: expected PublishTimeout to return false", timeout)
		}
		if elapsed > 10*time.Millisecond {
			t.Errorf("timeout=%v: returned too slowly (%v) — non-positive must short-circuit", timeout, elapsed)
		}
	}
}

/**
 * Benchmarks
 */
func BenchmarkMultiWorkerBusPublishPrimitive__1_Subs(b *testing.B) {
	bu := NewWorkerBus[int](b.N)
	c := make(chan int, b.N)
	s := &workerTestSub[int]{c}
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
	bu := NewWorkerBus[int](b.N)
	c := make(chan int, b.N)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub[int]{c}
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
type WorkerBenchObj = struct {
	i int
	j int
	s string
	v []int
}

var workerBenchObj = WorkerBenchObj{
	i: 1,
	j: 2,
	s: "myArg1",
	v: []int{1, 2, 3},
}

func BenchmarkMultiWorkerBusPublishStructByValue__1_Subs(b *testing.B) {
	bu := NewWorkerBus[WorkerBenchObj](b.N)
	c := make(chan WorkerBenchObj, b.N)
	s := &workerTestSub[WorkerBenchObj]{c}
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
	bu := NewWorkerBus[WorkerBenchObj](b.N)
	c := make(chan WorkerBenchObj, b.N)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub[WorkerBenchObj]{c}
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
	bu := NewWorkerBus[*WorkerBenchObj](b.N)
	c := make(chan *WorkerBenchObj, b.N)
	s := &workerTestSub[*WorkerBenchObj]{c}
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
	bu := NewWorkerBus[*WorkerBenchObj](b.N)
	c := make(chan *WorkerBenchObj, b.N*1000)
	for i := 0; i < 1000; i++ {
		s := &workerTestSub[*WorkerBenchObj]{c}
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
