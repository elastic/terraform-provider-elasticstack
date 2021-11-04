package provider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"username": {
					Description:  "Username to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					RequiredWith: []string{"password"},
					DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
				},
				"password": {
					Description:  "Password to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					RequiredWith: []string{"username"},
					DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
				},
				"url": {
					Description: "Endpoint where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_URL", "http://localhost:9200"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"elasticstack_elasticsearch_security_user": security.DataSourceUser(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"elasticstack_elasticsearch_cluster_settings": cluster.ResourceSettings(),
				"elasticstack_elasticsearch_index_template":   index.ResourceTemplate(),
				"elasticstack_elasticsearch_security_role":    security.ResourceRole(),
				"elasticstack_elasticsearch_security_user":    security.ResourceUser(),
			},
		}

		p.ConfigureContextFunc = clients.NewApiClientFunc(version, p)

		return p
	}
}
