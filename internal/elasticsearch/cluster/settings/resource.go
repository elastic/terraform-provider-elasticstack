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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure clusterSettingsResource satisfies the expected interfaces.
var (
	_ resource.Resource                = newClusterSettingsResource()
	_ resource.ResourceWithConfigure   = newClusterSettingsResource()
	_ resource.ResourceWithImportState = newClusterSettingsResource()
)

// clusterSettingsResource wraps the entitycore envelope, overriding Create,
// Update, and Delete because they require access to both plan and state
// (Update) or need special PUT-to-null semantics (Delete).
type clusterSettingsResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newClusterSettingsResource() *clusterSettingsResource {
	createFn, updateFn := entitycore.PlaceholderElasticsearchWriteCallbacks[tfModel]()
	return &clusterSettingsResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel](
			entitycore.ComponentElasticsearch,
			"cluster_settings",
			getSchema,
			readClusterSettings,
			deleteClusterSettings,
			createFn,
			updateFn,
		),
	}
}

// NewClusterSettingsResource returns the PF resource factory used by the provider registrar.
func NewClusterSettingsResource() resource.Resource {
	return newClusterSettingsResource()
}

// Create overrides the envelope's Create to expand settings and PUT them.
func (r *clusterSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, sdkDiags := client.ID(ctx, "cluster-settings")
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiSettings, diags := getConfiguredSettings(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSettings(ctx, client, apiSettings))...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(id.String())

	result, found, diags := readClusterSettings(ctx, client, "cluster-settings", plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Resource not found after create",
			"elasticstack_elasticsearch_cluster_settings was not found after create.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
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

	// Preserve ID from state.
	plan.ID = state.ID

	result, found, diags := readClusterSettings(ctx, client, "cluster-settings", plan)
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

// Delete overrides the envelope's Delete to null out all tracked settings.
func (r *clusterSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, state.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(deleteClusterSettings(ctx, client, "cluster-settings", state)...)
}

// ImportState implements resource.ResourceWithImportState as a passthrough on id.
func (r *clusterSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
