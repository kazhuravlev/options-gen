package gogenerate

import (
	"fmt"
)

type Client[K comparable, V any] struct {
	opts Options[K, V]

	m map[K]V
}

func New[K comparable, V any](opts Options[K, V]) (*Client[K, V], error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("bad configuration: %w", err)
	}

	return &Client[K, V]{
		opts: opts,
		m:    map[K]V{},
	}, nil
}

func (c *Client[K, V]) Set(key K, val V) {
	c.m[key] = val
}

func (c *Client[K, V]) Get(key K) V {
	val, ok := c.m[key]
	if !ok {
		return c.opts.defaultVal
	}

	return val
}
