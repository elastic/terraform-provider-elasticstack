package config

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
)

type Client struct {
	UserAgent     string
	Kibana        *kibana.Config
	KibanaOapi    *kibanaoapi.Config
	Elasticsearch *elasticsearch.Config
	Fleet         *fleet.Config
}
