package testcase

type Options[T any] struct {
	ch1 chan T   `option:"mandatory" validate:"required"`
	ch2 <-chan T `option:"mandatory"`
	ch3 chan T
	ch4 <-chan T
}
