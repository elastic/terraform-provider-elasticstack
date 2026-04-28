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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newAgentPolicyResource()
	_ resource.ResourceWithConfigure   = newAgentPolicyResource()
	_ resource.ResourceWithImportState = newAgentPolicyResource()
)

var (
	MinVersionGlobalDataTags      = version.Must(version.NewVersion("8.15.0"))
	MinSupportsAgentlessVersion   = version.Must(version.NewVersion("8.15.0"))
	MinVersionInactivityTimeout   = version.Must(version.NewVersion("8.7.0"))
	MinVersionUnenrollmentTimeout = version.Must(version.NewVersion("8.15.0"))
	MinVersionSpaceIDs            = version.Must(version.NewVersion("9.1.0"))
	MinVersionRequiredVersions    = version.Must(version.NewVersion("9.1.0"))
	MinVersionAgentFeatures       = version.Must(version.NewVersion("8.7.0"))
	MinVersionAdvancedMonitoring  = version.Must(version.NewVersion("8.16.0"))
	MinVersionAdvancedSettings    = version.Must(version.NewVersion("8.17.0"))
)

type agentPolicyResource struct {
	*entitycore.ResourceBase
	*fleet.SpaceImporter
}

func newAgentPolicyResource() *agentPolicyResource {
	return &agentPolicyResource{
		ResourceBase:  entitycore.NewResourceBase(entitycore.ComponentFleet, "agent_policy"),
		SpaceImporter: fleet.NewSpaceImporter(path.Root("policy_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newAgentPolicyResource()
}

func (r *agentPolicyResource) buildFeatures(ctx context.Context, apiClient *clients.KibanaScopedClient) (features, diag.Diagnostics) {
	supportsGDT, diags := apiClient.EnforceMinVersion(ctx, MinVersionGlobalDataTags)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsSupportsAgentless, diags := apiClient.EnforceMinVersion(ctx, MinSupportsAgentlessVersion)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsInactivityTimeout, diags := apiClient.EnforceMinVersion(ctx, MinVersionInactivityTimeout)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsUnenrollmentTimeout, diags := apiClient.EnforceMinVersion(ctx, MinVersionUnenrollmentTimeout)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsSpaceIDs, diags := apiClient.EnforceMinVersion(ctx, MinVersionSpaceIDs)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsRequiredVersions, diags := apiClient.EnforceMinVersion(ctx, MinVersionRequiredVersions)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsAgentFeatures, diags := apiClient.EnforceMinVersion(ctx, MinVersionAgentFeatures)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsAdvancedMonitoring, diags := apiClient.EnforceMinVersion(ctx, MinVersionAdvancedMonitoring)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsAdvancedSettings, diags := apiClient.EnforceMinVersion(ctx, MinVersionAdvancedSettings)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	return features{
		SupportsGlobalDataTags:      supportsGDT,
		SupportsSupportsAgentless:   supportsSupportsAgentless,
		SupportsInactivityTimeout:   supportsInactivityTimeout,
		SupportsUnenrollmentTimeout: supportsUnenrollmentTimeout,
		SupportsSpaceIDs:            supportsSpaceIDs,
		SupportsRequiredVersions:    supportsRequiredVersions,
		SupportsAgentFeatures:       supportsAgentFeatures,
		SupportsAdvancedMonitoring:  supportsAdvancedMonitoring,
		SupportsAdvancedSettings:    supportsAdvancedSettings,
	}, nil
}
