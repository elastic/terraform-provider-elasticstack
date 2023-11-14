package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func NewFromEnv(version string) Client {
	base := baseConfig{
		UserAgent: buildUserAgent(version),
	}.withEnvironmentOverrides()

	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg := base.toElasticsearchConfig().withEnvironmentOverrides()
	client.Elasticsearch = utils.Pointer(elasticsearch.Config(esCfg))

	kibanaCfg := base.toKibanaConfig().withEnvironmentOverrides()
	client.Kibana = (*kibana.Config)(&kibanaCfg)

	fleetCfg := kibanaCfg.toFleetConfig().withEnvironmentOverrides()
	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client
}
