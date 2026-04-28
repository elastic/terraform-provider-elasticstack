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

package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// SpaceImporter is an embeddable struct that provides a generic ImportState
// implementation for Fleet resources that support space-aware composite IDs.
//
// When embedded in a resource struct, Go promotes the ImportState method,
// satisfying resource.ResourceWithImportState without an explicit method.
//
// Usage:
//
//	type myResource struct {
//	    *fleet.SpaceImporter
//	    // ...
//	}
//
//	func newMyResource() *myResource {
//	    return &myResource{
//	        SpaceImporter: fleet.NewSpaceImporter(path.Root("resource_id")),
//	    }
//	}
type SpaceImporter struct {
	idFields []path.Path
}

// NewSpaceImporter constructs a SpaceImporter that will set each of the given
// fields to the resource ID on import. At least one field is required.
func NewSpaceImporter(fields ...path.Path) *SpaceImporter {
	if len(fields) == 0 {
		panic("NewSpaceImporter: at least one idField is required")
	}
	return &SpaceImporter{idFields: fields}
}

// ImportState handles import for resources with optional space-aware composite IDs.
//
// The import ID may be either:
//   - A plain resource ID (e.g. "my-policy-id") — sets all idFields to the ID; space_ids is NOT set.
//   - A composite ID (e.g. "my-space/my-policy-id") — sets all idFields to the resource ID portion
//     and sets space_ids to [spaceID].
func (s *SpaceImporter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var spaceID string
	var resourceID string

	compID, _ := clients.CompositeIDFromStrFw(req.ID)
	if compID == nil {
		resourceID = req.ID
	} else {
		spaceID = compID.ClusterID
		resourceID = compID.ResourceID
	}

	for _, f := range s.idFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, f, resourceID)...)
	}

	if spaceID != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_ids"), []string{spaceID})...)
	}
}
