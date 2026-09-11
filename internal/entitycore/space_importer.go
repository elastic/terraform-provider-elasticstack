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
// DefaultSpaceID to customize that behavior. RequireSpaceID and DefaultSpaceID
// are mutually exclusive.
func NewKibanaSpaceImporter(idField, spaceIDField path.Path, resourceIDFields ...path.Path) *KibanaSpaceImporter {
	if len(resourceIDFields) == 0 {
		panic("NewKibanaSpaceImporter: at least one resourceIDField is required")
	}
	return &KibanaSpaceImporter{idField: idField, spaceIDField: spaceIDField, resourceIDFields: resourceIDFields}
}

// RequireSpaceID configures the importer to add an error diagnostic (using
// the given summary and detail) instead of setting spaceIDField to an empty
// string when the composite import ID omits a space.
//
// Panics if DefaultSpaceID has already been configured.
func (s *KibanaSpaceImporter) RequireSpaceID(summary, detail string) *KibanaSpaceImporter {
	if s.defaultSpaceID != "" {
		panic("KibanaSpaceImporter: RequireSpaceID and DefaultSpaceID are mutually exclusive")
	}
	s.requireSpaceID = true
	s.requireSpaceSummary = summary
	s.requireSpaceDetail = detail
	return s
}

// DefaultSpaceID configures the importer to fall back to the given space ID
// instead of an empty string when the composite import ID omits a space. When
// the fallback is applied, idField is set to the canonical
// "<defaultSpaceID>/<resource_id>" form rather than the raw import ID.
//
// Panics if spaceID is empty or RequireSpaceID has already been configured.
func (s *KibanaSpaceImporter) DefaultSpaceID(spaceID string) *KibanaSpaceImporter {
	if spaceID == "" {
		panic("KibanaSpaceImporter: DefaultSpaceID must be non-empty")
	}
	if s.requireSpaceID {
		panic("KibanaSpaceImporter: RequireSpaceID and DefaultSpaceID are mutually exclusive")
	}
	s.defaultSpaceID = spaceID
	return s
}

// ImportState handles import for Kibana resources with required space-aware
// composite IDs in the format "<space_id>/<resource_id>".
func (s *KibanaSpaceImporter) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	composite, diags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	s.SeedState(ctx, resp, req.ID, composite)
}

// SeedState applies empty-space policy and sets idField, spaceIDField, and
// resourceIDFields from an already-parsed composite ID. Callers that need
// validation before any attributes are written should parse and validate
// first, then call SeedState.
func (s *KibanaSpaceImporter) SeedState(ctx context.Context, resp *resource.ImportStateResponse, importID string, composite *clients.CompositeID) {
	spaceID := composite.ClusterID
	idValue := importID
	if spaceID == "" {
		switch {
		case s.requireSpaceID:
			resp.Diagnostics.AddError(s.requireSpaceSummary, s.requireSpaceDetail)
			return
		case s.defaultSpaceID != "":
			spaceID = s.defaultSpaceID
			idValue = (&clients.CompositeID{ClusterID: spaceID, ResourceID: composite.ResourceID}).String()
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, s.idField, idValue)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, s.spaceIDField, spaceID)...)
	for _, f := range s.resourceIDFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, f, composite.ResourceID)...)
	}
}
