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

package filter

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newFilterResource()
	_ resource.ResourceWithConfigure   = newFilterResource()
	_ resource.ResourceWithImportState = newFilterResource()
)

// filterResource uses [entitycore.ElasticsearchResource], which wires the
// provider factory, injects the elasticsearch_connection block into Schema,
// resolves a scoped Elasticsearch client per resource, and passes the ML API
// identifier (filter id segment) into read/delete after parsing the composite
// state id. Create and update callbacks receive that identifier from
// [TFModel.GetResourceID] (filter_id).
type filterResource struct {
	*entitycore.ElasticsearchResource[TFModel]
}

func newFilterResource() *filterResource {
	return &filterResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[TFModel]("ml_filter", entitycore.ElasticsearchResourceOptions[TFModel]{
			Schema: getSchema,
			Read:   readFilter,
			Delete: deleteFilter,
			Create: createFilter,
			Update: updateFilter,
		}),
	}
}

func NewFilterResource() resource.Resource {
	return newFilterResource()
}

// ImportState accepts "<cluster_uuid>/<filter_id>" import ids and sets both id
// and filter_id so Destroy and Read use the same composite id shape as for
// normally managed resources.
func (r *filterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, diags := clients.CompositeIDFromStrFw(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("filter_id"), compID.ResourceID)...)
}
