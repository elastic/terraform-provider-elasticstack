package config

import (
	"net/http"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func NewFromEnv(version string) Client {
	ua := buildUserAgent(version)
	base := baseConfig{
		UserAgent: ua,
		Header:    http.Header{"User-Agent": []string{ua}},
	}.withEnvironmentOverrides()

	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg := base.toElasticsearchConfig().withEnvironmentOverrides()
	client.Elasticsearch = utils.Pointer(esCfg.toElasticsearchConfiguration())

	kibanaCfg := base.toKibanaConfig().withEnvironmentOverrides()
	client.Kibana = (*kibana.Config)(&kibanaCfg)

	fleetCfg := kibanaCfg.toFleetConfig().withEnvironmentOverrides()
	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client
}
