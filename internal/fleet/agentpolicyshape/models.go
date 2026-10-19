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

// Package agentpolicyshape holds the advanced_settings/advanced_monitoring_options
// model structs and Fleet API population logic shared between the
// elasticstack_fleet_agent_policy resource and its data source, so both stay
// in lockstep with the Kibana API shape instead of maintaining two
// independent copies.
package agentpolicyshape

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdvancedSettingsModel is the advanced_settings nested attribute model.
type AdvancedSettingsModel struct {
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

// AdvancedMonitoringOptionsModel is the advanced_monitoring_options nested
// attribute model.
type AdvancedMonitoringOptionsModel struct {
	HTTPMonitoringEndpoint types.Object `tfsdk:"http_monitoring_endpoint"`
	Diagnostics            types.Object `tfsdk:"diagnostics"`
}

// HTTPMonitoringEndpointModel is the advanced_monitoring_options.http_monitoring_endpoint model.
type HTTPMonitoringEndpointModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	Host          types.String `tfsdk:"host"`
	Port          types.Int32  `tfsdk:"port"`
	BufferEnabled types.Bool   `tfsdk:"buffer_enabled"`
	PprofEnabled  types.Bool   `tfsdk:"pprof_enabled"`
}

// DiagnosticsModel is the advanced_monitoring_options.diagnostics model.
type DiagnosticsModel struct {
	RateLimits   types.Object `tfsdk:"rate_limits"`
	FileUploader types.Object `tfsdk:"file_uploader"`
}

// RateLimitsModel is the advanced_monitoring_options.diagnostics.rate_limits model.
type RateLimitsModel struct {
	Interval customtypes.Duration `tfsdk:"interval"`
	Burst    types.Int32          `tfsdk:"burst"`
}

// FileUploaderModel is the advanced_monitoring_options.diagnostics.file_uploader model.
type FileUploaderModel struct {
	InitDuration    customtypes.Duration `tfsdk:"init_duration"`
	BackoffDuration customtypes.Duration `tfsdk:"backoff_duration"`
	MaxRetries      types.Int32          `tfsdk:"max_retries"`
}

// PopulateAdvancedSettings builds the advanced_settings model from the Fleet
// API response. ok is false when the API reported no advanced settings at
// all, in which case the caller should store a null object instead.
func PopulateAdvancedSettings(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) (settings AdvancedSettingsModel, ok bool, diags diag.Diagnostics) {
	if data.AdvancedSettings == nil {
		return AdvancedSettingsModel{}, false, nil
	}

	settings = AdvancedSettingsModel{
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

		FeaturesDisablePolicyChangeAcksEnabled: typeutils.BoolFromAny(data.AdvancedSettings.AgentFeaturesDisablePolicyChangeAcksEnabled),
		FeaturesIncludeTagsInEventsEnabled:     typeutils.BoolFromAny(data.AdvancedSettings.AgentFeaturesIncludeTagsInEventsEnabled),
		Internal:                               customtypes.NewNormalizedYamlNull(),
	}

	switch internal := data.AdvancedSettings.AgentInternal.(type) {
	case nil:
	case string:
		settings.Internal = customtypes.NewNormalizedYamlValue(internal)
	default:
		// Kibana may return structured YAML as a JSON object; JSON is valid YAML.
		if encoded, err := json.Marshal(internal); err == nil {
			settings.Internal = customtypes.NewNormalizedYamlValue(string(encoded))
		} else {
			diags.AddWarning("Unable to read advanced_settings.internal", err.Error())
		}
	}

	return settings, true, diags
}

// PopulateHTTPMonitoringEndpoint builds the http_monitoring_endpoint model
// from the Fleet API response. It starts from defaults and overlays any
// field the API explicitly returned, so a zero-value defaults argument
// yields nulls for computed-only (data source) callers, while a resource
// caller can pass its schema's static defaults to report a stable value
// when the API reports the section exists but omits a specific field. ok is
// false when the API reported neither the endpoint nor pprof_enabled.
func PopulateHTTPMonitoringEndpoint(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse, defaults HTTPMonitoringEndpointModel) (endpoint HTTPMonitoringEndpointModel, ok bool) {
	hasHTTPMonitoring := data.MonitoringHttp != nil
	hasPprofEnabled := data.MonitoringPprofEnabled != nil
	if !hasHTTPMonitoring && !hasPprofEnabled {
		return HTTPMonitoringEndpointModel{}, false
	}

	endpoint = defaults
	if data.MonitoringHttp != nil {
		if data.MonitoringHttp.Enabled != nil {
			endpoint.Enabled = types.BoolValue(*data.MonitoringHttp.Enabled)
		}
		if data.MonitoringHttp.Host != nil {
			endpoint.Host = types.StringValue(*data.MonitoringHttp.Host)
		}
		if data.MonitoringHttp.Port != nil {
			endpoint.Port = types.Int32Value(int32(*data.MonitoringHttp.Port))
		}
		if data.MonitoringHttp.Buffer != nil && data.MonitoringHttp.Buffer.Enabled != nil {
			endpoint.BufferEnabled = types.BoolValue(*data.MonitoringHttp.Buffer.Enabled)
		}
	}
	if data.MonitoringPprofEnabled != nil {
		endpoint.PprofEnabled = types.BoolValue(*data.MonitoringPprofEnabled)
	}

	return endpoint, true
}

// PopulateRateLimits builds the diagnostics.rate_limits model from the Fleet
// API response, overlaying defaults the same way PopulateHTTPMonitoringEndpoint
// does. ok is false when the API reported no rate limiting configuration.
func PopulateRateLimits(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse, defaults RateLimitsModel) (RateLimitsModel, bool) {
	if data.MonitoringDiagnostics == nil || data.MonitoringDiagnostics.Limit == nil {
		return RateLimitsModel{}, false
	}

	limit := data.MonitoringDiagnostics.Limit
	rateLimits := defaults
	if limit.Interval != nil {
		rateLimits.Interval = customtypes.NewDurationValue(*limit.Interval)
	}
	if limit.Burst != nil {
		rateLimits.Burst = types.Int32Value(int32(*limit.Burst))
	}

	return rateLimits, true
}

// PopulateFileUploader builds the diagnostics.file_uploader model from the
// Fleet API response, overlaying defaults the same way
// PopulateHTTPMonitoringEndpoint does. ok is false when the API reported no
// file uploader configuration.
func PopulateFileUploader(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse, defaults FileUploaderModel) (FileUploaderModel, bool) {
	if data.MonitoringDiagnostics == nil || data.MonitoringDiagnostics.Uploader == nil {
		return FileUploaderModel{}, false
	}

	uploader := data.MonitoringDiagnostics.Uploader
	fileUploader := defaults
	if uploader.InitDur != nil {
		fileUploader.InitDuration = customtypes.NewDurationValue(*uploader.InitDur)
	}
	if uploader.MaxDur != nil {
		fileUploader.BackoffDuration = customtypes.NewDurationValue(*uploader.MaxDur)
	}
	if uploader.MaxRetries != nil {
		fileUploader.MaxRetries = types.Int32Value(int32(*uploader.MaxRetries))
	}

	return fileUploader, true
}

// HasDiagnostics reports whether the Fleet API response carries any
// diagnostics configuration at all.
func HasDiagnostics(data *kbapi.KibanaHTTPAPIsAgentPolicyResponse) bool {
	return data.MonitoringDiagnostics != nil
}
