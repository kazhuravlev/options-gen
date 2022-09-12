package testcase

type Options[T string] struct {
	requiredKey T `option:"mandatory" validate:"required"`
	key         T `option:"mandatory"`
	optKey      T
}
