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

package dataview

import (
	"context"
	"fmt"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel dataViewModel
	var stateModel dataViewModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &stateModel) // read state
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, planModel.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	body, diags := planModel.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewID, spaceID := planModel.getViewIDAndSpaceID()
	dataView, diags := kibanaoapi.UpdateDataView(ctx, oapiClient, spaceID, viewID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update namespaces via spaces API
	var stateInner, planInner innerModel
	diags = stateModel.DataView.As(ctx, &stateInner, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	diags = planModel.DataView.As(ctx, &planInner, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	if !resp.Diagnostics.HasError() {
		var oldNS, newNS []string
		if !stateInner.Namespaces.IsNull() && !stateInner.Namespaces.IsUnknown() {
			diags = stateInner.Namespaces.ElementsAs(ctx, &oldNS, false)
			resp.Diagnostics.Append(diags...)
		}
		if !planInner.Namespaces.IsNull() && !planInner.Namespaces.IsUnknown() {
			diags = planInner.Namespaces.ElementsAs(ctx, &newNS, false)
			resp.Diagnostics.Append(diags...)
		}

		// A null namespaces list is semantically equal to "[spaceID]" (the resource's
		// own space; see populateFromAPI/handleNamespaces). Treat null the same here so
		// removing an explicit namespaces list does not orphan the data view by removing
		// it from every space (which would 404 every subsequent API call).
		if len(oldNS) == 0 {
			oldNS = []string{spaceID}
		}
		if len(newNS) == 0 {
			newNS = []string{spaceID}
		}

		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(
				kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, viewID, oldNS, newNS)...,
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Apply field_attrs delta via UpdateFieldMetadata (after main update and namespace reconciliation).
	if !planInner.FieldAttributes.IsUnknown() && !stateInner.FieldAttributes.IsUnknown() {
		planFA := map[string]fieldAttrModel{}
		stateFA := map[string]fieldAttrModel{}
		if !planInner.FieldAttributes.IsNull() {
			diags := planInner.FieldAttributes.ElementsAs(ctx, &planFA, false)
			resp.Diagnostics.Append(diags...)
		}
		if !stateInner.FieldAttributes.IsNull() {
			diags := stateInner.FieldAttributes.ElementsAs(ctx, &stateFA, false)
			resp.Diagnostics.Append(diags...)
		}
		if !resp.Diagnostics.HasError() {
			delta := buildFieldAttrsMetadataDelta(planFA, stateFA)
			if len(delta) > 0 {
				resp.Diagnostics.Append(
					kibanaoapi.UpdateFieldMetadata(ctx, oapiClient, spaceID, viewID, delta)...,
				)
				if resp.Diagnostics.HasError() {
					return
				}
				refreshed, refreshDiags := kibanaoapi.GetDataView(ctx, oapiClient, spaceID, viewID)
				resp.Diagnostics.Append(refreshDiags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if refreshed == nil {
					resp.Diagnostics.AddError(
						"Data view disappeared after field_attrs update",
						fmt.Sprintf("UpdateFieldMetadata succeeded for data view %q in space %q, but the subsequent GetDataView returned no data (likely 404). Refusing to write stale state.", viewID, spaceID),
					)
					return
				}
				dataView = refreshed
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	diags = planModel.populateFromAPI(ctx, dataView, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}

// clearedFieldAttrMetadataPayload is sent for fields removed from the plan so Kibana resets
// customLabel/count; an empty JSON object is rejected with HTTP 400 on some stacks.
func clearedFieldAttrMetadataPayload() map[string]any {
	return map[string]any{"customLabel": nil, "count": nil}
}

// buildFieldAttrsMetadataDelta returns a JSON-shaped map for POST .../fields: keys are field names,
// values are objects with camelCase keys matching kbapi.DataViewsFieldattrs (`customLabel`, `count`).
//
// Two subtleties keep server-side semantics safe:
//
//   - An "all-null" plan entry (`field_attrs = { foo = {} }`) is skipped instead of emitting `{}`,
//     because the empty payload is otherwise indistinguishable from the explicit clearing payload.
//     A user who declares an entry with no attributes intends a no-op against Kibana's stored
//     metadata, not a wipe of any server-derived popularity count.
//   - State entries that exist only because Kibana injected a popularity count (custom_label is
//     null and the entry is absent from plan) are NOT cleared. MapSemanticEquals already hides
//     these from the read path; treating them as user-driven removals here would reset Discover
//     popularity counters the user never managed.
func buildFieldAttrsMetadataDelta(planFA, stateFA map[string]fieldAttrModel) map[string]any {
	delta := map[string]any{}

	for name, planEntry := range planFA {
		stateEntry, exists := stateFA[name]
		if exists && planEntry.CustomLabel.Equal(stateEntry.CustomLabel) && planEntry.Count.Equal(stateEntry.Count) {
			continue
		}
		if planEntry.CustomLabel.IsNull() && planEntry.Count.IsNull() {
			continue
		}
		payload := map[string]any{}
		if !planEntry.CustomLabel.IsNull() {
			payload["customLabel"] = planEntry.CustomLabel.ValueString()
		}
		if !planEntry.Count.IsNull() {
			payload["count"] = planEntry.Count.ValueInt64()
		}
		delta[name] = payload
	}
	for name, stateEntry := range stateFA {
		if _, stillInPlan := planFA[name]; stillInPlan {
			continue
		}
		if stateEntry.CustomLabel.IsNull() {
			continue
		}
		delta[name] = clearedFieldAttrMetadataPayload()
	}
	return delta
}
