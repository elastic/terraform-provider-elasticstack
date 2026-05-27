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

package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
	_ resource.ResourceWithModifyPlan  = newResource()
)

type Resource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

// Equivalent to privatestate.ProviderData
type privateData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

func newResource() *Resource {
	placeholder := entitycore.PlaceholderElasticsearchWriteCallback[tfModel]()
	return &Resource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel]("index", entitycore.ElasticsearchResourceOptions[tfModel]{
			Schema: getSchema,
			Read: func(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, stateModel tfModel) (tfModel, bool, diag.Diagnostics) {
				return readIndex(ctx, client, resourceID, stateModel, false)
			},
			Delete: deleteIndex,
			Create: placeholder,
			Update: placeholder,
		}),
	}
}

// NewResource returns an index resource with shared bootstrap wiring.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStr(stateModel.GetID().ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, stateModel.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydrateAllBytes, privDiags := req.Private.GetKey(ctx, importHydrationPrivateStateKey)
	resp.Diagnostics.Append(privDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	hydrateAll := len(hydrateAllBytes) > 0

	finalModel, found, readDiags := readIndex(ctx, client, compID.ResourceID, stateModel, hydrateAll)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(saveSortConfig(ctx, finalModel, resp.Private)...)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, importHydrationPrivateStateKey, []byte("true"))...)
}

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	hydrateAllBytes, diags := req.Private.GetKey(ctx, importHydrationPrivateStateKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(hydrateAllBytes) == 0 {
		return
	}

	clearImportHydrationFlag := func() {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, importHydrationPrivateStateKey, nil)...)
	}

	if req.Plan.Raw.IsNull() {
		clearImportHydrationFlag()
		return
	}

	var planModel, configModel tfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pruneImportHydratedPlanFields(ctx, &planModel, &configModel)

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clearImportHydrationFlag()
}
