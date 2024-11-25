package config

import (
	"context"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NewFromFramework(ctx context.Context, cfg ProviderConfiguration, version string) (Client, diag.Diagnostics) {
	base := newBaseConfigFromFramework(cfg, version)
	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg, diags := newElasticsearchConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return Client{}, diags
	}

	if esCfg != nil {
		client.Elasticsearch = utils.Pointer(esCfg.toElasticsearchConfiguration())
	}

	kibanaCfg, diags := newKibanaConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Kibana = (*kibana.Config)(&kibanaCfg)

	kibanaOapiCfg, diags := newKibanaOapiConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.KibanaOapi = (*kibana_oapi.Config)(&kibanaOapiCfg)

	fleetCfg, diags := newFleetConfigFromFramework(ctx, cfg, kibanaOapiCfg)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client, nil
}
