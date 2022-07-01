package cli

import (
	"context"
	"github.com/pkg/errors"
)

type Client struct {
	opts Options
}

func New(opts Options) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "bad configuration")
	}

	return &Client{opts: opts}, nil
}

func (c *Client) SendRequest(ctx context.Context) error {
	_, err := c.opts.httpClient.Get("http://localhost:8000/hello?token=" + c.opts.token)
	return err
}
