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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// CompositeIDImporter is an embeddable struct that provides a generic
// ImportState implementation for Elasticsearch resources with a required
// "<cluster_uuid>/<resource_id>" composite state ID: idField is set to the
// raw import ID and each of resourceIDFields is set to the resource-ID
// portion of the composite ID.
//
// When embedded in a resource struct, Go promotes the ImportState method,
// satisfying resource.ResourceWithImportState without an explicit method.
//
// Usage:
//
//	type myResource struct {
//	    *entitycore.ElasticsearchResource[TFModel]
//	    *entitycore.CompositeIDImporter
//	}
//
//	func newMyResource() *myResource {
//	    return &myResource{
//	        ElasticsearchResource: ...,
//	        CompositeIDImporter:   entitycore.NewCompositeIDImporter(path.Root("id"), path.Root("resource_id")),
//	    }
//	}
//
// Resources that need to derive additional state from the resource-ID
// portion (for example splitting it further, or normalizing it via a live
// client call) should use [ParseCompositeID] directly instead.
type CompositeIDImporter struct {
	idField          path.Path
	resourceIDFields []path.Path
}

// NewCompositeIDImporter constructs a CompositeIDImporter that sets idField
// to the raw import ID and each of resourceIDFields to the resource-ID
// portion of the composite import ID. At least one resourceIDField is
// required.
func NewCompositeIDImporter(idField path.Path, resourceIDFields ...path.Path) *CompositeIDImporter {
	if len(resourceIDFields) == 0 {
		panic("NewCompositeIDImporter: at least one resourceIDField is required")
	}
	return &CompositeIDImporter{idField: idField, resourceIDFields: resourceIDFields}
}

// ImportState handles import for resources with a required
// "<cluster_uuid>/<resource_id>" composite ID.
func (c *CompositeIDImporter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	compID, ok := ParseCompositeID(req, resp)
	if !ok {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, c.idField, req.ID)...)
	for _, f := range c.resourceIDFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, f, compID.ResourceID)...)
	}
}

// ParseCompositeID parses req.ID as a "<cluster_uuid>/<resource_id>"
// composite ID, appending an error diagnostic to resp and returning ok=false
// if req.ID is not a valid composite ID.
//
// Use this directly (rather than [CompositeIDImporter]) when a resource's
// ImportState needs to do more than set idField/resourceIDFields verbatim,
// e.g. splitting the resource-ID portion further or normalizing it via a
// live client call, while still sharing the parse+bail-out boilerplate.
func ParseCompositeID(req resource.ImportStateRequest, resp *resource.ImportStateResponse) (*clients.CompositeID, bool) {
	compID, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, false
	}
	return compID, true
}
