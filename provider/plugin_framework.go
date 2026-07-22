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

	agentconfiguration "github.com/elastic/terraform-provider-elasticstack/internal/apm/agent_configuration"
	sourcemap "github.com/elastic/terraform-provider-elasticstack/internal/apm/source_map"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr/autofollow"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ccr/followerindex"
	clusterinfo "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/info"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/script"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/settings"
	connectordatasource "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector/data_source"
	connectorresource "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector/resource"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector/sync_job_create"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/enrich"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/alias"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/componenttemplate"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastream"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/ilm"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indexmappings"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indices"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/template"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateilmattachment"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/inference/inferenceendpoint"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ingest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/logstash"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/anomalydetectionjob"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/calendar"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/calendar_event"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/calendar_job"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	datafeedstate "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed_state"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/filter"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/jobstate"
	mltrainedmodel "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/trainedmodel"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/trainedmodelalias"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/trainedmodeldeployment"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/queryrulesets"
	apikeyephemeral "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey/ephemeral"
	apikeyresource "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey/resource"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/role"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/rolemapping"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/systemuser"
	securityuser "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/user"
	snapshotcreate "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/snapshot/create"
	snapshotlifecycle "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/snapshot/lifecycle"
	snapshotrepo "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/snapshot/repository"
	snapshotrestore "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/snapshot/restore"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/synonyms"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/transform"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/watcher/watch"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentdownloadsource"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/customintegration"
	elasticdefendintegrationpolicy "github.com/elastic/terraform-provider-elasticstack/internal/fleet/elastic_defend_integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/enrollmenttokens"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration"
	integrationpolicy "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integrationds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/managedintegration"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/outputds"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/proxy"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/serverhost"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilderagent"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilderskill"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuildertool"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilderworkflow"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alertingrule"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dataview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/defaultdataview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/exportsavedobjects"
	importsavedobjects "github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	maintenancewindow "github.com/elastic/terraform-provider-elasticstack/internal/kibana/maintenance_window"
	osquerypack "github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery_pack"
	osquerysavedquery "github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery_saved_query"
	prebuilt_rules "github.com/elastic/terraform-provider-elasticstack/internal/kibana/prebuilt_rules"
	security_detection_rule "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_detection_rule"
	securityenablerule "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_enable_rule"
	securityentitystore "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store"
	securityentitystoreentities "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entities"
	securityentitystoreentity "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entity"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store_entity_link"
	securityentitystoreresolutiongroup "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store_resolution_group"
	securityexceptionitem "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_exception_item"
	securitylistdatastreams "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_list_data_streams"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_role"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securityexceptionlist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securitylist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/securitylistitem"
	kibanaslo "github.com/elastic/terraform-provider-elasticstack/internal/kibana/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/spaces"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/streams"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/monitor"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/parameter"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/privatelocation"
	kibanatag "github.com/elastic/terraform-provider-elasticstack/internal/kibana/tag"
	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	esKeyName    = "elasticsearch"
	kbKeyName    = "kibana"
	fleetKeyName = "fleet"

	// ProviderTypeName is the Terraform provider type name (provider "elasticstack" { ... }).
	ProviderTypeName = "elasticstack"

	IncludeExperimentalEnvVar = "TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL"
	AccTestVersion            = "acctest"
	envVarEnabled             = "true"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ fwprovider.Provider                       = &Provider{}
	_ fwprovider.ProviderWithEphemeralResources = &Provider{}
	_ fwprovider.ProviderWithActions            = &Provider{}
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
	res.TypeName = ProviderTypeName
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

	factory, diags := clients.NewProviderClientFactoryFromFramework(ctx, cfg, p.version)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.DataSourceData = factory
	res.ResourceData = factory
	res.EphemeralResourceData = factory
	res.ActionData = factory
}

func (p *Provider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		snapshotrestore.NewRestoreAction,
		snapshotcreate.NewCreateAction,
		sync_job_create.NewAction,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := p.dataSources(ctx)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == envVarEnabled {
		datasources = append(datasources, p.experimentalDataSources(ctx)...)
	}

	return datasources
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	resources := p.resources(ctx)

	if p.version == AccTestVersion || os.Getenv(IncludeExperimentalEnvVar) == envVarEnabled {
		resources = append(resources, p.experimentalResources(ctx)...)
	}

	return resources
}

func (p *Provider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		apikeyephemeral.NewResource,
	}
}

func (p *Provider) resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		agentconfiguration.NewAgentConfigurationResource,
		sourcemap.NewSourceMapResource,
		importsavedobjects.NewResource,
		alertingrule.NewResource,
		dashboard.NewResource,
		dataview.NewResource,
		defaultdataview.NewResource,
		parameter.NewResource,
		privatelocation.NewResource,
		index.NewResource,
		componenttemplate.NewResource,
		monitor.NewResource,
		apikeyresource.NewResource,
		datastream.NewDataStreamResource,
		datastreamlifecycle.NewResource,
		ilm.NewResource,
		template.NewResource,
		connectors.NewResource,
		agentpolicy.NewResource,
		agentbuilderagent.NewResource,
		agentbuilderskill.NewResource,
		agentbuildertool.NewResource,
		agentbuilderworkflow.NewResource,
		integration.NewResource,
		integrationpolicy.NewResource,
		customintegration.NewResource,
		elasticdefendintegrationpolicy.NewResource,
		output.NewResource,
		agentdownloadsource.NewResource,
		serverhost.NewResource,
		proxy.NewResource,
		systemuser.NewSystemUserResource,
		securityuser.NewUserResource,
		role.NewRoleResource,
		inferenceendpoint.NewInferenceEndpointResource,
		watch.NewWatchResource,
		settings.NewClusterSettingsResource,
		script.NewScriptResource,
		logstash.NewLogstashPipelineResource,
		maintenancewindow.NewResource,
		osquerypack.NewResource,
		osquerysavedquery.NewResource,
		enrich.NewEnrichPolicyResource,
		synonyms.NewSynonymSetResource,
		connectorresource.NewContentConnectorResource,
		queryrulesets.NewQueryRulesetResource,
		ingest.NewIngestPipelineResource,
		rolemapping.NewRoleMappingResource,
		alias.NewAliasResource,
		indexmappings.NewIndexMappingsResource,
		templateilmattachment.NewResource,
		datafeed.NewDatafeedResource,
		anomalydetectionjob.NewAnomalyDetectionJobResource,
		calendar.NewCalendarResource,
		calendar_event.NewCalendarEventResource,
		calendar_job.NewCalendarJobResource,
		filter.NewFilterResource,
		trainedmodelalias.NewTrainedModelAliasResource,
		security_detection_rule.NewSecurityDetectionRuleResource,
		jobstate.NewMLJobStateResource,
		trainedmodeldeployment.NewTrainedModelDeploymentResource,
		datafeedstate.NewMLDatafeedStateResource,
		kibanaslo.NewResource,
		prebuilt_rules.NewResource,
		securityenablerule.NewResource,
		securitylistitem.NewResource,
		securitylist.NewResource,
		securitylistdatastreams.NewResource,
		securityexceptionlist.NewResource,
		securityexceptionitem.NewResource,
		security_entity_store_entity_link.NewEntityLinkResource,
		security_role.NewResource,
		securityentitystore.NewResource,
		securityentitystoreentity.NewResource,
		spaces.NewResource,
		snapshotlifecycle.NewSlmResource,
		snapshotrepo.NewSnapshotRepositoryResource,
		transform.NewTransformResource,
		followerindex.NewFollowerIndexResource,
		autofollow.NewAutoFollowPatternResource,
		streams.NewResource,
	}
}

func (p *Provider) experimentalResources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		kibanatag.NewResource,
		managedintegration.NewResource,
	}
}

func (p *Provider) dataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		snapshotrepo.NewSnapshotRepositoryDataSource,
		clusterinfo.NewDataSource,
		indices.NewDataSource,
		template.NewDataSource,
		spaces.NewDataSource,
		security_role.NewDataSource,
		securityentitystoreresolutiongroup.NewDataSource,
		securityentitystore.NewDataSource,
		securityentitystoreentities.NewDataSource,
		connectors.NewDataSource,
		osquerysavedquery.NewDataSource,
		agentbuilderagent.NewDataSource,
		agentbuilderskill.NewDataSource,
		agentbuildertool.NewDataSource,
		agentbuilderworkflow.NewDataSource,
		exportsavedobjects.NewDataSource,
		enrollmenttokens.NewDataSource,
		integrationds.NewDataSource,
		enrich.NewEnrichPolicyDataSource,
		synonyms.NewSynonymSetDataSource,
		connectordatasource.NewContentConnectorDataSource,
		queryrulesets.NewQueryRulesetDataSource,
		rolemapping.NewRoleMappingDataSource,
		role.NewRoleDataSource,
		securityuser.NewUserDataSource,
		outputds.NewDataSource,
		osquerypack.NewDataSource,
		ingest.NewProcessorAppendDataSource,
		ingest.NewProcessorBytesDataSource,
		ingest.NewProcessorCircleDataSource,
		ingest.NewProcessorCommunityIDDataSource,
		ingest.NewProcessorConvertDataSource,
		ingest.NewProcessorCSVDataSource,
		ingest.NewProcessorDateDataSource,
		ingest.NewProcessorDateIndexNameDataSource,
		ingest.NewProcessorDissectDataSource,
		ingest.NewProcessorDotExpanderDataSource,
		ingest.NewProcessorDropDataSource,
		ingest.NewProcessorEnrichDataSource,
		ingest.NewProcessorFailDataSource,
		ingest.NewProcessorFingerprintDataSource,
		ingest.NewProcessorForeachDataSource,
		ingest.NewProcessorGeoIPDataSource,
		ingest.NewProcessorGrokDataSource,
		ingest.NewProcessorGsubDataSource,
		ingest.NewProcessorHTMLStripDataSource,
		ingest.NewProcessorInferenceDataSource,
		ingest.NewProcessorJoinDataSource,
		ingest.NewProcessorJSONDataSource,
		ingest.NewProcessorKVDataSource,
		ingest.NewProcessorLowercaseDataSource,
		ingest.NewProcessorNetworkDirectionDataSource,
		ingest.NewProcessorPipelineDataSource,
		ingest.NewProcessorRegisteredDomainDataSource,
		ingest.NewProcessorRemoveDataSource,
		ingest.NewProcessorRenameDataSource,
		ingest.NewProcessorRerouteDataSource,
		ingest.NewProcessorScriptDataSource,
		ingest.NewProcessorSetDataSource,
		ingest.NewProcessorSetSecurityUserDataSource,
		ingest.NewProcessorSortDataSource,
		ingest.NewProcessorSplitDataSource,
		ingest.NewProcessorTrimDataSource,
		ingest.NewProcessorUppercaseDataSource,
		ingest.NewProcessorURIPartsDataSource,
		ingest.NewProcessorURLDecodeDataSource,
		ingest.NewProcessorUserAgentDataSource,
		mltrainedmodel.NewDataSource,
	}
}

func (p *Provider) experimentalDataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		kibanatag.NewDataSource,
	}
}
