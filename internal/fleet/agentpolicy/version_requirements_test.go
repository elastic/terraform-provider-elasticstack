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

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentpolicyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAgentPolicyVersionRequirements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	advancedSettings, diags := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), agentpolicyshape.AdvancedSettingsModel{
		LoggingLevel:                           types.StringValue("debug"),
		MonitoringRuntimeExperimental:          types.StringValue("process"),
		FeaturesDisablePolicyChangeAcksEnabled: types.BoolValue(false),
		FeaturesIncludeTagsInEventsEnabled:     types.BoolValue(false),
		Internal:                               customtypes.NewNormalizedYamlValue("a: 1"),
	})
	require.False(t, diags.HasError())

	model := agentPolicyModel{
		GlobalDataTags:            types.MapValueMust(types.StringType, map[string]attr.Value{"environment": types.StringValue("production")}),
		RequiredVersions:          types.MapValueMust(types.Int32Type, map[string]attr.Value{}),
		HostNameFormat:            types.StringValue(HostNameFormatFQDN),
		AdvancedSettings:          advancedSettings,
		AdvancedMonitoringOptions: types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		IsProtected:               types.BoolValue(true),
		SupportsAgentless:         types.BoolValue(true),
		InactivityTimeout:         customtypes.NewDurationValue("30s"),
		UnenrollmentTimeout:       customtypes.NewDurationValue("60s"),
		SpaceIDs:                  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
	}

	reqs, reqDiags := model.GetVersionRequirements(ctx)
	require.False(t, reqDiags.HasError())
	require.Len(t, reqs, 14)

	expected := []struct {
		path       string
		minVersion string
	}{
		{"global_data_tags", MinVersionGlobalDataTags.String()},
		{"required_versions", MinVersionRequiredVersions.String()},
		{"host_name_format", MinVersionAgentFeatures.String()},
		{"advanced_settings", MinVersionAdvancedSettings.String()},
		{"advanced_monitoring_options", MinVersionAdvancedMonitoring.String()},
		{"is_protected", MinVersionTamperProtection.String()},
		{"supports_agentless", MinSupportsAgentlessVersion.String()},
		{"inactivity_timeout", MinVersionInactivityTimeout.String()},
		{"unenrollment_timeout", MinVersionUnenrollmentTimeout.String()},
		{"space_ids", MinVersionSpaceIDs.String()},
		{"advanced_settings.monitoring_runtime_experimental", ""},
		{"advanced_settings.features_disable_policy_change_acks_enabled", MinVersionFeaturesDisablePolicyChangeAcksEnabled.String()},
		{"advanced_settings.features_include_tags_in_events_enabled", MinVersionFeaturesIncludeTagsInEventsEnabled.String()},
		{"advanced_settings.internal", MinVersionAgentInternal.String()},
	}
	for i, expected := range expected {
		require.NotNil(t, reqs[i].AttributePath)
		require.Equal(t, expected.path, reqs[i].AttributePath.String())
		if expected.minVersion != "" {
			require.Equal(t, expected.minVersion, reqs[i].MinVersion.String())
		}
	}
	for _, req := range reqs {
		if req.AttributePath.String() == "advanced_settings.monitoring_runtime_experimental" {
			require.NotNil(t, req.VersionCheck)
		}
	}
}

func TestAgentPolicyVersionRequirementsAgentFeaturesUnset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	advancedSettings, diags := types.ObjectValueFrom(ctx, advancedSettingsAttrTypes(), agentpolicyshape.AdvancedSettingsModel{
		LoggingLevel: types.StringValue("debug"),
	})
	require.False(t, diags.HasError())

	reqs, reqDiags := agentPolicyModel{AdvancedSettings: advancedSettings}.GetVersionRequirements(ctx)
	require.False(t, reqDiags.HasError())
	for _, req := range reqs {
		require.NotNil(t, req.AttributePath)
		require.NotContains(t, req.AttributePath.String(), "features_")
		require.NotEqual(t, "advanced_settings.internal", req.AttributePath.String())
	}
}

func TestAgentPolicyVersionRequirementsIgnoreCompatibleDefaults(t *testing.T) {
	t.Parallel()

	model := agentPolicyModel{
		GlobalDataTags:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
		HostNameFormat:    types.StringValue(HostNameFormatHostname),
		IsProtected:       types.BoolValue(false),
		SupportsAgentless: types.BoolNull(),
	}

	reqs, diags := model.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Empty(t, reqs)
}
