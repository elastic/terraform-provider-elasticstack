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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (model agentPolicyModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement

	addMinimum := func(configured bool, attr path.Path, minVersion *version.Version, detail string) {
		if !configured {
			return
		}
		reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
			attr,
			*minVersion,
			detail,
		))
	}

	addMinimum(
		len(model.GlobalDataTags.Elements()) > 0,
		path.Root("global_data_tags"),
		MinVersionGlobalDataTags,
		fmt.Sprintf("Global data tags are only supported in Elastic Stack %s and above", MinVersionGlobalDataTags),
	)
	addMinimum(
		typeutils.IsKnown(model.RequiredVersions),
		path.Root("required_versions"),
		MinVersionRequiredVersions,
		fmt.Sprintf("Required versions (automatic agent upgrades) are only supported in Elastic Stack %s and above", MinVersionRequiredVersions),
	)
	addMinimum(
		typeutils.IsKnown(model.HostNameFormat) && model.HostNameFormat.ValueString() == HostNameFormatFQDN,
		path.Root("host_name_format"),
		MinVersionAgentFeatures,
		fmt.Sprintf("host_name_format (agent_features) is only supported in Elastic Stack %s and above", MinVersionAgentFeatures),
	)
	addMinimum(
		typeutils.IsKnown(model.AdvancedSettings),
		path.Root("advanced_settings"),
		MinVersionAdvancedSettings,
		fmt.Sprintf("Advanced settings are only supported in Elastic Stack %s and above", MinVersionAdvancedSettings),
	)
	addMinimum(
		typeutils.IsKnown(model.AdvancedMonitoringOptions),
		path.Root("advanced_monitoring_options"),
		MinVersionAdvancedMonitoring,
		fmt.Sprintf("Advanced monitoring options are only supported in Elastic Stack %s and above", MinVersionAdvancedMonitoring),
	)
	addMinimum(
		typeutils.IsKnown(model.IsProtected) && model.IsProtected.ValueBool(),
		path.Root("is_protected"),
		MinVersionTamperProtection,
		fmt.Sprintf("Tamper protection (`is_protected`) is only supported in Elastic Stack %s and above", MinVersionTamperProtection),
	)
	addMinimum(
		typeutils.IsKnown(model.SupportsAgentless),
		path.Root("supports_agentless"),
		MinSupportsAgentlessVersion,
		fmt.Sprintf("Supports agentless is only supported in Elastic Stack %s and above", MinSupportsAgentlessVersion),
	)
	addMinimum(
		typeutils.IsKnown(model.InactivityTimeout),
		path.Root("inactivity_timeout"),
		MinVersionInactivityTimeout,
		fmt.Sprintf("Inactivity timeout is only supported in Elastic Stack %s and above", MinVersionInactivityTimeout),
	)
	addMinimum(
		typeutils.IsKnown(model.UnenrollmentTimeout),
		path.Root("unenrollment_timeout"),
		MinVersionUnenrollmentTimeout,
		fmt.Sprintf("Unenrollment timeout is only supported in Elastic Stack %s and above", MinVersionUnenrollmentTimeout),
	)
	addMinimum(
		typeutils.IsKnown(model.SpaceIDs),
		path.Root("space_ids"),
		MinVersionSpaceIDs,
		fmt.Sprintf("Space IDs are only supported in Elastic Stack %s and above", MinVersionSpaceIDs),
	)

	if !typeutils.IsKnown(model.AdvancedSettings) {
		return reqs, nil
	}

	var settings advancedSettingsModel
	diags := model.AdvancedSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	if typeutils.IsKnown(settings.MonitoringRuntimeExperimental) {
		reqs = append(reqs, entitycore.NewAttributeVersionCheckRequirement(
			path.Root("advanced_settings").AtName("monitoring_runtime_experimental"),
			MonitoringRuntimeExperimentalSupported,
			"monitoring_runtime_experimental is only supported in Elastic Stack 8.19.x or 9.1.0 and above",
		))
	}

	return reqs, diags
}
