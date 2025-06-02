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

func (c *Client) Options() *Options {
	return c.opts
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *StatusCmd {
	return &StatusCmd{}
}

type StatusCmd struct{}

func (s *StatusCmd) Err() error { return nil }

func (c *Client) Close() error { return nil }
