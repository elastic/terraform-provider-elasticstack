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

package agentpolicyds

import (
	"context"
	"math"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/globaldatatags"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type agentPolicyModel struct {
	entitycore.KibanaConnectionField
	ID                        types.String         `tfsdk:"id"`
	PolicyID                  types.String         `tfsdk:"policy_id"`
	Name                      types.String         `tfsdk:"name"`
	SpaceID                   types.String         `tfsdk:"space_id"`
	Namespace                 types.String         `tfsdk:"namespace"`
	Description               types.String         `tfsdk:"description"`
	DataOutputID              types.String         `tfsdk:"data_output_id"`
	MonitoringOutputID        types.String         `tfsdk:"monitoring_output_id"`
	FleetServerHostID         types.String         `tfsdk:"fleet_server_host_id"`
	DownloadSourceID          types.String         `tfsdk:"download_source_id"`
	MonitorLogs               types.Bool           `tfsdk:"monitor_logs"`
	MonitorMetrics            types.Bool           `tfsdk:"monitor_metrics"`
	IsProtected               types.Bool           `tfsdk:"is_protected"`
	HostNameFormat            types.String         `tfsdk:"host_name_format"`
	SupportsAgentless         types.Bool           `tfsdk:"supports_agentless"`
	InactivityTimeout         customtypes.Duration `tfsdk:"inactivity_timeout"`
	UnenrollmentTimeout       customtypes.Duration `tfsdk:"unenrollment_timeout"`
	GlobalDataTags            types.Map            `tfsdk:"global_data_tags"`
	SpaceIDs                  types.Set            `tfsdk:"space_ids"`
	RequiredVersions          types.Map            `tfsdk:"required_versions"`
	AdvancedSettings          types.Object         `tfsdk:"advanced_settings"`
	AdvancedMonitoringOptions types.Object         `tfsdk:"advanced_monitoring_options"`
}

func (m agentPolicyModel) GetID() types.String      { return m.ID }
func (m agentPolicyModel) GetSpaceID() types.String { return m.SpaceID }
func (m agentPolicyModel) GetResourceID() types.String {
	if !m.PolicyID.IsNull() && m.PolicyID.ValueString() != "" {
		return m.PolicyID
	}
	return types.StringValue("_")
}

const (
	hostNameFormatHostname = "hostname"
	hostNameFormatFQDN     = "fqdn"
	agentFeatureFQDN       = "fqdn"
)

func (m *agentPolicyModel) populateFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(data.Id)
	m.PolicyID = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.Namespace = types.StringValue(data.Namespace)
	m.Description = types.StringPointerValue(data.Description)
	m.DataOutputID = types.StringPointerValue(data.DataOutputId)
	m.MonitoringOutputID = types.StringPointerValue(data.MonitoringOutputId)
	m.FleetServerHostID = types.StringPointerValue(data.FleetServerHostId)
	m.DownloadSourceID = types.StringPointerValue(data.DownloadSourceId)
	m.MonitorLogs = types.BoolValue(monitoringEnabled(data, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledLogs))
	m.MonitorMetrics = types.BoolValue(monitoringEnabled(data, kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledMetrics))
	m.IsProtected = types.BoolValue(data.IsProtected)
	m.HostNameFormat = types.StringValue(hostNameFormatFromAPI(data))
	m.SupportsAgentless = types.BoolPointerValue(data.SupportsAgentless)
	m.InactivityTimeout = customtypes.NewDurationFromSeconds(data.InactivityTimeout)
	m.UnenrollmentTimeout = customtypes.NewDurationFromSeconds(data.UnenrollTimeout)

	if data.GlobalDataTags == nil {
		m.GlobalDataTags = types.MapNull(globaldatatags.ElementType())
	} else {
		items := make([]globaldatatags.Tag[kbapi.AgentPolicyGlobalDataTagsItem_Value], len(*data.GlobalDataTags))
		for i, tag := range *data.GlobalDataTags {
			items[i] = globaldatatags.Tag[kbapi.AgentPolicyGlobalDataTagsItem_Value]{Name: tag.Name, Value: tag.Value}
		}
		m.GlobalDataTags = globaldatatags.ToModel(
			ctx,
			items,
			path.Root("global_data_tags"),
			&diags,
			kbapi.AgentPolicyGlobalDataTagsItem_Value.AsAgentPolicyGlobalDataTagsItemValue1,
			kbapi.AgentPolicyGlobalDataTagsItem_Value.AsAgentPolicyGlobalDataTagsItemValue0,
		)
		if diags.HasError() {
			return diags
		}
	}

	if data.SpaceIds == nil {
		m.SpaceIDs = types.SetNull(types.StringType)
	} else {
		m.SpaceIDs = typeutils.SetValueFrom(ctx, *data.SpaceIds, types.StringType, path.Root("space_ids"), &diags)
		if diags.HasError() {
			return diags
		}
	}

	m.RequiredVersions = requiredVersionsFromAPI(data.RequiredVersions, &diags)
	if diags.HasError() {
		return diags
	}

	diags.Append(m.populateAdvancedSettingsFromAPI(ctx, data)...)
	if diags.HasError() {
		return diags
	}
	diags.Append(m.populateAdvancedMonitoringFromAPI(ctx, data)...)
	return diags
}

func monitoringEnabled(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse, kind kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabled) bool {
	return data.MonitoringEnabled != nil && slices.Contains(*data.MonitoringEnabled, kind)
}

func hostNameFormatFromAPI(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) string {
	if data.AgentFeatures != nil {
		for _, feature := range *data.AgentFeatures {
			if feature.Name == agentFeatureFQDN && feature.Enabled {
				return hostNameFormatFQDN
			}
		}
	}
	return hostNameFormatHostname
}

func requiredVersionsFromAPI(versions *[]struct {
	Percentage float32 `json:"percentage"`
	Version    string  `json:"version"`
}, diags *diag.Diagnostics) types.Map {
	if versions == nil {
		return types.MapNull(types.Int32Type)
	}
	versionMap := make(map[string]attr.Value, len(*versions))
	for _, rv := range *versions {
		versionMap[rv.Version] = types.Int32Value(int32(math.Round(float64(rv.Percentage))))
	}
	mapped, mapDiags := types.MapValue(types.Int32Type, versionMap)
	diags.Append(mapDiags...)
	if mapDiags.HasError() {
		return types.MapNull(types.Int32Type)
	}
	return mapped
}

// populateAdvancedSettingsFromAPI populates advanced_settings using the
// Fleet API population logic shared with the elasticstack_fleet_agent_policy
// resource via agentpolicyshape.
func (m *agentPolicyModel) populateAdvancedSettingsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	attrTypes := objectAttrTypes(ctx, "advanced_settings")
	settings, ok, diags := agentpolicyshape.PopulateAdvancedSettings(data)
	if !ok {
		m.AdvancedSettings = types.ObjectNull(attrTypes)
		return diags
	}

	obj, objDiags := types.ObjectValueFrom(ctx, attrTypes, settings)
	diags.Append(objDiags...)
	if diags.HasError() {
		return diags
	}
	m.AdvancedSettings = obj
	return diags
}

// Advanced Monitoring Options models, shared with the
// elasticstack_fleet_agent_policy resource via agentpolicyshape. The data
// source has no schema defaults, so API-omitted fields are always left null.
type advancedMonitoringOptionsModel = agentpolicyshape.AdvancedMonitoringOptionsModel

type httpMonitoringEndpointModel = agentpolicyshape.HTTPMonitoringEndpointModel

type diagnosticsModel = agentpolicyshape.DiagnosticsModel

type rateLimitsModel = agentpolicyshape.RateLimitsModel

type fileUploaderModel = agentpolicyshape.FileUploaderModel

func (m *agentPolicyModel) populateAdvancedMonitoringFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	httpEndpoint, hasHTTP := agentpolicyshape.PopulateHTTPMonitoringEndpoint(data, httpMonitoringEndpointModel{})
	hasDiagnostics := agentpolicyshape.HasDiagnostics(data)
	if !hasHTTP && !hasDiagnostics {
		m.AdvancedMonitoringOptions = types.ObjectNull(objectAttrTypes(ctx, "advanced_monitoring_options"))
		return nil
	}

	httpAttrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "http_monitoring_endpoint")
	var httpObj types.Object
	var diags diag.Diagnostics
	if hasHTTP {
		obj, objDiags := types.ObjectValueFrom(ctx, httpAttrTypes, httpEndpoint)
		diags.Append(objDiags...)
		if diags.HasError() {
			return diags
		}
		httpObj = obj
	} else {
		httpObj = types.ObjectNull(httpAttrTypes)
	}

	diagnosticsObj, diagDiags := diagnosticsFromAPI(ctx, data, hasDiagnostics)
	diags.Append(diagDiags...)
	if diags.HasError() {
		return diags
	}

	obj, objDiags := types.ObjectValueFrom(ctx, objectAttrTypes(ctx, "advanced_monitoring_options"), advancedMonitoringOptionsModel{
		HTTPMonitoringEndpoint: httpObj,
		Diagnostics:            diagnosticsObj,
	})
	diags.Append(objDiags...)
	if diags.HasError() {
		return diags
	}
	m.AdvancedMonitoringOptions = obj
	return diags
}

func diagnosticsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse, hasDiagnostics bool) (types.Object, diag.Diagnostics) {
	attrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics")
	if !hasDiagnostics {
		return types.ObjectNull(attrTypes), nil
	}

	rateLimitsAttrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics", "rate_limits")
	var rateLimitsObj types.Object
	var diags diag.Diagnostics
	if rateLimits, ok := agentpolicyshape.PopulateRateLimits(data, rateLimitsModel{}); ok {
		obj, objDiags := types.ObjectValueFrom(ctx, rateLimitsAttrTypes, rateLimits)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ObjectNull(attrTypes), diags
		}
		rateLimitsObj = obj
	} else {
		rateLimitsObj = types.ObjectNull(rateLimitsAttrTypes)
	}

	fileUploaderAttrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics", "file_uploader")
	var fileUploaderObj types.Object
	if fileUploader, ok := agentpolicyshape.PopulateFileUploader(data, fileUploaderModel{}); ok {
		obj, objDiags := types.ObjectValueFrom(ctx, fileUploaderAttrTypes, fileUploader)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ObjectNull(attrTypes), diags
		}
		fileUploaderObj = obj
	} else {
		fileUploaderObj = types.ObjectNull(fileUploaderAttrTypes)
	}

	obj, objDiags := types.ObjectValueFrom(ctx, attrTypes, diagnosticsModel{
		RateLimits:   rateLimitsObj,
		FileUploader: fileUploaderObj,
	})
	diags.Append(objDiags...)
	return obj, diags
}

func objectAttrTypes(ctx context.Context, names ...string) map[string]attr.Type {
	attribute := getDataSourceSchema(ctx).Attributes[names[0]]
	for _, name := range names[1:] {
		attribute = attribute.(dsschema.SingleNestedAttribute).Attributes[name]
	}
	return attribute.GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
