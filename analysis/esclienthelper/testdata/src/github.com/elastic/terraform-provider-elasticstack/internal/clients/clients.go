package clients

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type APIClient struct{}

func MaybeNewAPIClientFromFrameworkResource(_ context.Context, _ any, _ any) (*APIClient, error) {
	return &APIClient{}, nil
}

func NewAPIClientFromSDKResource(_ *schema.ResourceData, _ any) (*APIClient, error) {
	return &APIClient{}, nil
}

func (c *APIClient) GetESClient() error { return nil }

func (c *APIClient) ID(_ context.Context, id string) (string, error) { return id, nil }

func (c *APIClient) ServerVersion(_ context.Context) (string, error) { return "8.0.0", nil }
