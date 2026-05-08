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
	"maps"

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

// clusterSettingsResource wraps the entitycore envelope. Create and Delete are
// supplied as callbacks. Update is overridden because it requires both plan
// and prior state to compute the null entries for removed settings, which the
// envelope's update callback contract does not expose.
type clusterSettingsResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newClusterSettingsResource() *clusterSettingsResource {
	_, updatePlaceholder := entitycore.PlaceholderElasticsearchWriteCallbacks[tfModel]()

	return &clusterSettingsResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel](
			entitycore.ComponentElasticsearch,
			"cluster_settings",
			getSchema,
			readClusterSettings,
			deleteClusterSettings,
			createClusterSettings,
			updatePlaceholder,
		),
	}
}

// NewClusterSettingsResource returns the PF resource factory used by the provider registrar.
func NewClusterSettingsResource() resource.Resource {
	return newClusterSettingsResource()
}

// createClusterSettings implements the envelope's create callback: PUT the
// configured settings and stamp the composite ID onto the returned model. The
// envelope handles plan decoding, client resolution, and the read-after-write.
func createClusterSettings(ctx context.Context, client *clients.ElasticsearchScopedClient, _ string, plan tfModel) (tfModel, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	apiSettings, ds := getConfiguredSettings(ctx, plan)
	diags.Append(ds...)
	if diags.HasError() {
		return plan, diags
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(id.String())
	return plan, diags
}

// Update overrides the envelope's Update to compare old and new settings,
// null out removed keys, then PUT the merged settings map.
func (r *clusterSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldSettings, diags := getConfiguredSettings(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newSettings, diags := getConfiguredSettings(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Start from the new settings and add null entries for removed keys.
	apiSettings := make(map[string]any)
	maps.Copy(apiSettings, newSettings)
	for _, category := range []string{"persistent", "transient"} {
		oldCat, _ := oldSettings[category].(map[string]any)
		newCat, _ := newSettings[category].(map[string]any)
		if oldCat == nil {
			oldCat = make(map[string]any)
		}
		if newCat == nil {
			newCat = make(map[string]any)
		}
		updateRemovedSettings(category, oldCat, newCat, apiSettings)
	}

	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	result, found, diags := readClusterSettings(ctx, client, resourceID, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Resource not found after update",
			"elasticstack_elasticsearch_cluster_settings was not found after update.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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
