package bus

import (
	"testing"
)

func BenchmarkBusNoArgs(b *testing.B) {
	bus := NewBus()

	fun1Ctr := 0
	fun1 := func() {
		fun1Ctr++
	}

	fun2Ctr := 0
	fun2 := func() {
		fun2Ctr++
	}

	bus.Subscribe(fun1)
	bus.Subscribe(fun2)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Publish()
	}
}

func BenchmarkBusPrimitiveArgs(b *testing.B) {
	bus := NewBus()

	fun1Ctr := 0
	fun1 := func(i, j int) {
		fun1Ctr++
	}

	fun2Ctr := 0
	fun2 := func(i, j int) {
		fun2Ctr++
	}

	bus.Subscribe(fun1)
	bus.Subscribe(fun2)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Publish(i, i)
	}
}

func BenchmarkBusComplexArgsByValue(b *testing.B) {
	bus := NewBus()

	type myType struct {
		i int
		j int
		s string
		v []int
	}

	myArg1 := myType{
		i: 1,
		j: 2,
		s: "myArg1",
		v: []int{1, 2, 3},
	}

	myArg2 := myType{
		i: -1,
		j: -2,
		s: "myArg2",
		v: []int{1, 2, 3},
	}

	fun1Ctr := 0
	fun1 := func(arg1, arg2 myType) {
		fun1Ctr++
	}

	fun2Ctr := 0
	fun2 := func(arg1, arg2 myType) {
		fun2Ctr++
	}

	bus.Subscribe(fun1)
	bus.Subscribe(fun2)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Publish(myArg1, myArg2)
	}
}

func BenchmarkBusComplexArgsByReference(b *testing.B) {
	bus := NewBus()

	type myType struct {
		i int
		j int
		s string
		v []int
	}

	myArg1 := myType{
		i: 1,
		j: 2,
		s: "myArg1",
		v: []int{1, 2, 3},
	}

	myArg2 := myType{
		i: -1,
		j: -2,
		s: "myArg2",
		v: []int{1, 2, 3},
	}

	fun1Ctr := 0
	fun1 := func(arg1, arg2 *myType) {
		fun1Ctr++
	}

	fun2Ctr := 0
	fun2 := func(arg1, arg2 *myType) {
		fun2Ctr++
	}

	bus.Subscribe(fun1)
	bus.Subscribe(fun2)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bus.Publish(&myArg1, &myArg2)
	}
}
