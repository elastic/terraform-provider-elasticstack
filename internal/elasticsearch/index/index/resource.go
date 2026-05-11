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
	createFn, updateFn := entitycore.PlaceholderElasticsearchWriteCallbacks[tfModel]()
	return &Resource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel](
			entitycore.ComponentElasticsearch,
			"index",
			getSchema,
			readIndex,
			deleteIndex,
			createFn,
			updateFn,
		),
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

	compID, diags := clients.CompositeIDFromStrFw(stateModel.GetID().ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, connDiags := r.Client().GetElasticsearchClient(ctx, stateModel.GetElasticsearchConnection())
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, found, readDiags := readIndex(ctx, client, compID.ResourceID, stateModel)
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
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
