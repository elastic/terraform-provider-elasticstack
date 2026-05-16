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

package sourcemap

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = newResourceSourceMap()
	_ resource.ResourceWithConfigure   = newResourceSourceMap()
	_ resource.ResourceWithImportState = newResourceSourceMap()
)

type resourceSourceMap struct {
	*entitycore.ResourceBase
}

func newResourceSourceMap() *resourceSourceMap {
	return &resourceSourceMap{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentAPM, "source_map"),
	}
}

// NewSourceMapResource returns the resource.Resource for use in provider registration.
func NewSourceMapResource() resource.Resource {
	return newResourceSourceMap()
}

// Update is not supported. All write attributes use RequireReplace, so any
// change triggers destroy + create. This method is required by the Plugin
// Framework interface but should never be invoked in practice.
func (r *resourceSourceMap) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected Update call on elasticstack_apm_source_map",
		"All attributes require replacement; in-place updates are not supported. This is a provider bug.",
	)
}

// ImportState handles space-aware composite import IDs.
//
// Accepted formats:
//   - "<space_id>/<artifact_id>" — sets space_id and id from the two parts.
//   - "<artifact_id>"           — sets id only; space_id is left unset (default space).
func (r *resourceSourceMap) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, diags := clients.CompositeIDFromStrFw(req.ID)
	if diags.HasError() {
		// Plain ID (no slash) — set id only, space_id remains unset.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), compID.ClusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), compID.ResourceID)...)
}
