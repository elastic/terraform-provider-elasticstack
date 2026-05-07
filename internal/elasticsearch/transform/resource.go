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

package transform

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                 = newTransformResource()
	_ resource.ResourceWithConfigure    = newTransformResource()
	_ resource.ResourceWithImportState  = newTransformResource()
	_ resource.ResourceWithUpgradeState = newTransformResource()
)

// transformResource wraps the entitycore envelope and overrides Create and
// Update because the transform lifecycle requires real create/update callbacks
// and update needs to compare old vs new enabled state to issue Start/Stop.
type transformResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newTransformResource() *transformResource {
	createFn, updateFn := entitycore.PlaceholderElasticsearchWriteCallbacks[tfModel]()
	return &transformResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel](
			entitycore.ComponentElasticsearch,
			"transform",
			getSchema,
			readTransform,
			deleteTransform,
			createFn,
			updateFn,
		),
	}
}

// NewTransformResource returns the PF resource factory for registration.
func NewTransformResource() resource.Resource {
	return newTransformResource()
}

// Create overrides the envelope Create to use our real createTransform callback.
func (r *transformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := model.GetResourceID().ValueString()
	if resourceID == "" {
		resp.Diagnostics.AddError("Invalid resource identifier", "Transform name is empty")
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdModel, createDiags := createTransform(ctx, client, resourceID, model)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back to populate all computed fields.
	resultModel, found, readDiags := readTransform(ctx, client, resourceID, createdModel)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Resource not found after create",
			"Transform was not found immediately after creation",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
}

// Update overrides the envelope Update to detect enabled state changes and
// issue Start/Stop Transform calls as required.
func (r *transformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.GetResourceID().ValueString()
	if resourceID == "" {
		resp.Diagnostics.AddError("Invalid resource identifier", "Transform name is empty")
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, plan.GetElasticsearchConnection())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve server version for version-gated fields.
	serverVersion, sdkDiags := client.ServerVersion(ctx)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to API model.
	apiTransform, convDiags := toAPIModel(ctx, plan, serverVersion)
	resp.Diagnostics.Append(convDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pivot and Latest are immutable; omit them from the update request.
	apiTransform.Pivot = nil
	apiTransform.Latest = nil

	timeout, parseDiags := plan.Timeout.Parse()
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deferValidation := plan.DeferValidation.ValueBool()

	// Detect enabled change: only call start/stop when the enabled flag changes.
	wasEnabled := state.Enabled.ValueBool()
	willBeEnabled := plan.Enabled.ValueBool()
	enabledChanged := wasEnabled != willBeEnabled

	sdkDiags = elasticsearch.UpdateTransform(ctx, client, apiTransform, deferValidation, timeout, willBeEnabled, enabledChanged)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back to refresh state.
	resultModel, found, readDiags := readTransform(ctx, client, resourceID, plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Resource not found after update",
			"Transform was not found immediately after update",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
}

// UpgradeState provides state upgraders for prior schema versions. The v0→v1
// upgrade unwraps singleton-list nested blocks (source, destination,
// retention_policy, sync, and their inner time blocks) into single objects
// after the schema migration from ListNestedBlock to SingleNestedBlock.
func (r *transformResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: migrateStateV0ToV1},
	}
}

// ImportState implements passthrough import on the composite id attribute.
// It also extracts the transform name from the composite ID so Read can use it.
func (r *transformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	compID, sdkDiags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), compID.ResourceID)...)
}
