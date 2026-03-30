package facts

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func inferredWrapperNoAllowlist(d *schema.ResourceData, meta any) error {
	client, _ := helpers.NewSDKClient(d, meta)
	return elasticsearch.Do(client)
}

func inferredNonCompliantWrapper(d *schema.ResourceData, meta any) error {
	client, _ := helpers.NotAllowedSDKWrapper(d, meta)
	return elasticsearch.Do(client) // want "Elasticsearch client usage must use a helper-derived \\*clients.APIClient from clients.NewAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func derivedReceiverThroughFact(ctx context.Context, d *schema.ResourceData, meta any) error {
	client, _ := helpers.NewSDKClient(d, meta)
	_, _ = client.ID(ctx, "id")
	return nil
}
