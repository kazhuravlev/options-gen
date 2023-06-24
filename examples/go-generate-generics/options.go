package gogenerate

//go:generate options-gen -from-struct=Options -defaults-from=func
type Options[K comparable, V any] struct {
	defaultVal V
}

func getDefaultOptions[K comparable, V any]() Options[K, V] {
	var val V

	return Options[K, V]{
		defaultVal: val,
	}
}
