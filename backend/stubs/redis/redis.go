package redis

import (
	"context"
	"time"
)

type Options struct {
	Addr string
}

type Client struct {
	opts *Options
}

func NewClient(opts *Options) *Client {
	return &Client{opts: opts}
}

func (c *Client) Close() error { return nil }

func (c *Client) Options() *Options { return c.opts }

type StatusCmd struct{ err error }

func (c *Client) Set(ctx context.Context, key string, value interface{}, exp time.Duration) *StatusCmd {
	return &StatusCmd{}
}

func (s *StatusCmd) Err() error { return s.err }
