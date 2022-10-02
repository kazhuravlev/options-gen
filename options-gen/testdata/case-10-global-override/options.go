package testcase

type Options struct {
	amount int `option:"mandatory"`
	age    int `option:"mandatory" validate:"child"` // Unknown tag!
}
