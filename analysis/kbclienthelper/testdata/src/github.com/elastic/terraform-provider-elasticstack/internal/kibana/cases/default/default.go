package defaultcases

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkResource struct {
	client *clients.APIClient
}

func sdkReadOK(_ *schema.ResourceData, meta any) error {
	client, _ := clients.NewKibanaAPIClientFromSDKResource(nil, meta)
	return client.GetKibanaClient()
}

func sdkReadMissingHelper() error {
	client := &clients.APIClient{}
	return client.GetKibanaClient() // want "Kibana/Fleet client usage must use a helper-derived \\*clients.APIClient from clients.NewKibanaAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewKibanaAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func sdkReadFleetMissingHelper() error {
	client := &clients.APIClient{}
	return client.GetFleetClient() // want "Kibana/Fleet client usage must use a helper-derived \\*clients.APIClient from clients.NewKibanaAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewKibanaAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func sdkNoKibanaUsage(_ *schema.ResourceData, _ any) error {
	return nil
}

func (r *frameworkResource) Read(ctx context.Context, req resource.ReadRequest, _ *resource.ReadResponse) {
	client, _ := clients.MaybeNewKibanaAPIClientFromFrameworkResource(ctx, req, r.client)
	_ = client.GetKibanaClient()
}

func (r *frameworkResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	_ = r.client.GetKibanaClient() // want "Kibana/Fleet client usage must use a helper-derived \\*clients.APIClient from clients.NewKibanaAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewKibanaAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}
