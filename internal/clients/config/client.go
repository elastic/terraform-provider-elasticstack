package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

type Client struct {
	UserAgent     string
	Kibana        *kibana.Config
	Elasticsearch *elasticsearch.Config
	Fleet         *fleet.Config
}
