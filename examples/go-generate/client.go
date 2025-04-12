package gogenerate

import (
	"context"
	"fmt"
)

type Client struct {
	opts Options
}

func New(opts Options) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("bad configuration: %w", err)
	}

	return &Client{opts: opts}, nil
}

func (c *Client) SendRequest(_ context.Context) error {
	_, err := c.opts.httpClient.Get("http://localhost:8000/hello?token=" + c.opts.token) //nolint:bodyclose,noctx

	return err
}
