package flow

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func bothBranchesDerived(ctx context.Context, d *schema.ResourceData, meta any, cond bool) error {
	var client *clients.APIClient
	if cond {
		client, _ = clients.NewAPIClientFromSDKResource(d, meta)
	} else {
		client, _ = clients.NewAPIClientFromSDKResource(d, meta)
	}
	_, _ = client.ID(ctx, "id")
	return elasticsearch.Do(client)
}

func oneBranchDerived(ctx context.Context, d *schema.ResourceData, meta any, cond bool) error {
	var client *clients.APIClient
	if cond {
		client, _ = clients.NewAPIClientFromSDKResource(d, meta)
	} else {
		client = &clients.APIClient{}
	}
	_, _ = client.ID(ctx, "id")     // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
	return elasticsearch.Do(client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func aliasAndReassign(d *schema.ResourceData, meta any) error {
	client, _ := clients.NewAPIClientFromSDKResource(d, meta)
	alias := client
	_ = elasticsearch.Do(alias)
	alias = &clients.APIClient{}
	return elasticsearch.Do(alias) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func directClient() (*clients.APIClient, error) {
	return &clients.APIClient{}, nil
}

func multiAssignmentClearsDerived(d *schema.ResourceData, meta any) error {
	client, _ := clients.NewAPIClientFromSDKResource(d, meta)
	client, _ = directClient()
	return elasticsearch.Do(client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}
