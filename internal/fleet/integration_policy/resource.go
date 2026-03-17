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

package integrationpolicy

import (
	"context"
	"fmt"
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                 = &integrationPolicyResource{}
	_ resource.ResourceWithConfigure    = &integrationPolicyResource{}
	_ resource.ResourceWithImportState  = &integrationPolicyResource{}
	_ resource.ResourceWithUpgradeState = &integrationPolicyResource{}
)

var (
	MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))
	MinVersionOutputID  = version.Must(version.NewVersion("8.16.0"))
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &integrationPolicyResource{}
}

type integrationPolicyResource struct {
	client *clients.APIClient
}

func (r *integrationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *integrationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_integration_policy")
}

func (r *integrationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)
}

func (r *integrationPolicyResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {PriorSchema: getSchemaV0(), StateUpgrader: upgradeV0ToV2},
		1: {PriorSchema: getSchemaV1(), StateUpgrader: upgradeV1ToV2},
	}
}

func (r *integrationPolicyResource) buildFeatures(ctx context.Context) (features, diag.Diagnostics) {
	supportsPolicyIDs, diags := r.client.EnforceMinVersion(ctx, MinVersionPolicyIDs)
	if diags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(diags)
	}

	supportsOutputID, outputIDDiags := r.client.EnforceMinVersion(ctx, MinVersionOutputID)
	if outputIDDiags.HasError() {
		return features{}, diagutil.FrameworkDiagsFromSDK(outputIDDiags)
	}

	return features{
		SupportsPolicyIDs: supportsPolicyIDs,
		SupportsOutputID:  supportsOutputID,
	}, nil
}

var knownPackages sync.Map

func getPackageCacheKey(name string, version string) string {
	return fmt.Sprintf("%s-%s", name, version)
}

func getPackageInfo(ctx context.Context, client *fleet.Client, name string, version string) (*kbapi.PackageInfo, diag.Diagnostics) {
	var diags diag.Diagnostics

	if pkg, ok := getCachedPackageInfo(name, version); ok {
		return &pkg, diags
	}

	pkg, diags := fleet.GetPackage(ctx, client, name, version)
	if diags.HasError() {
		return nil, diags
	}
	knownPackages.Store(getPackageCacheKey(name, version), *pkg)
	return pkg, diags
}

func getCachedPackageInfo(name string, version string) (kbapi.PackageInfo, bool) {
	value, ok := knownPackages.Load(getPackageCacheKey(name, version))
	if !ok {
		return kbapi.PackageInfo{}, false
	}
	pkg, ok := value.(kbapi.PackageInfo)
	if !ok {
		return kbapi.PackageInfo{}, false
	}
	return pkg, true
}
