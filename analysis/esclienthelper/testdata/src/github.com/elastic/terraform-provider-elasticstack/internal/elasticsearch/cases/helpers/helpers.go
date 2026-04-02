package helpers

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NewSDKClient(d *schema.ResourceData, meta any) (*clients.APIClient, error) {
	return clients.NewAPIClientFromSDKResource(d, meta)
}

func NewFrameworkClient(ctx context.Context, connection any, defaultClient any) (*clients.APIClient, error) {
	return clients.MaybeNewAPIClientFromFrameworkResource(ctx, connection, defaultClient)
}

func NotAllowedSDKWrapper(_ *schema.ResourceData, _ any) (*clients.APIClient, error) {
	return &clients.APIClient{}, nil
}
