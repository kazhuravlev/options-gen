package pkg

type Options[T string] struct {
	requiredKey T `option:"mandatory" validate:"required"`
	key         T `option:"mandatory"`
	OptKey      T
}
