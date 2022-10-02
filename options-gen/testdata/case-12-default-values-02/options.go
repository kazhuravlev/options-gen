package testcase

import "time"

type Options struct {
	pingPeriod  time.Duration `default:"3s" validate:"min=100ms,max=30s"`
	name        string        `default:"unknown" validate:"required"`
	maxAttempts int           `default:"10" validate:"min=1,max=10"`
	eps         float32       `default:"0.0001" validate:"gt=0"`
}
