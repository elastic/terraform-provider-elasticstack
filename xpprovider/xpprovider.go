package xpprovider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// Package xpprovider exports needed internal types and functions used by Crossplane for instantiating, interacting and
// configuring the underlying Terraform Elasticstack providers.

// XPAPIClient exports the internal type clients.APIClient of the Terraform provider
type XPAPIClient = clients.APIClient

// XPProviderConfiguration exports the internal type config.ProviderConfiguration of the Terraform provider
type XPProviderConfiguration = config.ProviderConfiguration

// XPElasticsearchConnection exports the internal type config.ElasticsearchConnection of the Terraform provider
type XPElasticsearchConnection = config.ElasticsearchConnection

// XPKibanaConnection exports the internal type config.KibanaConnection of the Terraform provider
type XPKibanaConnection = config.KibanaConnection

// XPFleetConnection exports the internal type config.FleetConnection of the Terraform provider
type XPFleetConnection = config.FleetConnection

func NewAPIClientFromFramework(ctx context.Context, config XPProviderConfiguration, version string) (*XPAPIClient, fwdiags.Diagnostics) {
	return clients.NewAPIClientFromFramework(ctx, config, version)
}
