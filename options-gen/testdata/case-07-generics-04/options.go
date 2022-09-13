package testcase

type SomeData[X any] struct {
	value X
}

type Options[T comparable] struct {
	d1 SomeData[T]  `option:"mandatory"`
	d2 *SomeData[T] `option:"mandatory"`

	d3 SomeData[T]
	d4 *SomeData[T]
}
