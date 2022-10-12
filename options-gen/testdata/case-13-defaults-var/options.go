package testcase

import (
	"net/http"
	"time"
)

type Options struct {
	name        string        `validate:"required"`
	timeout     time.Duration `validate:"min=100ms,max=30s"`
	maxAttempts int           `validate:"min=1,max=10"`
	httpClient  *http.Client  `validate:"gt=0"`
}

var defaultOptions = Options{
	name:        "some-name",
	timeout:     3 * time.Second,
	maxAttempts: 4,
	httpClient:  &http.Client{Transport: &http.Transport{MaxConnsPerHost: 10}},
}
