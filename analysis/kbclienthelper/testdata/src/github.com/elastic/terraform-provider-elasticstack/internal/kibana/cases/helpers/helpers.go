package helpers

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NewSDKKibanaClient(d *schema.ResourceData, meta any) (*clients.APIClient, error) {
	return clients.NewKibanaAPIClientFromSDKResource(d, meta)
}

func NewFrameworkKibanaClient(ctx context.Context, connection any, defaultClient any) (*clients.APIClient, error) {
	return clients.MaybeNewKibanaAPIClientFromFrameworkResource(ctx, connection, defaultClient)
}

func NotAllowedKibanaWrapper(_ *schema.ResourceData, _ any) (*clients.APIClient, error) {
	return &clients.APIClient{}, nil
}
