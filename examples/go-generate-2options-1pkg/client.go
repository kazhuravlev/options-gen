package gogenerate

import (
	"fmt"
)

type Client struct {
	opts1 Options1
	opts2 Options2
}

func New(opts1 Options1, opts2 Options2) (*Client, error) {
	if err := opts1.Validate(); err != nil {
		return nil, fmt.Errorf("bad configuration opts1: %w", err)
	}

	if err := opts2.Validate(); err != nil {
		return nil, fmt.Errorf("bad configuration opts2: %w", err)
	}

	return &Client{
		opts1: opts1,
		opts2: opts2,
	}, nil
}
