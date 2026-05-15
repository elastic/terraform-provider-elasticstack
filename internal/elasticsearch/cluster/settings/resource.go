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

package settings

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure clusterSettingsResource satisfies the expected interfaces.
var (
	_ resource.Resource                   = newClusterSettingsResource()
	_ resource.ResourceWithConfigure      = newClusterSettingsResource()
	_ resource.ResourceWithImportState    = newClusterSettingsResource()
	_ resource.ResourceWithUpgradeState   = newClusterSettingsResource()
	_ resource.ResourceWithValidateConfig = newClusterSettingsResource()
)

// clusterSettingsResource wraps the entitycore envelope for the singleton
// cluster_settings resource.
type clusterSettingsResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newClusterSettingsResource() *clusterSettingsResource {
	return &clusterSettingsResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel]("cluster_settings", entitycore.ElasticsearchResourceOptions[tfModel]{
			Schema: getSchema,
			Read:   readClusterSettings,
			Delete: deleteClusterSettings,
			Create: createClusterSettings,
			Update: updateClusterSettings,
		}),
	}
}

// NewClusterSettingsResource returns the PF resource factory used by the provider registrar.
func NewClusterSettingsResource() resource.Resource {
	return newClusterSettingsResource()
}

// createClusterSettings implements the envelope's Create callback: PUT the
// configured settings and stamp the composite ID onto the returned model. The
// envelope handles plan decoding, client resolution, and the read-after-write.
func createClusterSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	plan := req.Plan

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	apiSettings, ds := getConfiguredSettings(ctx, plan)
	diags.Append(ds...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	plan.ID = types.StringValue(id.String())
	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}

// ImportState implements resource.ResourceWithImportState as a passthrough on id.
func (r *clusterSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig ensures that at least one of persistent or transient is
// configured so the resource does not represent an empty no-op configuration.
func (r *clusterSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config tfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateConfigModel(config)...)
}

// validateConfigModel implements the rule that at least one of persistent or
// transient must be a non-empty block. Extracted so it can be unit-tested
// without constructing a tfsdk.Config.
func validateConfigModel(config tfModel) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if categoryBlockEmpty(config.Persistent) && categoryBlockEmpty(config.Transient) {
		diags.AddError(
			"No cluster settings configured",
			`At least one of "persistent" or "transient" must contain at least one "setting" block.`,
		)
	}
	return diags
}

// categoryBlockEmpty reports whether the given persistent/transient block is
// effectively empty: null, unknown, or contains a setting set with no
// elements.
func categoryBlockEmpty(block types.Object) bool {
	if block.IsNull() || block.IsUnknown() {
		return true
	}
	settingAttr, ok := block.Attributes()["setting"]
	if !ok {
		return true
	}
	settingSet, ok := settingAttr.(types.Set)
	if !ok {
		return true
	}
	return settingSet.IsNull() || settingSet.IsUnknown() || len(settingSet.Elements()) == 0
}

// UpgradeState migrates state written by the SDKv2-based implementation
// (schema version 0) where Optional list/string attributes were serialised as
// empty lists / empty strings rather than nulls. This caused set-element
// identity churn after upgrading to the Plugin Framework implementation
// (schema version 1) where the same logical absence is represented as null.
func (r *clusterSettingsResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: migrateClusterSettingsStateV0ToV1},
	}
}
