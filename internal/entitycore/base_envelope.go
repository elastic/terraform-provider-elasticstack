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

package entitycore

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// baseResourceEnvelope holds common wiring shared by [ElasticsearchResource]
// and [KibanaResource]: the provider client base, schema factory, connection
// block, and the standard Read / Delete lifecycle implementation.
type baseResourceEnvelope[T any, C MinVersionClient] struct {
	*ResourceBase
	schemaFactory   func(context.Context) rschema.Schema
	connectionKey   string
	connectionBlock rschema.Block
	resolveID       func(T) (string, diag.Diagnostics)
	getClient       func(context.Context, T) (C, diag.Diagnostics)
	read            func(context.Context, C, string, T) (T, bool, diag.Diagnostics)
	delete          func(context.Context, C, string, T) diag.Diagnostics
	postRead        func(context.Context, C, T, T, PrivateStateStorage) (T, diag.Diagnostics)
}

// Schema implements [resource.Resource], injecting the connection block into
// the schema returned by the concrete schema factory.
func (b *baseResourceEnvelope[T, C]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := b.schemaFactory(ctx)
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks[b.connectionKey] = b.connectionBlock
	schema.Blocks = blocks
	resp.Schema = schema
}

// Read implements [resource.Resource] with the standard prelude: check read
// callback is non-nil, deserialize prior state into the generic model T,
// resolve read identity, resolve the scoped client, enforce optional version
// requirements, then delegate to the concrete readFunc.
func (b *baseResourceEnvelope[T, C]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if b.read == nil {
		resp.Diagnostics.Append(requireReadFuncDiag(b.component)...)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, idDiags := b.resolveID(model)
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resolved read identity is empty; cannot read.",
		)
		return
	}

	client, diags := b.getClient(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &model); vDiags.HasError() {
		resp.Diagnostics.Append(vDiags...)
		return
	}

	resultModel, found, callDiags := b.read(ctx, client, resourceID, model)
	resp.Diagnostics.Append(callDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if found {
		if b.postRead != nil {
			var prDiags diag.Diagnostics
			resultModel, prDiags = b.postRead(ctx, client, model, resultModel, resp.Private)
			resp.Diagnostics.Append(prDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &resultModel)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

// Delete implements [resource.Resource] with the standard prelude, then
// delegates to the concrete deleteFunc.
func (b *baseResourceEnvelope[T, C]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if b.delete == nil {
		resp.Diagnostics.Append(requireDeleteFuncDiag(b.component)...)
		return
	}

	var model T
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, idDiags := b.resolveID(model)
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Invalid resource identifier",
			"The resolved delete identity is empty; cannot delete.",
		)
		return
	}

	client, diags := b.getClient(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(b.delete(ctx, client, resourceID, model)...)
}
