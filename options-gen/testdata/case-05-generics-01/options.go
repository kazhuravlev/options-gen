package testcase

type Options[T string] struct {
	RequiredKey T `option:"mandatory" validate:"required"`
	Key         T `option:"mandatory"`
	OptKey      T
}
