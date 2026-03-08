package scope

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
)

// This package intentionally lives outside internal/elasticsearch, so analyzer should skip it.
func shouldBeIgnored(ctx context.Context) error {
	client := &clients.APIClient{}
	_, _ = client.ID(ctx, "id")
	return elasticsearch.Do(client)
}
