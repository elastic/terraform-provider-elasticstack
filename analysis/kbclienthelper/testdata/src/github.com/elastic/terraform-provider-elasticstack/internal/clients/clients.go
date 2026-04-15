package clients

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type APIClient struct{}

func MaybeNewKibanaAPIClientFromFrameworkResource(_ context.Context, _ any, _ any) (*APIClient, error) {
	return &APIClient{}, nil
}

func NewKibanaAPIClientFromSDKResource(_ *schema.ResourceData, _ any) (*APIClient, error) {
	return &APIClient{}, nil
}

func (c *APIClient) GetKibanaClient() error   { return nil }
func (c *APIClient) GetFleetClient() error    { return nil }
func (c *APIClient) GetKibanaOapiClient() error { return nil }
func (c *APIClient) GetSloClient() error      { return nil }

func (c *APIClient) ID(_ context.Context, id string) (string, error) { return id, nil }
