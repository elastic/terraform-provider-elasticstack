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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// commentsAppendOnlyRequireReplacePlanModifier triggers a full resource
// replacement whenever the planned `comments` value differs from the prior
// state in any way other than appending one or more brand-new entries at
// the end. Kibana's PUT /api/exception_lists/items treats the field as
// append-only and immutable: reordering/prepending returns HTTP 400;
// omitting existing entries (shrink) and editing existing entries' text
// both return 200 but are silently ignored. Without this modifier the
// first case surfaces as an opaque API error and the latter two as
// "produced inconsistent result after apply" diagnostics.
//
// Trade-off: Kibana stamps every comment's `created_by` from the
// authenticated API key owner — there is no client-supplied way to
// preserve the original author across the new POST. A replacement
// therefore re-stamps every comment on the recreated item with the
// deploying identity. The warning emitted by this modifier makes that
// explicit so the user can choose whether the change is worth the
// audit-attribution loss.
type commentsAppendOnlyRequireReplacePlanModifier struct{}

func (commentsAppendOnlyRequireReplacePlanModifier) Description(_ context.Context) string {
	return "Triggers resource replacement when the comments diff is not pure append-at-end, because Kibana's exception_list items API is append-only and immutable for existing comments."
}

func (m commentsAppendOnlyRequireReplacePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (commentsAppendOnlyRequireReplacePlanModifier) PlanModifyList(
	ctx context.Context,
	req planmodifier.ListRequest,
	resp *planmodifier.ListResponse,
) {
	// Don't act on create (no prior state) or destroy (no plan): the
	// API permits arbitrary comments on create and the resource is going
	// away on destroy.
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}
	// Don't act while the plan value is still unknown — the framework
	// will call us again once it resolves.
	if req.PlanValue.IsUnknown() {
		return
	}

	stateComments := typeutils.ListTypeAs[CommentModel](ctx, req.StateValue, path.Empty(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	planComments := typeutils.ListTypeAs[CommentModel](ctx, req.PlanValue, path.Empty(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// 1. The plan must contain at least every existing comment, in order.
	//    Anything shorter means a remove was attempted (which Kibana would
	//    silently drop) → must replace.
	if len(planComments) < len(stateComments) {
		resp.RequiresReplace = true
		resp.Diagnostics.AddAttributeWarning(req.Path,
			"Comment removal requires resource replacement",
			"Kibana's `PUT /api/exception_lists/items` cannot shrink the "+
				"comments list — omitted entries are silently kept. "+
				"Terraform will destroy and recreate the exception item "+
				"to apply this change. NOTE: all comments on the recreated "+
				"item will be attributed to the API-key owner "+
				"(`created_by`); original per-comment authorship cannot "+
				"be preserved across a recreate via the public API.")
		return
	}

	// 2. Walk through state in order. For each state comment, the plan must
	//    have the same id and the same text at the same position. If id is
	//    Unknown in the plan we treat it as "carried through by
	//    UseStateForUnknown" — fine. If id is Known but doesn't match, the
	//    user has reordered or replaced an existing comment.
	for i, sc := range stateComments {
		pc := planComments[i]

		// Branch on whether the planned id is Known.
		//
		// Known id: must match state's id at the same index, otherwise an
		// existing comment was displaced (reorder/prepend). If ids match,
		// text must match too — a same-id different-text plan is a
		// text-edit attempt, which Kibana silently ignores via PUT.
		//
		// Unknown id: typically a UseStateForUnknown placeholder for an
		// existing comment whose id hasn't been resolved by the framework
		// yet. We accept it as long as the planned text matches state's
		// at the same position. A mismatched text means a new entry was
		// inserted at position i (push-down), which Kibana would reject
		// with HTTP 400.
		if !pc.ID.IsUnknown() && !pc.ID.IsNull() {
			if pc.ID.ValueString() != sc.ID.ValueString() {
				resp.RequiresReplace = true
				resp.Diagnostics.AddAttributeWarning(req.Path,
					"Comment reordering requires resource replacement",
					"Kibana's `PUT /api/exception_lists/items` enforces "+
						"append-only ordering for existing comments and would "+
						"reject this change with HTTP 400. Terraform will "+
						"destroy and recreate the exception item. NOTE: "+
						"`created_by` will be re-stamped from the API-key owner.")
				return
			}
			if pc.Comment.ValueString() != sc.Comment.ValueString() {
				resp.RequiresReplace = true
				resp.Diagnostics.AddAttributeWarning(req.Path,
					"Comment text edit requires resource replacement",
					"Kibana's `PUT /api/exception_lists/items` silently "+
						"ignores text edits on existing comments. Terraform "+
						"will destroy and recreate the exception item to apply "+
						"the new text. NOTE: `created_by` will be re-stamped "+
						"from the API-key owner.")
				return
			}
			continue
		}

		// Unknown id at position i: accept iff the planned text matches
		// state's at the same position. A non-matching text means a new
		// entry was inserted ahead of (or in place of) existing comments
		// — i.e. a prepend/insert that Kibana would reject with HTTP 400.
		if pc.Comment.ValueString() != sc.Comment.ValueString() {
			resp.RequiresReplace = true
			resp.Diagnostics.AddAttributeWarning(req.Path,
				"Comment reordering requires resource replacement",
				"A new comment appears to have been inserted ahead of "+
					"existing entries. Kibana's `PUT /api/exception_lists/items` "+
					"enforces append-only ordering and would reject this change "+
					"with HTTP 400. Terraform will destroy and recreate the "+
					"exception item. NOTE: `created_by` will be re-stamped "+
					"from the API-key owner.")
			return
		}
	}

	// 3. Any plan entries past the length of state must be new comments
	//    (no id / Unknown id). A Known id in this region would mean a
	//    duplicate or out-of-place existing id, which Kibana would reject
	//    or silently drop → must replace.
	for i := len(stateComments); i < len(planComments); i++ {
		pc := planComments[i]
		if !pc.ID.IsUnknown() && !pc.ID.IsNull() && pc.ID.ValueString() != "" {
			resp.RequiresReplace = true
			resp.Diagnostics.AddAttributeWarning(req.Path,
				"Unknown comment id past end of state requires resource replacement",
				"A comment id appears in the plan beyond the prior state's "+
					"length, which is incompatible with Kibana's "+
					"append-only-immutable comments API. Terraform will "+
					"destroy and recreate the exception item.")
			return
		}
	}

	// Pure append-at-end (zero or more new entries without ids past the
	// state's length): no replacement needed.
}
