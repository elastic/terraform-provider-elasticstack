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

// SpaceImporter is an embeddable struct that provides a generic ImportState
// implementation for resources that support space-aware composite IDs and
// expose the space as a "space_ids" list attribute (the Fleet convention).
//
// When embedded in a resource struct, Go promotes the ImportState method,
// satisfying resource.ResourceWithImportState without an explicit method.
//
// Usage:
//
//	type myResource struct {
//	    *entitycore.SpaceImporter
//	    // ...
//	}
//
//	func newMyResource() *myResource {
//	    return &myResource{
//	        SpaceImporter: entitycore.NewSpaceImporter(path.Root("resource_id")),
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

	compID, _ := clients.CompositeIDFromStr(req.ID)
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

// KibanaSpaceImporter is an embeddable struct that provides a generic
// ImportState implementation for Kibana resources that support space-aware
// composite IDs and expose the space as a singular "space_id" string
// attribute, rather than Fleet's "space_ids" list.
//
// Unlike SpaceImporter, the import ID is always required to be a composite
// "<space_id>/<resource_id>" string; a diagnostic is added and no attributes
// are set if it is not.
//
// Usage:
//
//	type myResource struct {
//	    *entitycore.KibanaSpaceImporter
//	    // ...
//	}
//
//	func newMyResource() *myResource {
//	    return &myResource{
//	        KibanaSpaceImporter: entitycore.NewKibanaSpaceImporter(
//	            path.Root("id"), path.Root("space_id"), path.Root("rule_id"),
//	        ),
//	    }
//	}
type KibanaSpaceImporter struct {
	idField             path.Path
	spaceIDField        path.Path
	resourceIDFields    []path.Path
	requireSpaceID      bool
	requireSpaceSummary string
	requireSpaceDetail  string
	defaultSpaceID      string
}

// NewKibanaSpaceImporter constructs a KibanaSpaceImporter that will set idField
// to the full import ID, spaceIDField to the space-ID portion of the composite
// ID, and each of resourceIDFields to the resource-ID portion. At least one
// resourceIDField is required.
//
// By default, a composite ID whose space portion is empty (e.g. "/my-id")
// results in spaceIDField being set to an empty string. Use RequireSpaceID or
// DefaultSpaceID to customize that behavior.
func NewKibanaSpaceImporter(idField, spaceIDField path.Path, resourceIDFields ...path.Path) *KibanaSpaceImporter {
	if len(resourceIDFields) == 0 {
		panic("NewKibanaSpaceImporter: at least one resourceIDField is required")
	}
	return &KibanaSpaceImporter{idField: idField, spaceIDField: spaceIDField, resourceIDFields: resourceIDFields}
}

// RequireSpaceID configures the importer to add an error diagnostic (using
// the given summary and detail) instead of setting spaceIDField to an empty
// string when the composite import ID omits a space.
func (s *KibanaSpaceImporter) RequireSpaceID(summary, detail string) *KibanaSpaceImporter {
	s.requireSpaceID = true
	s.requireSpaceSummary = summary
	s.requireSpaceDetail = detail
	return s
}

// DefaultSpaceID configures the importer to fall back to the given space ID
// instead of an empty string when the composite import ID omits a space.
func (s *KibanaSpaceImporter) DefaultSpaceID(spaceID string) *KibanaSpaceImporter {
	s.defaultSpaceID = spaceID
	return s
}

// ImportState handles import for Kibana resources with required space-aware
// composite IDs in the format "<space_id>/<resource_id>".
func (s *KibanaSpaceImporter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite := ParseCompositeImportID(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := composite.ClusterID
	if spaceID == "" {
		switch {
		case s.requireSpaceID:
			resp.Diagnostics.AddError(s.requireSpaceSummary, s.requireSpaceDetail)
			return
		case s.defaultSpaceID != "":
			spaceID = s.defaultSpaceID
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, s.idField, req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, s.spaceIDField, spaceID)...)
	for _, f := range s.resourceIDFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, f, composite.ResourceID)...)
	}
}

// ParseCompositeImportID parses a space-aware composite import ID of the form
// "<space_id>/<resource_id>", appending any diagnostics to resp. Callers must
// check resp.Diagnostics.HasError() before using the returned value.
func ParseCompositeImportID(req resource.ImportStateRequest, resp *resource.ImportStateResponse) *clients.CompositeID {
	composite, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	return composite
}
