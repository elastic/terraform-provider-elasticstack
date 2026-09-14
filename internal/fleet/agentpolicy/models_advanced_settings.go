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

package agentpolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type advancedSettingsModel struct {
	LoggingLevel                  types.String         `tfsdk:"logging_level"`
	LoggingToFiles                types.Bool           `tfsdk:"logging_to_files"`
	LoggingFilesInterval          customtypes.Duration `tfsdk:"logging_files_interval"`
	LoggingFilesKeepfiles         types.Int32          `tfsdk:"logging_files_keepfiles"`
	LoggingFilesRotateeverybytes  types.Int64          `tfsdk:"logging_files_rotateeverybytes"`
	LoggingMetricsPeriod          customtypes.Duration `tfsdk:"logging_metrics_period"`
	GoMaxProcs                    types.Int32          `tfsdk:"go_max_procs"`
	DownloadTimeout               customtypes.Duration `tfsdk:"download_timeout"`
	DownloadTargetDirectory       types.String         `tfsdk:"download_target_directory"`
	MonitoringRuntimeExperimental types.String         `tfsdk:"monitoring_runtime_experimental"`
}

// advancedSettingsAttrTypes returns attribute types for advanced_settings pulled from the schema
func advancedSettingsAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["advanced_settings"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

type advancedSettingsAPIValues = struct {
	AgentDownloadTargetDirectory                any `json:"agent_download_target_directory,omitempty"`
	AgentDownloadTimeout                        any `json:"agent_download_timeout,omitempty"`
	AgentFeaturesDisablePolicyChangeAcksEnabled any `json:"agent_features_disable_policy_change_acks_enabled,omitempty"`
	AgentFeaturesIncludeTagsInEventsEnabled     any `json:"agent_features_include_tags_in_events_enabled,omitempty"`
	AgentInternal                               any `json:"agent_internal,omitempty"`
	AgentLimitsGoMaxProcs                       any `json:"agent_limits_go_max_procs,omitempty"`
	AgentLoggingFilesInterval                   any `json:"agent_logging_files_interval,omitempty"`
	AgentLoggingFilesKeepfiles                  any `json:"agent_logging_files_keepfiles,omitempty"`
	AgentLoggingFilesRotateeverybytes           any `json:"agent_logging_files_rotateeverybytes,omitempty"`
	AgentLoggingLevel                           any `json:"agent_logging_level,omitempty"`
	AgentLoggingMetricsPeriod                   any `json:"agent_logging_metrics_period,omitempty"`
	AgentLoggingToFiles                         any `json:"agent_logging_to_files,omitempty"`
	AgentMonitoringRuntimeExperimental          any `json:"agent_monitoring_runtime_experimental,omitempty"`
}

// populateAdvancedSettingsFromAPI populates the advanced settings from API response
func (model *agentPolicyModel) populateAdvancedSettingsFromAPI(ctx context.Context, data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) diag.Diagnostics {
	if data.AdvancedSettings == nil {
		model.AdvancedSettings = types.ObjectNull(advancedSettingsAttrTypes())
		return nil
	}

	settings := advancedSettingsModel{
		LoggingLevel:                  typeutils.StringFromAny(data.AdvancedSettings.AgentLoggingLevel),
		LoggingToFiles:                typeutils.BoolFromAny(data.AdvancedSettings.AgentLoggingToFiles),
		LoggingFilesInterval:          customtypes.NewDurationFromAny(data.AdvancedSettings.AgentLoggingFilesInterval),
		LoggingFilesKeepfiles:         typeutils.Int32FromAnyFloat64(data.AdvancedSettings.AgentLoggingFilesKeepfiles),
		LoggingFilesRotateeverybytes:  typeutils.Int64FromAnyFloat64(data.AdvancedSettings.AgentLoggingFilesRotateeverybytes),
		LoggingMetricsPeriod:          customtypes.NewDurationFromAny(data.AdvancedSettings.AgentLoggingMetricsPeriod),
		GoMaxProcs:                    typeutils.Int32FromAnyFloat64(data.AdvancedSettings.AgentLimitsGoMaxProcs),
		DownloadTimeout:               customtypes.NewDurationFromAny(data.AdvancedSettings.AgentDownloadTimeout),
		DownloadTargetDirectory:       typeutils.StringFromAny(data.AdvancedSettings.AgentDownloadTargetDirectory),
		MonitoringRuntimeExperimental: typeutils.StringFromAny(data.AdvancedSettings.AgentMonitoringRuntimeExperimental),
	}

	obj, diags := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), settings)
	if diags.HasError() {
		return diags
	}
	model.AdvancedSettings = obj
	return nil
}

// convertAdvancedSettingsToAPI converts the advanced settings config to API format
func (model *agentPolicyModel) convertAdvancedSettingsToAPI(ctx context.Context, feat agentPolicyFeatures) (*advancedSettingsAPIValues, diag.Diagnostics) {
	if !typeutils.IsKnown(model.AdvancedSettings) {
		return nil, nil
	}

	var settings advancedSettingsModel
	diags := model.AdvancedSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	// Check if any values are set
	hasValues := typeutils.IsKnown(settings.LoggingLevel) ||
		typeutils.IsKnown(settings.LoggingToFiles) ||
		typeutils.IsKnown(settings.LoggingFilesInterval) ||
		typeutils.IsKnown(settings.LoggingFilesKeepfiles) ||
		typeutils.IsKnown(settings.LoggingFilesRotateeverybytes) ||
		typeutils.IsKnown(settings.LoggingMetricsPeriod) ||
		typeutils.IsKnown(settings.GoMaxProcs) ||
		typeutils.IsKnown(settings.DownloadTimeout) ||
		typeutils.IsKnown(settings.DownloadTargetDirectory) ||
		typeutils.IsKnown(settings.MonitoringRuntimeExperimental)

	if !hasValues {
		return nil, nil
	}

	result := &advancedSettingsAPIValues{}

	if typeutils.IsKnown(settings.LoggingLevel) {
		result.AgentLoggingLevel = settings.LoggingLevel.ValueString()
	}
	if typeutils.IsKnown(settings.LoggingToFiles) {
		result.AgentLoggingToFiles = settings.LoggingToFiles.ValueBool()
	}
	if typeutils.IsKnown(settings.LoggingFilesInterval) {
		result.AgentLoggingFilesInterval = settings.LoggingFilesInterval.ValueString()
	}
	if typeutils.IsKnown(settings.LoggingFilesKeepfiles) {
		result.AgentLoggingFilesKeepfiles = settings.LoggingFilesKeepfiles.ValueInt32()
	}
	if typeutils.IsKnown(settings.LoggingFilesRotateeverybytes) {
		result.AgentLoggingFilesRotateeverybytes = settings.LoggingFilesRotateeverybytes.ValueInt64()
	}
	if typeutils.IsKnown(settings.LoggingMetricsPeriod) {
		result.AgentLoggingMetricsPeriod = settings.LoggingMetricsPeriod.ValueString()
	}
	if typeutils.IsKnown(settings.GoMaxProcs) {
		result.AgentLimitsGoMaxProcs = settings.GoMaxProcs.ValueInt32()
	}
	if typeutils.IsKnown(settings.DownloadTimeout) {
		result.AgentDownloadTimeout = settings.DownloadTimeout.ValueString()
	}
	if typeutils.IsKnown(settings.DownloadTargetDirectory) {
		result.AgentDownloadTargetDirectory = settings.DownloadTargetDirectory.ValueString()
	}
	if typeutils.IsKnown(settings.MonitoringRuntimeExperimental) && feat.SupportsMonitoringRuntimeExperimental {
		result.AgentMonitoringRuntimeExperimental = settings.MonitoringRuntimeExperimental.ValueString()
	}

	return result, diags
}
