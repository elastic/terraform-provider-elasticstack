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

package agentpolicyshape_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateAdvancedSettings(t *testing.T) {
	t.Run("nil advanced settings returns ok=false", func(t *testing.T) {
		settings, ok, diags := agentpolicyshape.PopulateAdvancedSettings(&kbapi.KibanaHTTPAPIsAgentPolicyResponse{})
		require.False(t, diags.HasError())
		assert.False(t, ok)
		assert.Equal(t, agentpolicyshape.AdvancedSettingsModel{}, settings)
	})

	t.Run("populates fields from API values", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"advanced_settings":{
			"agent_logging_level": "debug",
			"agent_logging_to_files": true,
			"agent_logging_files_interval": "30s",
			"agent_logging_files_keepfiles": 7,
			"agent_logging_files_rotateeverybytes": 10485760,
			"agent_logging_metrics_period": "1m",
			"agent_limits_go_max_procs": 2,
			"agent_download_timeout": "2h",
			"agent_download_target_directory": "/tmp/elastic",
			"agent_monitoring_runtime_experimental": "otel",
			"agent_features_disable_policy_change_acks_enabled": true,
			"agent_features_include_tags_in_events_enabled": false,
			"agent_internal": "a:\n  b: 1\n"
		}}`), &data))

		settings, ok, diags := agentpolicyshape.PopulateAdvancedSettings(&data)
		require.False(t, diags.HasError())
		require.True(t, ok)

		assert.Equal(t, "debug", settings.LoggingLevel.ValueString())
		assert.True(t, settings.LoggingToFiles.ValueBool())
		assert.Equal(t, "30s", settings.LoggingFilesInterval.ValueString())
		assert.Equal(t, int32(7), settings.LoggingFilesKeepfiles.ValueInt32())
		assert.Equal(t, int64(10485760), settings.LoggingFilesRotateeverybytes.ValueInt64())
		assert.Equal(t, "1m", settings.LoggingMetricsPeriod.ValueString())
		assert.Equal(t, int32(2), settings.GoMaxProcs.ValueInt32())
		assert.Equal(t, "2h", settings.DownloadTimeout.ValueString())
		assert.Equal(t, "/tmp/elastic", settings.DownloadTargetDirectory.ValueString())
		assert.Equal(t, "otel", settings.MonitoringRuntimeExperimental.ValueString())
		assert.True(t, settings.FeaturesDisablePolicyChangeAcksEnabled.ValueBool())
		assert.False(t, settings.FeaturesIncludeTagsInEventsEnabled.ValueBool())
		assert.Equal(t, "a:\n  b: 1\n", settings.Internal.ValueString())
	})

	t.Run("structured internal YAML is re-encoded as JSON", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"advanced_settings":{"agent_internal":{"agent":{"logging":{"level":"debug"}}}}}`), &data))

		settings, ok, diags := agentpolicyshape.PopulateAdvancedSettings(&data)
		require.False(t, diags.HasError())
		require.True(t, ok)
		assert.JSONEq(t, `{"agent":{"logging":{"level":"debug"}}}`, settings.Internal.ValueString())
	})
}

func TestPopulateHTTPMonitoringEndpoint(t *testing.T) {
	defaults := agentpolicyshape.HTTPMonitoringEndpointModel{
		Enabled:       types.BoolValue(false),
		Host:          types.StringValue("localhost"),
		Port:          types.Int32Value(6791),
		BufferEnabled: types.BoolValue(false),
		PprofEnabled:  types.BoolValue(false),
	}

	t.Run("no http monitoring data returns ok=false", func(t *testing.T) {
		_, ok := agentpolicyshape.PopulateHTTPMonitoringEndpoint(&kbapi.KibanaHTTPAPIsAgentPolicyResponse{}, defaults)
		assert.False(t, ok)
	})

	t.Run("present but empty section falls back to defaults", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_http":{}}`), &data))

		endpoint, ok := agentpolicyshape.PopulateHTTPMonitoringEndpoint(&data, defaults)
		require.True(t, ok)
		assert.Equal(t, defaults, endpoint)
	})

	t.Run("present but empty section falls back to nulls for zero-value defaults", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_http":{}}`), &data))

		endpoint, ok := agentpolicyshape.PopulateHTTPMonitoringEndpoint(&data, agentpolicyshape.HTTPMonitoringEndpointModel{})
		require.True(t, ok)
		assert.True(t, endpoint.Enabled.IsNull())
		assert.True(t, endpoint.Host.IsNull())
		assert.True(t, endpoint.Port.IsNull())
		assert.True(t, endpoint.BufferEnabled.IsNull())
		assert.True(t, endpoint.PprofEnabled.IsNull())
	})

	t.Run("explicit values override defaults", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_http":{
			"enabled": true,
			"host": "0.0.0.0",
			"port": 8080,
			"buffer": {"enabled": true}
		}, "monitoring_pprof_enabled": true}`), &data))

		endpoint, ok := agentpolicyshape.PopulateHTTPMonitoringEndpoint(&data, defaults)
		require.True(t, ok)
		assert.True(t, endpoint.Enabled.ValueBool())
		assert.Equal(t, "0.0.0.0", endpoint.Host.ValueString())
		assert.Equal(t, int32(8080), endpoint.Port.ValueInt32())
		assert.True(t, endpoint.BufferEnabled.ValueBool())
		assert.True(t, endpoint.PprofEnabled.ValueBool())
	})

	t.Run("pprof_enabled alone is sufficient to report present", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_pprof_enabled": true}`), &data))

		endpoint, ok := agentpolicyshape.PopulateHTTPMonitoringEndpoint(&data, defaults)
		require.True(t, ok)
		assert.True(t, endpoint.PprofEnabled.ValueBool())
		// Fields not covered by monitoring_http stay at defaults.
		assert.Equal(t, defaults.Host, endpoint.Host)
	})
}

func TestPopulateRateLimits(t *testing.T) {
	defaults := agentpolicyshape.RateLimitsModel{
		Interval: customtypes.NewDurationValue("1m"),
		Burst:    types.Int32Value(1),
	}

	t.Run("no diagnostics returns ok=false", func(t *testing.T) {
		_, ok := agentpolicyshape.PopulateRateLimits(&kbapi.KibanaHTTPAPIsAgentPolicyResponse{}, defaults)
		assert.False(t, ok)
	})

	t.Run("no rate limiting configuration returns ok=false", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{}}`), &data))

		_, ok := agentpolicyshape.PopulateRateLimits(&data, defaults)
		assert.False(t, ok)
	})

	t.Run("explicit values override defaults", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{"limit":{"interval":"2m","burst":5}}}`), &data))

		rateLimits, ok := agentpolicyshape.PopulateRateLimits(&data, defaults)
		require.True(t, ok)
		assert.Equal(t, "2m", rateLimits.Interval.ValueString())
		assert.Equal(t, int32(5), rateLimits.Burst.ValueInt32())
	})

	t.Run("partial values fall back to defaults for the rest", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{"limit":{"burst":9}}}`), &data))

		rateLimits, ok := agentpolicyshape.PopulateRateLimits(&data, defaults)
		require.True(t, ok)
		assert.Equal(t, defaults.Interval, rateLimits.Interval)
		assert.Equal(t, int32(9), rateLimits.Burst.ValueInt32())
	})
}

func TestPopulateFileUploader(t *testing.T) {
	defaults := agentpolicyshape.FileUploaderModel{
		InitDuration:    customtypes.NewDurationValue("1s"),
		BackoffDuration: customtypes.NewDurationValue("1m"),
		MaxRetries:      types.Int32Value(10),
	}

	t.Run("no diagnostics returns ok=false", func(t *testing.T) {
		_, ok := agentpolicyshape.PopulateFileUploader(&kbapi.KibanaHTTPAPIsAgentPolicyResponse{}, defaults)
		assert.False(t, ok)
	})

	t.Run("no uploader configuration returns ok=false", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{}}`), &data))

		_, ok := agentpolicyshape.PopulateFileUploader(&data, defaults)
		assert.False(t, ok)
	})

	t.Run("explicit values override defaults", func(t *testing.T) {
		var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
		require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{"uploader":{"init_dur":"2s","max_dur":"2m","max_retries":20}}}`), &data))

		uploader, ok := agentpolicyshape.PopulateFileUploader(&data, defaults)
		require.True(t, ok)
		assert.Equal(t, "2s", uploader.InitDuration.ValueString())
		assert.Equal(t, "2m", uploader.BackoffDuration.ValueString())
		assert.Equal(t, int32(20), uploader.MaxRetries.ValueInt32())
	})
}

func TestHasDiagnostics(t *testing.T) {
	assert.False(t, agentpolicyshape.HasDiagnostics(&kbapi.KibanaHTTPAPIsAgentPolicyResponse{}))

	var data kbapi.KibanaHTTPAPIsAgentPolicyResponse
	require.NoError(t, json.Unmarshal([]byte(`{"monitoring_diagnostics":{}}`), &data))
	assert.True(t, agentpolicyshape.HasDiagnostics(&data))
}
