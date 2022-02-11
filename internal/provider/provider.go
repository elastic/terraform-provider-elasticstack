package provider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ingest"
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
				"elasticsearch": {
					Description: "Default Elasticsearch connection configuration block.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"username": {
								Description: "Username to use for API authentication to Elasticsearch.",
								Type:        schema.TypeString,
								Optional:    true,
								DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
							},
							"password": {
								Description: "Password to use for API authentication to Elasticsearch.",
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
							},
							"endpoints": {
								Description: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
								Type:        schema.TypeList,
								Optional:    true,
								Sensitive:   true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"insecure": {
								Description: "Disable TLS certificate validation",
								Type:        schema.TypeBool,
								Optional:    true,
								Default:     false,
							},
							"ca_file": {
								Description: "Path to a custom Certificate Authority certificate",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"elasticstack_elasticsearch_security_user":       security.DataSourceUser(),
				"elasticstack_elasticsearch_snapshot_repository": cluster.DataSourceSnapshotRespository(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"elasticstack_elasticsearch_cluster_settings":      cluster.ResourceSettings(),
				"elasticstack_elasticsearch_component_template":    index.ResourceComponentTemplate(),
				"elasticstack_elasticsearch_data_stream":           index.ResourceDataStream(),
				"elasticstack_elasticsearch_index":                 index.ResourceIndex(),
				"elasticstack_elasticsearch_index_lifecycle":       index.ResourceIlm(),
				"elasticstack_elasticsearch_index_template":        index.ResourceTemplate(),
				"elasticstack_elasticsearch_ingest_pipeline":       ingest.ResourceIngestPipeline(),
				"elasticstack_elasticsearch_security_role":         security.ResourceRole(),
				"elasticstack_elasticsearch_security_role_mapping": security.ResourceRoleMapping(),
				"elasticstack_elasticsearch_security_user":         security.ResourceUser(),
				"elasticstack_elasticsearch_snapshot_lifecycle":    cluster.ResourceSlm(),
				"elasticstack_elasticsearch_snapshot_repository":   cluster.ResourceSnapshotRepository(),
			},
		}

		p.ConfigureContextFunc = clients.NewApiClientFunc(version, p)

		return p
	}
}
