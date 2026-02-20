package config

import (
	"net/http"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
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
	client.Elasticsearch = schemautil.Pointer(esCfg.toElasticsearchConfiguration())

	kibanaCfg := base.toKibanaConfig().withEnvironmentOverrides()
	client.Kibana = (*kibana.Config)(&kibanaCfg)

	kibanaOapiCfg := base.toKibanaOapiConfig().withEnvironmentOverrides()
	client.KibanaOapi = (*kibanaoapi.Config)(&kibanaOapiCfg)

	fleetCfg := kibanaOapiCfg.toFleetConfig().withEnvironmentOverrides()
	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client
}
