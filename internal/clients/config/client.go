package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
)

type Client struct {
	UserAgent     string
	Kibana        *kibana.Config
	KibanaOapi    *kibana_oapi.Config
	Elasticsearch *elasticsearch.Config
	Fleet         *fleet.Config
}
