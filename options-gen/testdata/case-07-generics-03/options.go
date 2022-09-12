package testcase

type Options[A comparable, B, C any, D int | string, E []A, F, G []any] struct {
	a A
	b B
	c C
	d D
	e E
	f F
	g G
}
