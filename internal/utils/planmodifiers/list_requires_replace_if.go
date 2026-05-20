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

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ListShouldRequiresReplaceFunc decides whether a list attribute change should force
// resource replacement during plan modification.
type ListShouldRequiresReplaceFunc func(ctx context.Context, plan, state types.List) bool

// ListRequiresReplaceIf returns a planmodifier.List that marks the attribute for
// replacement when shouldReplace returns true. Unknown plan or state values are
// ignored so computed connection blocks do not spuriously force replace.
func ListRequiresReplaceIf(description string, shouldReplace ListShouldRequiresReplaceFunc) planmodifier.List {
	return listRequiresReplaceIf{description: description, shouldReplace: shouldReplace}
}

type listRequiresReplaceIf struct {
	description   string
	shouldReplace ListShouldRequiresReplaceFunc
}

func (m listRequiresReplaceIf) Description(context.Context) string { return m.description }

func (m listRequiresReplaceIf) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m listRequiresReplaceIf) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}
	if req.StateValue.IsUnknown() || req.PlanValue.IsUnknown() {
		return
	}
	if m.shouldReplace(ctx, req.PlanValue, req.StateValue) {
		resp.RequiresReplace = true
	}
}

// ConnectionBlockRequiresReplaceOnChange forces replacement when a per-resource
// elasticsearch_connection or kibana_connection block changes between plan and state.
func ConnectionBlockRequiresReplaceOnChange() planmodifier.List {
	return ListRequiresReplaceIf(
		"Changes to the connection block require replacing the resource because the provider cannot migrate existing remote state to a different cluster endpoint.",
		func(_ context.Context, plan, state types.List) bool {
			return !plan.Equal(state)
		},
	)
}

// ModifyPlanRequiresReplaceOnConnectionChange appends RequiresReplace for connectionPath when
// plan and state connection lists differ. Used by entity envelopes because provider-schema
// connection blocks cannot host resource-level plan modifiers.
func ModifyPlanRequiresReplaceOnConnectionChange(
	planConn, stateConn types.List,
	connectionPath path.Path,
	resp *resource.ModifyPlanResponse,
) {
	if planConn.IsUnknown() || stateConn.IsUnknown() {
		return
	}
	if !planConn.Equal(stateConn) {
		resp.RequiresReplace = append(resp.RequiresReplace, connectionPath)
	}
}
