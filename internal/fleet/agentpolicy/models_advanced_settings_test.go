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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateAdvancedSettingsFromAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("nil advanced_settings sets object null", func(t *testing.T) {
		model := &agentPolicyModel{}
		diags := model.populateAdvancedSettingsFromAPI(ctx, &kbapi.KibanaHTTPAPIsAgentPolicyResponse{})
		require.False(t, diags.HasError())
		assert.True(t, model.AdvancedSettings.IsNull())
	})

	t.Run("all fields populated from API values", func(t *testing.T) {
		model := &agentPolicyModel{}
		data := &kbapi.KibanaHTTPAPIsAgentPolicyResponse{
			AdvancedSettings: &struct {
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
			}{
				AgentLoggingLevel:                  "debug",
				AgentLoggingToFiles:                true,
				AgentLoggingFilesInterval:          "30s",
				AgentLoggingFilesKeepfiles:         float64(7),
				AgentLoggingFilesRotateeverybytes:  float64(10485760),
				AgentLoggingMetricsPeriod:          "1m",
				AgentLimitsGoMaxProcs:              float64(2),
				AgentDownloadTimeout:               "2h",
				AgentDownloadTargetDirectory:       "/tmp/elastic",
				AgentMonitoringRuntimeExperimental: "enabled",
			},
		}

		diags := model.populateAdvancedSettingsFromAPI(ctx, data)
		require.False(t, diags.HasError())
		require.False(t, model.AdvancedSettings.IsNull())

		var settings advancedSettingsModel
		diags = model.AdvancedSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())

		assert.Equal(t, "debug", settings.LoggingLevel.ValueString())
		assert.True(t, settings.LoggingToFiles.ValueBool())
		assert.Equal(t, "30s", settings.LoggingFilesInterval.ValueString())
		assert.Equal(t, int32(7), settings.LoggingFilesKeepfiles.ValueInt32())
		assert.Equal(t, int64(10485760), settings.LoggingFilesRotateeverybytes.ValueInt64())
		assert.Equal(t, "1m", settings.LoggingMetricsPeriod.ValueString())
		assert.Equal(t, int32(2), settings.GoMaxProcs.ValueInt32())
		assert.Equal(t, "2h", settings.DownloadTimeout.ValueString())
		assert.Equal(t, "/tmp/elastic", settings.DownloadTargetDirectory.ValueString())
		assert.Equal(t, "enabled", settings.MonitoringRuntimeExperimental.ValueString())
	})

	t.Run("fields with wrong types fall back to null", func(t *testing.T) {
		model := &agentPolicyModel{}
		data := &kbapi.KibanaHTTPAPIsAgentPolicyResponse{
			AdvancedSettings: &struct {
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
			}{
				AgentLoggingLevel:   42,
				AgentLoggingToFiles: "not-a-bool",
			},
		}

		diags := model.populateAdvancedSettingsFromAPI(ctx, data)
		require.False(t, diags.HasError())

		var settings advancedSettingsModel
		diags = model.AdvancedSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())

		assert.True(t, settings.LoggingLevel.IsNull())
		assert.True(t, settings.LoggingToFiles.IsNull())
	})
}

func TestConvertAdvancedSettingsToAPI(t *testing.T) {
	ctx := context.Background()

	createAdvancedSettingsObject := func(settings advancedSettingsModel) types.Object {
		obj, _ := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), settings)
		return obj
	}

	tests := []struct {
		name             string
		advancedSettings types.Object
		wantNil          bool
		checkResult      func(t *testing.T, result *advancedSettingsAPIValues)
	}{
		{
			name:             "null advanced_settings returns nil",
			advancedSettings: types.ObjectNull(advancedSettingsAttrTypes()),
			wantNil:          true,
		},
		{
			name: "all null values returns nil",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringNull(),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Null(),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.StringNull(),
			}),
			wantNil: true,
		},
		{
			name: "logging_level set returns value",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringValue("debug"),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Null(),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.StringNull(),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIValues) {
				assert.Equal(t, "debug", result.AgentLoggingLevel)
				assert.Nil(t, result.AgentLoggingToFiles)
			},
		},
		{
			name: "go_max_procs set returns value",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringNull(),
				LoggingToFiles:                types.BoolNull(),
				LoggingFilesInterval:          customtypes.NewDurationNull(),
				LoggingFilesKeepfiles:         types.Int32Null(),
				LoggingFilesRotateeverybytes:  types.Int64Null(),
				LoggingMetricsPeriod:          customtypes.NewDurationNull(),
				GoMaxProcs:                    types.Int32Value(4),
				DownloadTimeout:               customtypes.NewDurationNull(),
				DownloadTargetDirectory:       types.StringNull(),
				MonitoringRuntimeExperimental: types.StringNull(),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIValues) {
				assert.Equal(t, int32(4), result.AgentLimitsGoMaxProcs)
			},
		},
		{
			name: "multiple values set returns all values",
			advancedSettings: createAdvancedSettingsObject(advancedSettingsModel{
				LoggingLevel:                  types.StringValue("info"),
				LoggingToFiles:                types.BoolValue(true),
				LoggingFilesInterval:          customtypes.NewDurationValue("30s"),
				LoggingFilesKeepfiles:         types.Int32Value(7),
				LoggingFilesRotateeverybytes:  types.Int64Value(10485760),
				LoggingMetricsPeriod:          customtypes.NewDurationValue("1m"),
				GoMaxProcs:                    types.Int32Value(2),
				DownloadTimeout:               customtypes.NewDurationValue("2h"),
				DownloadTargetDirectory:       types.StringValue("/tmp/elastic"),
				MonitoringRuntimeExperimental: types.StringValue(""),
			}),
			wantNil: false,
			checkResult: func(t *testing.T, result *advancedSettingsAPIValues) {
				assert.Equal(t, "info", result.AgentLoggingLevel)
				assert.Equal(t, true, result.AgentLoggingToFiles)
				assert.Equal(t, "30s", result.AgentLoggingFilesInterval)
				assert.Equal(t, int32(7), result.AgentLoggingFilesKeepfiles)
				assert.Equal(t, int64(10485760), result.AgentLoggingFilesRotateeverybytes)
				assert.Equal(t, "1m", result.AgentLoggingMetricsPeriod)
				assert.Equal(t, int32(2), result.AgentLimitsGoMaxProcs)
				assert.Equal(t, "2h", result.AgentDownloadTimeout)
				assert.Equal(t, "/tmp/elastic", result.AgentDownloadTargetDirectory)
				assert.Empty(t, result.AgentMonitoringRuntimeExperimental)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				AdvancedSettings: tt.advancedSettings,
			}

			got, diags := model.convertAdvancedSettingsToAPI(ctx, agentPolicyFeatures{SupportsMonitoringRuntimeExperimental: true})
			assert.False(t, diags.HasError())

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}
