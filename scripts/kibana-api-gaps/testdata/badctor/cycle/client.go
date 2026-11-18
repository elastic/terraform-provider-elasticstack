package cycle

import (
	"context"
	"net/http"
)

type Client struct{}

func (c *Client) Op(_ context.Context) (*http.Response, error) {
	_, err := NewOpRequest("")
	return nil, err
}
