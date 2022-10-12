package testcase

import "time"

type Options struct {
	pingPeriod  time.Duration `my-default-tag:"3s" validate:"min=100ms,max=30s"`
	name        string        `my-default-tag:"unknown" validate:"required"`
	maxAttempts int           `my-default-tag:"10" validate:"min=1,max=10"`
	eps         float32       `my-default-tag:"0.0001" validate:"gt=0"`
}
