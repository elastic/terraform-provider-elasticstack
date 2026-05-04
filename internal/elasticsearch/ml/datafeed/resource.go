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

package datafeed

import (
	"context"
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newDatafeedResource()
	_ resource.ResourceWithConfigure   = newDatafeedResource()
	_ resource.ResourceWithImportState = newDatafeedResource()
	_ resource.ResourceWithModifyPlan  = newDatafeedResource()
)

type datafeedResource struct {
	*entitycore.ResourceBase
}

func newDatafeedResource() *datafeedResource {
	return &datafeedResource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentElasticsearch, "ml_datafeed"),
	}
}

func NewDatafeedResource() resource.Resource {
	return newDatafeedResource()
}

func (r *datafeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.create(ctx, req, resp)
}

func (r *datafeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Datafeed
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datafeedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.update(ctx, req, resp)
}

func (r *datafeedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.delete(ctx, req, resp)
}

// resourceReady checks if the client is ready for API calls
func (r *datafeedResource) resourceReady(diags *fwdiags.Diagnostics) bool {
	if r.Client() == nil {
		diags.AddError("Client not configured", "Provider client is not configured")
		return false
	}
	return true
}

// ModifyPlan normalises the planned query value to the canonical form that
// Elasticsearch stores (e.g. term shorthand → verbose value struct). Without
// this, a plan with {"term":{"field":"v"}} would differ from the state
// {"term":{"field":{"value":"v"}}} after apply, causing an "inconsistent
// result after apply" error.
func (r *datafeedResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only normalise on create/update (plan is null on destroy).
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan Datafeed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Query.IsNull() && !plan.Query.IsUnknown() {
		normalized, diags := normalizeQueryJSON(plan.Query.ValueString())
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.Query = jsontypes.NewNormalizedValue(normalized)
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
	}
}

// normalizeQueryJSON round-trips a query JSON string through types.Query so
// that all shorthand forms (e.g. term value-shorthand) are expanded to the
// canonical verbose form that Elasticsearch stores and returns.
func normalizeQueryJSON(queryJSON string) (string, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	var q estypes.Query
	if err := json.Unmarshal([]byte(queryJSON), &q); err != nil {
		// If unmarshal fails, return the original string unchanged.
		return queryJSON, diags
	}

	normalized, err := json.Marshal(q)
	if err != nil {
		diags.AddError("Failed to normalise query JSON", err.Error())
		return queryJSON, diags
	}

	return string(normalized), diags
}

func (r *datafeedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	compID, sdkDiags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := compID.ResourceID

	// Set the datafeed_id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("datafeed_id"), datafeedID)...)
}
