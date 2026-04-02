// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indices"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateilmattachment"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/inference/inferenceendpoint"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/anomalydetectionjob"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed_state"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/jobstate"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/role"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/rolemapping"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/systemuser"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/user"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/enrollmenttokens"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integrationds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/outputds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/serverhost"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilderworkflow"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alertingrule"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dataview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/defaultdataview"
	exportagentbuilderworkflow "github.com/elastic/terraform-provider-elasticstack/internal/kibana/exportagentbuilder/workflow"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/exportsavedobjects"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/maintenance_window"
	prebuilt_rules "github.com/elastic/terraform-provider-elasticstack/internal/kibana/prebuilt_rules"
	security_detection_rule "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_detection_rule"
	securityenablerule "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_enable_rule"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_exception_item"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_list_data_streams"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securityexceptionlist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securitylist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securitylistitem"
	kibanaslo "github.com/elastic/terraform-provider-elasticstack/internal/kibana/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/spaces"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/streams"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/monitor"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/parameter"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/privatelocation"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	IncludeExperimentalEnvVar    = "TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL"
	SkipLocationValidationEnvVar = "TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION"
	AccTestVersion               = "acctest"
	envVarEnabled                = "true"
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

func (p *Provider) Schema(_ context.Context, _ fwprovider.SchemaRequest, res *fwprovider.SchemaResponse) {
	res.Schema = fwschema.Schema{
		Blocks: map[string]fwschema.Block{
			esKeyName:    schema.GetEsFWConnectionBlock(),
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

	client, diags := clients.NewAPIClientFromFramework(ctx, cfg, p.version)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.DataSourceData = client
	res.ResourceData = client
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := p.dataSources(ctx)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == envVarEnabled {
		datasources = append(datasources, p.experimentalDataSources(ctx)...)
	}

	return datasources
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	validateLocation := os.Getenv(SkipLocationValidationEnvVar) != envVarEnabled
	resources := p.resources(ctx, validateLocation)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == envVarEnabled {
		resources = append(resources, p.experimentalResources(ctx)...)
	}

	return resources
}

func (p *Provider) resources(_ context.Context, validateLocation bool) []func() resource.Resource {
	return []func() resource.Resource{
		agentconfiguration.NewAgentConfigurationResource,
		func() resource.Resource { return &importsavedobjects.Resource{} },
		alertingrule.NewResource,
		dataview.NewResource,
		defaultdataview.NewResource,
		func() resource.Resource { return &parameter.Resource{} },
		func() resource.Resource { return &privatelocation.Resource{} },
		func() resource.Resource { return &index.Resource{} },
		func() resource.Resource { return monitor.NewResource(validateLocation) },
		func() resource.Resource { return &apikey.Resource{} },
		func() resource.Resource { return &datastreamlifecycle.Resource{} },
		func() resource.Resource { return &connectors.Resource{} },
		agentpolicy.NewResource,
		agentbuilderworkflow.NewResource,
		integration.NewResource,
		integrationpolicy.NewResource,
		output.NewResource,
		serverhost.NewResource,
		systemuser.NewSystemUserResource,
		securityuser.NewUserResource,
		role.NewRoleResource,
		inferenceendpoint.NewInferenceEndpointResource,
		script.NewScriptResource,
		maintenancewindow.NewResource,
		enrich.NewEnrichPolicyResource,
		rolemapping.NewRoleMappingResource,
		alias.NewAliasResource,
		templateilmattachment.NewResource,
		datafeed.NewDatafeedResource,
		anomalydetectionjob.NewAnomalyDetectionJobResource,
		security_detection_rule.NewSecurityDetectionRuleResource,
		jobstate.NewMLJobStateResource,
		datafeedstate.NewMLDatafeedStateResource,
		kibanaslo.NewResource,
		prebuilt_rules.NewResource,
		securityenablerule.NewResource,
		securitylistitem.NewResource,
		securitylist.NewResource,
		securitylistdatastreams.NewResource,
		securityexceptionlist.NewResource,
		securityexceptionitem.NewResource,
	}
}

func (p *Provider) experimentalResources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		dashboard.NewResource,
		streams.NewResource,
	}
}

func (p *Provider) dataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		indices.NewDataSource,
		spaces.NewDataSource,
		exportagentbuilderworkflow.NewDataSource,
		exportsavedobjects.NewDataSource,
		enrollmenttokens.NewDataSource,
		integrationds.NewDataSource,
		enrich.NewEnrichPolicyDataSource,
		rolemapping.NewRoleMappingDataSource,
		outputds.NewDataSource,
	}
}

func (p *Provider) experimentalDataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
