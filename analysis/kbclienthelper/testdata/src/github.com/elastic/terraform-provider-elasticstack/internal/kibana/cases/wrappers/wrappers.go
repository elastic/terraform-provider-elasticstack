package wrappers

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/cases/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkResource struct{}

func sdkWrapperAllowed(_ context.Context, d *schema.ResourceData, meta any) error {
	client, _ := helpers.NewSDKKibanaClient(d, meta)
	return kibana.Do(client)
}

func sdkWrapperDenied(_ context.Context, d *schema.ResourceData, meta any) error {
	client, _ := helpers.NotAllowedKibanaWrapper(d, meta)
	return kibana.Do(client) // want "Kibana/Fleet client usage must use a helper-derived \\*clients.APIClient from clients.NewKibanaAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewKibanaAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func (r *frameworkResource) Read(ctx context.Context, req resource.ReadRequest, _ *resource.ReadResponse) {
	client, _ := helpers.NewFrameworkKibanaClient(ctx, req, r)
	_ = kibana.Do(client)
}
