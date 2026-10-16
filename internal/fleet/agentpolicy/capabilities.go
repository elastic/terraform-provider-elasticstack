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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// agentPolicyFeatures records which Kibana/Fleet features the connected stack supports.
type agentPolicyFeatures struct {
	SupportsGlobalDataTags                bool
	SupportsSupportsAgentless             bool
	SupportsInactivityTimeout             bool
	SupportsUnenrollmentTimeout           bool
	SupportsSpaceIDs                      bool
	SupportsRequiredVersions              bool
	SupportsAgentFeatures                 bool
	SupportsAdvancedMonitoring            bool
	SupportsAdvancedSettings              bool
	SupportsMonitoringRuntimeExperimental bool
	SupportsTamperProtection              bool

	SupportsAgentInternal                          bool
	SupportsFeaturesDisablePolicyChangeAcksEnabled bool
	SupportsFeaturesIncludeTagsInEventsEnabled     bool
}

func resolveAgentPolicyFeatures(ctx context.Context, client *clients.KibanaScopedClient) (agentPolicyFeatures, diag.Diagnostics) {
	var diags diag.Diagnostics
	var f agentPolicyFeatures

	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsGlobalDataTags, &diags, MinVersionGlobalDataTags)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsSupportsAgentless, &diags, MinSupportsAgentlessVersion)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsInactivityTimeout, &diags, MinVersionInactivityTimeout)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsUnenrollmentTimeout, &diags, MinVersionUnenrollmentTimeout)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsSpaceIDs, &diags, MinVersionSpaceIDs)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsRequiredVersions, &diags, MinVersionRequiredVersions)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsAgentFeatures, &diags, MinVersionAgentFeatures)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsAdvancedMonitoring, &diags, MinVersionAdvancedMonitoring)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsAdvancedSettings, &diags, MinVersionAdvancedSettings)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsTamperProtection, &diags, MinVersionTamperProtection)
	fleet.EnforceVersionCheckInto(ctx, client, &f.SupportsMonitoringRuntimeExperimental, &diags, MonitoringRuntimeExperimentalSupported)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsAgentInternal, &diags, MinVersionAgentInternal)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsFeaturesDisablePolicyChangeAcksEnabled, &diags, MinVersionFeaturesDisablePolicyChangeAcksEnabled)
	fleet.EnforceMinVersionInto(ctx, client, &f.SupportsFeaturesIncludeTagsInEventsEnabled, &diags, MinVersionFeaturesIncludeTagsInEventsEnabled)

	return f, diags
}
