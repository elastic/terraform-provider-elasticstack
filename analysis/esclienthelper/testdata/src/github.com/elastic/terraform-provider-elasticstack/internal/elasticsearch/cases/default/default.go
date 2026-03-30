package defaultcases

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkResource struct {
	client *clients.APIClient
}

func sdkReadOK(ctx context.Context, d *schema.ResourceData, meta any) error {
	client, _ := clients.NewAPIClientFromSDKResource(d, meta)
	_, _ = client.ID(ctx, "id")
	return elasticsearch.Do(client)
}

func sdkReadMissingHelper(ctx context.Context, d *schema.ResourceData, meta any) error {
	_ = d
	_ = meta
	client := &clients.APIClient{}
	_, _ = client.ID(ctx, "id")     // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
	return elasticsearch.Do(client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func sdkReadWrongHelper(ctx context.Context, d *schema.ResourceData, meta any) error {
	_ = d
	client, _ := clients.MaybeNewAPIClientFromFrameworkResource(ctx, nil, meta)
	return elasticsearch.Do(client)
}

func sdkNoESUsage(_ context.Context, d *schema.ResourceData, _ any) error {
	_ = d
	return nil
}

func (r *frameworkResource) Read(ctx context.Context, req resource.ReadRequest, _ *resource.ReadResponse) {
	client, _ := clients.MaybeNewAPIClientFromFrameworkResource(ctx, req, r.client)
	_ = elasticsearch.Do(client)
}

func (r *frameworkResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	_ = r.client.GetESClient() // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func (r *frameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, _ *resource.DeleteResponse) {
	client, _ := clients.NewAPIClientFromSDKResource(nil, req)
	_ = elasticsearch.Do(client)
}
