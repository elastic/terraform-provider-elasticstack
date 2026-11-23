package kbapi

import (
	"context"
	"net/http"
)

type Client struct{}

func (c *Client) Orphan(_ context.Context) (*http.Response, error) {
	return nil, nil
}
