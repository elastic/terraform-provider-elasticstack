package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	esKey           string = "elasticsearch"
	esConnectionKey string = "elasticsearch_connection"
)

func NewFromSDK(d *schema.ResourceData, version string) (Client, diag.Diagnostics) {
	return newFromSDK(d, version, esKey)
}

func NewFromSDKResource(d *schema.ResourceData, version string) (*Client, diag.Diagnostics) {
	if _, ok := d.GetOk(esConnectionKey); !ok {
		return nil, nil
	}

	client, diags := newFromSDK(d, version, esConnectionKey)
	return &client, diags
}

func newFromSDK(d *schema.ResourceData, version, esConfigKey string) (Client, diag.Diagnostics) {
	base := newBaseConfigFromSDK(d, version, esConfigKey)
	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg, diags := newElasticsearchConfigFromSDK(d, base, esConfigKey, true)
	if diags.HasError() {
		return Client{}, diags
	}

	if esCfg != nil {
		client.Elasticsearch = utils.Pointer(esCfg.toElasticsearchConfiguration())
	}

	kibanaCfg, diags := newKibanaConfigFromSDK(d, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Kibana = (*kibana.Config)(&kibanaCfg)

	kibanaOapiCfg, diags := newKibanaOapiConfigFromSDK(d, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.KibanaOapi = (*kibana_oapi.Config)(&kibanaOapiCfg)

	fleetCfg, diags := newFleetConfigFromSDK(d, kibanaOapiCfg)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client, nil
}
