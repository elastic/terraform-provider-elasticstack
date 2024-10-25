package config

import (
	"context"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
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

	kibana2Cfg, diags := newKibana2ConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Kibana2 = (*kibana2.Config)(&kibana2Cfg)

	fleetCfg, diags := newFleetConfigFromFramework(ctx, cfg, kibana2Cfg)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client, nil
}
