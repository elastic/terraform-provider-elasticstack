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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sortPrivateState holds the ordered sort configuration stored in private state during Read.
type sortPrivateState struct {
	Fields  []string `json:"fields"`
	Orders  []string `json:"orders"`
	Missing []string `json:"missing,omitempty"`
	Mode    []string `json:"mode,omitempty"`
}

// sortMigrationPlanModifier suppresses replace when migrating from legacy
// sort_field/sort_order attributes to the new sort ListNestedAttribute with
// semantically equivalent settings.
type sortMigrationPlanModifier struct{}

func (m sortMigrationPlanModifier) Description(_ context.Context) string {
	return "Suppresses replace when migrating from legacy sort attributes with equivalent settings."
}

func (m sortMigrationPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sortMigrationPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If the attribute is being removed (plan is null) but state has a sort
	// block, require replace: index sorting is immutable and cannot be removed
	// in-place without recreating the index.
	if req.PlanValue.IsNull() {
		if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
			resp.RequiresReplace = true
		}
		return
	}

	// If sort is non-null in state, this is a normal update of an existing
	// sort attribute. Sort settings are immutable, so any change requires replace.
	if !req.StateValue.IsNull() {
		resp.RequiresReplace = true
		return
	}

	// sort is null in state and non-null in plan. This could be either:
	// 1. A create of a new index with sort settings.
	// 2. A migration from legacy sort_field/sort_order to the new sort attribute.
	//
	// Distinguish by checking if private state has sort_config. If present,
	// this is a migration and we can potentially suppress the replace.

	// Read private state.
	privateStateBytes, err := req.Private.GetKey(ctx, "sort_config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read private state",
			fmt.Sprintf("Could not read sort_config from private state: %s", err),
		)
		resp.RequiresReplace = true
		return
	}

	// If private state is absent (first apply after provider upgrade without prior Read),
	// default to requiring replace.
	if privateStateBytes == nil {
		resp.RequiresReplace = true
		return
	}

	var ps sortPrivateState
	if err := json.Unmarshal(privateStateBytes, &ps); err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse private state",
			fmt.Sprintf("Could not parse sort_config from private state: %s", err),
		)
		resp.RequiresReplace = true
		return
	}

	// Parse the plan's sort elements.
	planElements := make([]attr.Value, len(req.PlanValue.Elements()))
	copy(planElements, req.PlanValue.Elements())

	// Compare fields in order.
	if len(planElements) != len(ps.Fields) {
		resp.RequiresReplace = true
		return
	}

	defaultOrder := types.StringValue("asc")

	for i, elem := range planElements {
		elemObj, ok := elem.(types.Object)
		if !ok {
			resp.RequiresReplace = true
			return
		}

		attrs := elemObj.Attributes()

		// Field must match.
		field, ok := attrs["field"].(types.String)
		if !ok || !field.Equal(types.StringValue(ps.Fields[i])) {
			resp.RequiresReplace = true
			return
		}

		// Order: treat null/unknown as "asc" default.
		order, ok := attrs["order"].(types.String)
		if !ok {
			resp.RequiresReplace = true
			return
		}
		if order.IsNull() || order.IsUnknown() {
			order = defaultOrder
		}
		// Guard against ps.Orders being shorter than ps.Fields (malformed private state).
		// Treat absent order entries as the ES default "asc".
		expectedOrderStr := "asc"
		if i < len(ps.Orders) {
			expectedOrderStr = ps.Orders[i]
		}
		expectedOrder := types.StringValue(expectedOrderStr)
		if !order.Equal(expectedOrder) {
			resp.RequiresReplace = true
			return
		}

		// Missing: treat null/unknown or explicit "_last" as equivalent to absent.
		missing, ok := attrs["missing"].(types.String)
		if !ok {
			resp.RequiresReplace = true
			return
		}
		expectedMissing := ""
		if i < len(ps.Missing) {
			expectedMissing = ps.Missing[i]
		}
		if !isSemanticallyEquivalentMissing(missing, expectedMissing) {
			resp.RequiresReplace = true
			return
		}

		// Mode: treat null/unknown or explicit default as equivalent to absent.
		mode, ok := attrs["mode"].(types.String)
		if !ok {
			resp.RequiresReplace = true
			return
		}
		expectedMode := ""
		if i < len(ps.Mode) {
			expectedMode = ps.Mode[i]
		}
		if !isSemanticallyEquivalentMode(mode, expectedMode, order) {
			resp.RequiresReplace = true
			return
		}
	}

	// All checks passed — suppress replace.
	// The config change (legacy attrs → new sort block) will produce an update action,
	// but no destroy+create cycle. Index sort settings are stored in ES private state
	// and read back via the new sort block on the next Read.
	resp.RequiresReplace = false
}

// isSemanticallyEquivalentMissing returns true if the planned missing value
// (which may be null/unknown) is equivalent to the existing index setting.
// Elasticsearch default missing is "_last" regardless of order.
func isSemanticallyEquivalentMissing(planned types.String, existing string) bool {
	// Determine the effective planned value, treating null/unknown as "_last".
	plannedVal := "_last"
	if !planned.IsNull() && !planned.IsUnknown() {
		plannedVal = planned.ValueString()
	}

	// Determine the effective existing value, treating empty as the ES default "_last".
	existingVal := existing
	if existingVal == "" {
		existingVal = "_last"
	}

	return plannedVal == existingVal
}

// isSemanticallyEquivalentMode returns true if the planned mode value
// (which may be null/unknown) is equivalent to the existing index setting.
// ES default mode: "min" for asc, "max" for desc.
func isSemanticallyEquivalentMode(planned types.String, existing string, order types.String) bool {
	// Determine the effective planned value, treating null/unknown as default based on order.
	plannedVal := ""
	if !planned.IsNull() && !planned.IsUnknown() {
		plannedVal = planned.ValueString()
	} else {
		// Default mode based on order
		if order.ValueString() == "desc" {
			plannedVal = "max"
		} else {
			plannedVal = "min"
		}
	}

	// Determine the effective existing value, treating empty as the ES default based on order.
	existingVal := existing
	if existingVal == "" {
		if order.ValueString() == "desc" {
			existingVal = "max"
		} else {
			existingVal = "min"
		}
	}

	return plannedVal == existingVal
}

// legacySortFieldPlanModifier suppresses replace for sort_field when sort is
// present in the plan (the new sort attribute handles the replace decision).
type legacySortFieldPlanModifier struct{}

func (m legacySortFieldPlanModifier) Description(_ context.Context) string {
	return "Suppresses replace when the new sort attribute is used in the plan."
}

func (m legacySortFieldPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m legacySortFieldPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Read sort from plan
	var planSort types.List
	diags := req.Plan.GetAttribute(ctx, path.Root("sort"), &planSort)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = true
		return
	}

	// If sort is non-null in the plan, suppress replace for sort_field.
	// The sortMigrationPlanModifier on the new attribute handles the replace decision.
	if !planSort.IsNull() && !planSort.IsUnknown() {
		resp.RequiresReplace = false
		return
	}

	// Otherwise, require replace (immutable static setting).
	resp.RequiresReplace = true
}

// legacySortOrderPlanModifier suppresses replace for sort_order when sort is
// present in the plan (the new sort attribute handles the replace decision).
type legacySortOrderPlanModifier struct{}

func (m legacySortOrderPlanModifier) Description(_ context.Context) string {
	return "Suppresses replace when the new sort attribute is used in the plan."
}

func (m legacySortOrderPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m legacySortOrderPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Read sort from plan
	var planSort types.List
	diags := req.Plan.GetAttribute(ctx, path.Root("sort"), &planSort)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		resp.RequiresReplace = true
		return
	}

	// If sort is non-null in the plan, suppress replace for sort_order.
	if !planSort.IsNull() && !planSort.IsUnknown() {
		resp.RequiresReplace = false
		return
	}

	// Otherwise, require replace (immutable static setting).
	resp.RequiresReplace = true
}
