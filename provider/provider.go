package provider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ingest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/logstash"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const esKeyName = "elasticsearch"

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			esKeyName: providerSchema.GetConnectionSchema(esKeyName, true),
			"kibana": {
				Description: "Kibana connection configuration block.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description:  "Username to use for API authentication to Kibana.",
							Type:         schema.TypeString,
							Optional:     true,
							DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
							RequiredWith: []string{"kibana.0.password"},
							ExactlyOneOf: []string{"elasticsearch.0.username", "kibana.0.username"},
						},
						"password": {
							Description:  "Password to use for API authentication to Kibana.",
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
							RequiredWith: []string{"kibana.0.username"},
							ExactlyOneOf: []string{"elasticsearch.0.password", "kibana.0.password"},
						},
						"endpoints": {
							Description: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
							Type:        schema.TypeList,
							MaxItems:    1, // Current API restriction
							Required:    true,
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
					},
				},
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"elasticstack_elasticsearch_ingest_processor_append":            ingest.DataSourceProcessorAppend(),
			"elasticstack_elasticsearch_ingest_processor_bytes":             ingest.DataSourceProcessorBytes(),
			"elasticstack_elasticsearch_ingest_processor_circle":            ingest.DataSourceProcessorCircle(),
			"elasticstack_elasticsearch_ingest_processor_community_id":      ingest.DataSourceProcessorCommunityId(),
			"elasticstack_elasticsearch_ingest_processor_convert":           ingest.DataSourceProcessorConvert(),
			"elasticstack_elasticsearch_ingest_processor_csv":               ingest.DataSourceProcessorCSV(),
			"elasticstack_elasticsearch_ingest_processor_date":              ingest.DataSourceProcessorDate(),
			"elasticstack_elasticsearch_ingest_processor_date_index_name":   ingest.DataSourceProcessorDateIndexName(),
			"elasticstack_elasticsearch_ingest_processor_dissect":           ingest.DataSourceProcessorDissect(),
			"elasticstack_elasticsearch_ingest_processor_dot_expander":      ingest.DataSourceProcessorDotExpander(),
			"elasticstack_elasticsearch_ingest_processor_drop":              ingest.DataSourceProcessorDrop(),
			"elasticstack_elasticsearch_ingest_processor_enrich":            ingest.DataSourceProcessorEnrich(),
			"elasticstack_elasticsearch_ingest_processor_fail":              ingest.DataSourceProcessorFail(),
			"elasticstack_elasticsearch_ingest_processor_fingerprint":       ingest.DataSourceProcessorFingerprint(),
			"elasticstack_elasticsearch_ingest_processor_foreach":           ingest.DataSourceProcessorForeach(),
			"elasticstack_elasticsearch_ingest_processor_geoip":             ingest.DataSourceProcessorGeoip(),
			"elasticstack_elasticsearch_ingest_processor_grok":              ingest.DataSourceProcessorGrok(),
			"elasticstack_elasticsearch_ingest_processor_gsub":              ingest.DataSourceProcessorGsub(),
			"elasticstack_elasticsearch_ingest_processor_html_strip":        ingest.DataSourceProcessorHtmlStrip(),
			"elasticstack_elasticsearch_ingest_processor_join":              ingest.DataSourceProcessorJoin(),
			"elasticstack_elasticsearch_ingest_processor_json":              ingest.DataSourceProcessorJson(),
			"elasticstack_elasticsearch_ingest_processor_kv":                ingest.DataSourceProcessorKV(),
			"elasticstack_elasticsearch_ingest_processor_lowercase":         ingest.DataSourceProcessorLowercase(),
			"elasticstack_elasticsearch_ingest_processor_network_direction": ingest.DataSourceProcessorNetworkDirection(),
			"elasticstack_elasticsearch_ingest_processor_pipeline":          ingest.DataSourceProcessorPipeline(),
			"elasticstack_elasticsearch_ingest_processor_registered_domain": ingest.DataSourceProcessorRegisteredDomain(),
			"elasticstack_elasticsearch_ingest_processor_remove":            ingest.DataSourceProcessorRemove(),
			"elasticstack_elasticsearch_ingest_processor_rename":            ingest.DataSourceProcessorRename(),
			"elasticstack_elasticsearch_ingest_processor_script":            ingest.DataSourceProcessorScript(),
			"elasticstack_elasticsearch_ingest_processor_set":               ingest.DataSourceProcessorSet(),
			"elasticstack_elasticsearch_ingest_processor_set_security_user": ingest.DataSourceProcessorSetSecurityUser(),
			"elasticstack_elasticsearch_ingest_processor_sort":              ingest.DataSourceProcessorSort(),
			"elasticstack_elasticsearch_ingest_processor_split":             ingest.DataSourceProcessorSplit(),
			"elasticstack_elasticsearch_ingest_processor_trim":              ingest.DataSourceProcessorTrim(),
			"elasticstack_elasticsearch_ingest_processor_uppercase":         ingest.DataSourceProcessorUppercase(),
			"elasticstack_elasticsearch_ingest_processor_urldecode":         ingest.DataSourceProcessorUrldecode(),
			"elasticstack_elasticsearch_ingest_processor_uri_parts":         ingest.DataSourceProcessorUriParts(),
			"elasticstack_elasticsearch_ingest_processor_user_agent":        ingest.DataSourceProcessorUserAgent(),
			"elasticstack_elasticsearch_security_role":                      security.DataSourceRole(),
			"elasticstack_elasticsearch_security_role_mapping":              security.DataSourceRoleMapping(),
			"elasticstack_elasticsearch_security_user":                      security.DataSourceUser(),
			"elasticstack_elasticsearch_snapshot_repository":                cluster.DataSourceSnapshotRespository(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"elasticstack_elasticsearch_cluster_settings":      cluster.ResourceSettings(),
			"elasticstack_elasticsearch_component_template":    index.ResourceComponentTemplate(),
			"elasticstack_elasticsearch_data_stream":           index.ResourceDataStream(),
			"elasticstack_elasticsearch_index":                 index.ResourceIndex(),
			"elasticstack_elasticsearch_index_lifecycle":       index.ResourceIlm(),
			"elasticstack_elasticsearch_index_template":        index.ResourceTemplate(),
			"elasticstack_elasticsearch_ingest_pipeline":       ingest.ResourceIngestPipeline(),
			"elasticstack_elasticsearch_logstash_pipeline":     logstash.ResourceLogstashPipeline(),
			"elasticstack_elasticsearch_security_api_key":      security.ResourceApiKey(),
			"elasticstack_elasticsearch_security_role":         security.ResourceRole(),
			"elasticstack_elasticsearch_security_role_mapping": security.ResourceRoleMapping(),
			"elasticstack_elasticsearch_security_user":         security.ResourceUser(),
			"elasticstack_elasticsearch_security_system_user":  security.ResourceSystemUser(),
			"elasticstack_elasticsearch_snapshot_lifecycle":    cluster.ResourceSlm(),
			"elasticstack_elasticsearch_snapshot_repository":   cluster.ResourceSnapshotRepository(),
			"elasticstack_elasticsearch_script":                cluster.ResourceScript(),
		},
	}

	p.ConfigureContextFunc = clients.NewApiClientFunc(version)

	return p
}
