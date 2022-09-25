package main

import "testing"

func TestMultiFile(t *testing.T) {
	_ = NewOptions(nil, "s3")
	_ = NewConfig()
	_ = NewParams("d285afedd3e14589ddfe2d6bf4319bfd")
}
