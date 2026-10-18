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
	"encoding/json"
	"math"
	"slices"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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
	m.InactivityTimeout = durationFromSeconds(data.InactivityTimeout)
	m.UnenrollmentTimeout = durationFromSeconds(data.UnenrollTimeout)

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

func durationFromSeconds(seconds *float32) customtypes.Duration {
	if seconds == nil {
		return customtypes.NewDurationNull()
	}
	d := time.Duration(int64(*seconds)) * time.Second
	return customtypes.NewDurationValue(d.String())
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

type advancedSettingsModel struct {
	LoggingLevel                           types.String                    `tfsdk:"logging_level"`
	LoggingToFiles                         types.Bool                      `tfsdk:"logging_to_files"`
	LoggingFilesInterval                   customtypes.Duration            `tfsdk:"logging_files_interval"`
	LoggingFilesKeepfiles                  types.Int32                     `tfsdk:"logging_files_keepfiles"`
	LoggingFilesRotateeverybytes           types.Int64                     `tfsdk:"logging_files_rotateeverybytes"`
	LoggingMetricsPeriod                   customtypes.Duration            `tfsdk:"logging_metrics_period"`
	GoMaxProcs                             types.Int32                     `tfsdk:"go_max_procs"`
	DownloadTimeout                        customtypes.Duration            `tfsdk:"download_timeout"`
	DownloadTargetDirectory                types.String                    `tfsdk:"download_target_directory"`
	MonitoringRuntimeExperimental          types.String                    `tfsdk:"monitoring_runtime_experimental"`
	FeaturesDisablePolicyChangeAcksEnabled types.Bool                      `tfsdk:"features_disable_policy_change_acks_enabled"`
	FeaturesIncludeTagsInEventsEnabled     types.Bool                      `tfsdk:"features_include_tags_in_events_enabled"`
	Internal                               customtypes.NormalizedYamlValue `tfsdk:"internal"`
}

func (m *agentPolicyModel) populateAdvancedSettingsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	attrTypes := objectAttrTypes(ctx, "advanced_settings")
	if data.AdvancedSettings == nil {
		m.AdvancedSettings = types.ObjectNull(attrTypes)
		return nil
	}

	settings := advancedSettingsModel{
		LoggingLevel:                           typeutils.StringFromAny(data.AdvancedSettings.AgentLoggingLevel),
		LoggingToFiles:                         typeutils.BoolFromAny(data.AdvancedSettings.AgentLoggingToFiles),
		LoggingFilesInterval:                   customtypes.NewDurationFromAny(data.AdvancedSettings.AgentLoggingFilesInterval),
		LoggingFilesKeepfiles:                  typeutils.Int32FromAnyFloat64(data.AdvancedSettings.AgentLoggingFilesKeepfiles),
		LoggingFilesRotateeverybytes:           typeutils.Int64FromAnyFloat64(data.AdvancedSettings.AgentLoggingFilesRotateeverybytes),
		LoggingMetricsPeriod:                   customtypes.NewDurationFromAny(data.AdvancedSettings.AgentLoggingMetricsPeriod),
		GoMaxProcs:                             typeutils.Int32FromAnyFloat64(data.AdvancedSettings.AgentLimitsGoMaxProcs),
		DownloadTimeout:                        customtypes.NewDurationFromAny(data.AdvancedSettings.AgentDownloadTimeout),
		DownloadTargetDirectory:                typeutils.StringFromAny(data.AdvancedSettings.AgentDownloadTargetDirectory),
		MonitoringRuntimeExperimental:          typeutils.StringFromAny(data.AdvancedSettings.AgentMonitoringRuntimeExperimental),
		FeaturesDisablePolicyChangeAcksEnabled: typeutils.BoolFromAny(data.AdvancedSettings.AgentFeaturesDisablePolicyChangeAcksEnabled),
		FeaturesIncludeTagsInEventsEnabled:     typeutils.BoolFromAny(data.AdvancedSettings.AgentFeaturesIncludeTagsInEventsEnabled),
		Internal:                               customtypes.NewNormalizedYamlNull(),
	}
	var diags diag.Diagnostics
	switch internal := data.AdvancedSettings.AgentInternal.(type) {
	case nil:
	case string:
		settings.Internal = customtypes.NewNormalizedYamlValue(internal)
	default:
		encoded, err := json.Marshal(internal)
		if err != nil {
			diags.AddWarning("Unable to read advanced_settings.internal", err.Error())
		} else {
			settings.Internal = customtypes.NewNormalizedYamlValue(string(encoded))
		}
	}

	obj, objDiags := types.ObjectValueFrom(ctx, attrTypes, settings)
	diags.Append(objDiags...)
	if diags.HasError() {
		return diags
	}
	m.AdvancedSettings = obj
	return diags
}

type advancedMonitoringOptionsModel struct {
	HTTPMonitoringEndpoint types.Object `tfsdk:"http_monitoring_endpoint"`
	Diagnostics            types.Object `tfsdk:"diagnostics"`
}

type httpMonitoringEndpointModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	Host          types.String `tfsdk:"host"`
	Port          types.Int32  `tfsdk:"port"`
	BufferEnabled types.Bool   `tfsdk:"buffer_enabled"`
	PprofEnabled  types.Bool   `tfsdk:"pprof_enabled"`
}

type diagnosticsModel struct {
	RateLimits   types.Object `tfsdk:"rate_limits"`
	FileUploader types.Object `tfsdk:"file_uploader"`
}

type rateLimitsModel struct {
	Interval customtypes.Duration `tfsdk:"interval"`
	Burst    types.Int32          `tfsdk:"burst"`
}

type fileUploaderModel struct {
	InitDuration    customtypes.Duration `tfsdk:"init_duration"`
	BackoffDuration customtypes.Duration `tfsdk:"backoff_duration"`
	MaxRetries      types.Int32          `tfsdk:"max_retries"`
}

func (m *agentPolicyModel) populateAdvancedMonitoringFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	hasHTTP := data.MonitoringHttp != nil || data.MonitoringPprofEnabled != nil
	hasDiagnostics := data.MonitoringDiagnostics != nil
	if !hasHTTP && !hasDiagnostics {
		m.AdvancedMonitoringOptions = types.ObjectNull(objectAttrTypes(ctx, "advanced_monitoring_options"))
		return nil
	}

	httpObj, diags := httpMonitoringEndpointFromAPI(ctx, data)
	if diags.HasError() {
		return diags
	}
	diagnosticsObj, diagDiags := diagnosticsFromAPI(ctx, data)
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

func httpMonitoringEndpointFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (types.Object, diag.Diagnostics) {
	attrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "http_monitoring_endpoint")
	if data.MonitoringHttp == nil && data.MonitoringPprofEnabled == nil {
		return types.ObjectNull(attrTypes), nil
	}

	endpoint := httpMonitoringEndpointModel{
		Enabled:       types.BoolNull(),
		Host:          types.StringNull(),
		Port:          types.Int32Null(),
		BufferEnabled: types.BoolNull(),
		PprofEnabled:  types.BoolNull(),
	}
	if data.MonitoringHttp != nil {
		endpoint.Enabled = types.BoolPointerValue(data.MonitoringHttp.Enabled)
		endpoint.Host = types.StringPointerValue(data.MonitoringHttp.Host)
		if data.MonitoringHttp.Port != nil {
			endpoint.Port = types.Int32Value(int32(*data.MonitoringHttp.Port))
		}
		if data.MonitoringHttp.Buffer != nil {
			endpoint.BufferEnabled = types.BoolPointerValue(data.MonitoringHttp.Buffer.Enabled)
		}
	}
	endpoint.PprofEnabled = types.BoolPointerValue(data.MonitoringPprofEnabled)
	return types.ObjectValueFrom(ctx, attrTypes, endpoint)
}

func diagnosticsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (types.Object, diag.Diagnostics) {
	attrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics")
	if data.MonitoringDiagnostics == nil {
		return types.ObjectNull(attrTypes), nil
	}

	rateLimits, diags := rateLimitsFromAPI(ctx, data)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), diags
	}
	uploader, uploaderDiags := fileUploaderFromAPI(ctx, data)
	diags.Append(uploaderDiags...)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), diags
	}
	return types.ObjectValueFrom(ctx, attrTypes, diagnosticsModel{
		RateLimits:   rateLimits,
		FileUploader: uploader,
	})
}

func rateLimitsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (types.Object, diag.Diagnostics) {
	attrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics", "rate_limits")
	if data.MonitoringDiagnostics.Limit == nil {
		return types.ObjectNull(attrTypes), nil
	}
	limits := rateLimitsModel{
		Interval: customtypes.NewDurationNull(),
		Burst:    types.Int32Null(),
	}
	if data.MonitoringDiagnostics.Limit.Interval != nil {
		limits.Interval = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Limit.Interval)
	}
	if data.MonitoringDiagnostics.Limit.Burst != nil {
		limits.Burst = types.Int32Value(int32(*data.MonitoringDiagnostics.Limit.Burst))
	}
	return types.ObjectValueFrom(ctx, attrTypes, limits)
}

func fileUploaderFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (types.Object, diag.Diagnostics) {
	attrTypes := objectAttrTypes(ctx, "advanced_monitoring_options", "diagnostics", "file_uploader")
	if data.MonitoringDiagnostics.Uploader == nil {
		return types.ObjectNull(attrTypes), nil
	}
	uploader := fileUploaderModel{
		InitDuration:    customtypes.NewDurationNull(),
		BackoffDuration: customtypes.NewDurationNull(),
		MaxRetries:      types.Int32Null(),
	}
	if data.MonitoringDiagnostics.Uploader.InitDur != nil {
		uploader.InitDuration = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Uploader.InitDur)
	}
	if data.MonitoringDiagnostics.Uploader.MaxDur != nil {
		uploader.BackoffDuration = customtypes.NewDurationValue(*data.MonitoringDiagnostics.Uploader.MaxDur)
	}
	if data.MonitoringDiagnostics.Uploader.MaxRetries != nil {
		uploader.MaxRetries = types.Int32Value(int32(*data.MonitoringDiagnostics.Uploader.MaxRetries))
	}
	return types.ObjectValueFrom(ctx, attrTypes, uploader)
}

func objectAttrTypes(ctx context.Context, names ...string) map[string]attr.Type {
	attribute := getDataSourceSchema(ctx).Attributes[names[0]]
	for _, name := range names[1:] {
		attribute = attribute.(dsschema.SingleNestedAttribute).Attributes[name]
	}
	return attribute.GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
