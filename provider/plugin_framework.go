package provider

import (
	"context"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/apm/agent_configuration"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/script"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/enrich"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/alias"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/data_stream_lifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indices"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/anomaly_detection_job"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed_state"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/job_state"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/role_mapping"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/system_user"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/user"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agent_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/enrollment_tokens"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_ds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/server_host"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/data_view"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/default_data_view"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/export_saved_objects"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/maintenance_window"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_detection_rule"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_exception_item"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_exception_list"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_list"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_list_item"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/spaces"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/monitor"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/parameter"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/private_location"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	IncludeExperimentalEnvVar = "TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL"
	AccTestVersion            = "acctest"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ fwprovider.Provider = &Provider{}
)

type Provider struct {
	version string
}

// NewFrameworkProvider instantiates plugin framework's provider
func NewFrameworkProvider(version string) fwprovider.Provider {
	return &Provider{
		version: version,
	}
}

func (p *Provider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, res *fwprovider.MetadataResponse) {
	res.TypeName = "elasticstack"
	res.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req fwprovider.SchemaRequest, res *fwprovider.SchemaResponse) {
	res.Schema = fwschema.Schema{
		Blocks: map[string]fwschema.Block{
			esKeyName:    schema.GetEsFWConnectionBlock(esKeyName, true),
			kbKeyName:    schema.GetKbFWConnectionBlock(),
			fleetKeyName: schema.GetFleetFWConnectionBlock(),
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, res *fwprovider.ConfigureResponse) {
	var cfg config.ProviderConfiguration

	res.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if res.Diagnostics.HasError() {
		return
	}

	client, diags := clients.NewApiClientFromFramework(ctx, cfg, p.version)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.DataSourceData = client
	res.ResourceData = client
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := p.dataSources(ctx)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == "true" {
		datasources = append(datasources, p.experimentalDataSources(ctx)...)
	}

	return datasources
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	resources := p.resources(ctx)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == "true" {
		resources = append(resources, p.experimentalResources(ctx)...)
	}

	return resources
}

func (p *Provider) resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		agent_configuration.NewAgentConfigurationResource,
		func() resource.Resource { return &import_saved_objects.Resource{} },
		data_view.NewResource,
		default_data_view.NewResource,
		func() resource.Resource { return &parameter.Resource{} },
		func() resource.Resource { return &private_location.Resource{} },
		func() resource.Resource { return &index.Resource{} },
		monitor.NewResource,
		func() resource.Resource { return &api_key.Resource{} },
		func() resource.Resource { return &data_stream_lifecycle.Resource{} },
		func() resource.Resource { return &connectors.Resource{} },
		agent_policy.NewResource,
		integration.NewResource,
		integration_policy.NewResource,
		output.NewResource,
		server_host.NewResource,
		system_user.NewSystemUserResource,
		user.NewUserResource,
		script.NewScriptResource,
		maintenance_window.NewResource,
		enrich.NewEnrichPolicyResource,
		role_mapping.NewRoleMappingResource,
		alias.NewAliasResource,
		datafeed.NewDatafeedResource,
		anomaly_detection_job.NewAnomalyDetectionJobResource,
		security_detection_rule.NewSecurityDetectionRuleResource,
		job_state.NewMLJobStateResource,
		datafeed_state.NewMLDatafeedStateResource,
	}
}

func (p *Provider) experimentalResources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		security_list_item.NewResource,
		security_list.NewResource,
		security_exception_list.NewResource,
		security_exception_item.NewResource,
	}
}

func (p *Provider) dataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		indices.NewDataSource,
		spaces.NewDataSource,
		export_saved_objects.NewDataSource,
		enrollment_tokens.NewDataSource,
		integration_ds.NewDataSource,
		enrich.NewEnrichPolicyDataSource,
		role_mapping.NewRoleMappingDataSource,
	}
}

func (p *Provider) experimentalDataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
