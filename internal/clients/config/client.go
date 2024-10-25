package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
)

type Client struct {
	UserAgent     string
	Kibana        *kibana.Config
	Kibana2       *kibana2.Config
	Elasticsearch *elasticsearch.Config
	Fleet         *fleet.Config
}
