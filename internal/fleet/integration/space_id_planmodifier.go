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

package integration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func spaceIDRequiresReplace(defaultSpaceID string) planmodifier.String {
	return spaceIDPlanModifier{defaultSpaceID: defaultSpaceID}
}

type spaceIDPlanModifier struct {
	defaultSpaceID string
}

func (m spaceIDPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Explicit change (known -> known).
	if typeutils.IsKnown(req.StateValue) && typeutils.IsKnown(req.PlanValue) {
		resp.RequiresReplace = req.StateValue.ValueString() != req.PlanValue.ValueString()
		return
	}

	stateIsNonDefaultKnown := typeutils.IsKnown(req.StateValue) && req.StateValue.ValueString() != m.defaultSpaceID
	planIsNonDefaultKnown := typeutils.IsKnown(req.PlanValue) && req.PlanValue.ValueString() != m.defaultSpaceID

	// Unknown/null <-> non-default value requires replacement.
	if stateIsNonDefaultKnown && (req.PlanValue.IsUnknown() || req.PlanValue.IsNull()) {
		resp.RequiresReplace = true
		return
	}
	if planIsNonDefaultKnown && (req.StateValue.IsUnknown() || req.StateValue.IsNull()) {
		resp.RequiresReplace = true
		return
	}
}

func (m spaceIDPlanModifier) Description(context.Context) string {
	return fmt.Sprintf("Requires replacement when space_id changes, or when toggling unknown/null with a non-default value (default=%q)", m.defaultSpaceID)
}

func (m spaceIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
