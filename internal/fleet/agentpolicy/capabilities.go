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
}

func resolveAgentPolicyFeatures(ctx context.Context, client *clients.KibanaScopedClient) (agentPolicyFeatures, diag.Diagnostics) {
	var diags diag.Diagnostics
	var f agentPolicyFeatures

	var bitDiags diag.Diagnostics
	f.SupportsGlobalDataTags, bitDiags = client.EnforceMinVersion(ctx, MinVersionGlobalDataTags)
	diags.Append(bitDiags...)
	f.SupportsSupportsAgentless, bitDiags = client.EnforceMinVersion(ctx, MinSupportsAgentlessVersion)
	diags.Append(bitDiags...)
	f.SupportsInactivityTimeout, bitDiags = client.EnforceMinVersion(ctx, MinVersionInactivityTimeout)
	diags.Append(bitDiags...)
	f.SupportsUnenrollmentTimeout, bitDiags = client.EnforceMinVersion(ctx, MinVersionUnenrollmentTimeout)
	diags.Append(bitDiags...)
	f.SupportsSpaceIDs, bitDiags = client.EnforceMinVersion(ctx, MinVersionSpaceIDs)
	diags.Append(bitDiags...)
	f.SupportsRequiredVersions, bitDiags = client.EnforceMinVersion(ctx, MinVersionRequiredVersions)
	diags.Append(bitDiags...)
	f.SupportsAgentFeatures, bitDiags = client.EnforceMinVersion(ctx, MinVersionAgentFeatures)
	diags.Append(bitDiags...)
	f.SupportsAdvancedMonitoring, bitDiags = client.EnforceMinVersion(ctx, MinVersionAdvancedMonitoring)
	diags.Append(bitDiags...)
	f.SupportsAdvancedSettings, bitDiags = client.EnforceMinVersion(ctx, MinVersionAdvancedSettings)
	diags.Append(bitDiags...)
	f.SupportsTamperProtection, bitDiags = client.EnforceMinVersion(ctx, MinVersionTamperProtection)
	diags.Append(bitDiags...)
	f.SupportsMonitoringRuntimeExperimental, bitDiags = client.EnforceVersionCheck(ctx, MonitoringRuntimeExperimentalSupported)
	diags.Append(bitDiags...)

	return f, diags
}
