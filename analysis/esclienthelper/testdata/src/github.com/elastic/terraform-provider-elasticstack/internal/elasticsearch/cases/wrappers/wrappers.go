package wrappers

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkResource struct{}

func sdkWrapperAllowed(_ context.Context, d *schema.ResourceData, meta any) error {
	client, _ := helpers.NewSDKClient(d, meta)
	return elasticsearch.Do(client)
}

func sdkWrapperDenied(_ context.Context, d *schema.ResourceData, meta any) error {
	client, _ := helpers.NotAllowedSDKWrapper(d, meta)
	return elasticsearch.Do(client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func (r *frameworkResource) Read(ctx context.Context, req resource.ReadRequest, _ *resource.ReadResponse) {
	client, _ := helpers.NewFrameworkClient(ctx, req, r)
	_ = elasticsearch.Do(client)
}
