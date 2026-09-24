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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// makeCommentsList builds a types.List of comment object values for plan
// modifier tests. Each entry is `{id, comment}`; pass an empty id ("") to
// model an unknown / not-yet-assigned id (typical of brand-new entries
// appended through config).
func makeCommentsList(t *testing.T, entries []struct{ id, comment string }) types.List {
	t.Helper()
	objs := make([]attr.Value, 0, len(entries))
	for _, e := range entries {
		idVal := types.StringValue(e.id)
		if e.id == "" {
			idVal = types.StringUnknown()
		}
		obj, diags := types.ObjectValue(getCommentAttrTypes(), map[string]attr.Value{
			"id":      idVal,
			"comment": types.StringValue(e.comment),
		})
		require.Empty(t, diags)
		objs = append(objs, obj)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: getCommentAttrTypes()}, objs)
	require.Empty(t, diags)
	return list
}

func runModifier(t *testing.T, state, plan types.List) *planmodifier.ListResponse {
	t.Helper()
	resp := &planmodifier.ListResponse{PlanValue: plan}
	commentsAppendOnlyRequireReplacePlanModifier{}.PlanModifyList(
		context.Background(),
		planmodifier.ListRequest{StateValue: state, PlanValue: plan},
		resp,
	)
	return resp
}

// Test_commentsAppendOnly_Create verifies the modifier is inert on a create
// (state is null), allowing any combination of comments through unchanged.
func Test_commentsAppendOnly_Create(t *testing.T) {
	elemType := types.ObjectType{AttrTypes: getCommentAttrTypes()}
	state := types.ListNull(elemType)
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"", "first"}, {"", "second"},
	})

	resp := runModifier(t, state, plan)

	require.False(t, resp.RequiresReplace)
	require.Empty(t, resp.Diagnostics)
}

// Test_commentsAppendOnly_Destroy verifies the modifier is inert on a destroy
// (plan is null), letting the framework destroy the resource without surprises.
func Test_commentsAppendOnly_Destroy(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"},
	})
	elemType := types.ObjectType{AttrTypes: getCommentAttrTypes()}
	plan := types.ListNull(elemType)

	resp := runModifier(t, state, plan)

	require.False(t, resp.RequiresReplace)
	require.Empty(t, resp.Diagnostics)
}

// Test_commentsAppendOnly_NoOp verifies that an identical plan and state
// (same ids in the same order with the same texts) does not require replace.
// This is the steady-state path that should be the most common in practice.
func Test_commentsAppendOnly_NoOp(t *testing.T) {
	entries := []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"}, {"id-c", "third"},
	}
	state := makeCommentsList(t, entries)
	plan := makeCommentsList(t, entries)

	resp := runModifier(t, state, plan)

	require.False(t, resp.RequiresReplace)
	require.Empty(t, resp.Diagnostics)
}

// Test_commentsAppendOnly_AppendAtEnd verifies that adding new entries with
// no id at the END of the list does not require replace. This is the only
// non-no-op operation Kibana's append-only API actually accepts.
func Test_commentsAppendOnly_AppendAtEnd(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"},
	})
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"}, {"", "new third"},
	})

	resp := runModifier(t, state, plan)

	require.False(t, resp.RequiresReplace)
	require.Empty(t, resp.Diagnostics)
}

// Test_commentsAppendOnly_EditText verifies that editing the text of an
// existing comment (same id, different text) triggers replacement, because
// Kibana's PUT silently ignores text edits on existing comments.
func Test_commentsAppendOnly_EditText(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"},
	})
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first EDITED"}, {"id-b", "second"},
	})

	resp := runModifier(t, state, plan)

	require.True(t, resp.RequiresReplace)
	require.Len(t, resp.Diagnostics.Warnings(), 1)
	require.Contains(t, resp.Diagnostics.Warnings()[0].Summary(), "text edit")
}

// Test_commentsAppendOnly_Shrink verifies that omitting an existing entry
// from the plan triggers replacement, because Kibana's PUT cannot shrink the
// comments list (silently keeps all existing entries).
func Test_commentsAppendOnly_Shrink(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"}, {"id-c", "third"},
	})
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"},
	})

	resp := runModifier(t, state, plan)

	require.True(t, resp.RequiresReplace)
	require.Len(t, resp.Diagnostics.Warnings(), 1)
	require.Contains(t, resp.Diagnostics.Warnings()[0].Summary(), "removal")
}

// Test_commentsAppendOnly_Prepend verifies that prepending a new entry (no id
// at position 0, existing ids shifted) triggers replacement, because Kibana's
// PUT rejects this with HTTP 400 ("item comments are append only").
func Test_commentsAppendOnly_Prepend(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"},
	})
	plan := makeCommentsList(t, []struct{ id, comment string }{
		// New entry at index 0 with the existing entries shifted down. Because
		// `id-a` and `id-b` are still Known they end up at indices 1 and 2 of
		// the plan, mismatching their state indices.
		{"", "new prepended"}, {"id-a", "first"}, {"id-b", "second"},
	})

	resp := runModifier(t, state, plan)

	require.True(t, resp.RequiresReplace)
	require.Len(t, resp.Diagnostics.Warnings(), 1)
	require.Contains(t, resp.Diagnostics.Warnings()[0].Summary(), "reordering")
}

// Test_commentsAppendOnly_ReorderExisting verifies that swapping two existing
// entries (no new entries) triggers replacement.
func Test_commentsAppendOnly_ReorderExisting(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"}, {"id-b", "second"},
	})
	// Swap the two existing entries.
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"id-b", "second"}, {"id-a", "first"},
	})

	resp := runModifier(t, state, plan)

	require.True(t, resp.RequiresReplace)
}

// Test_commentsAppendOnly_AppendWithUnknownIdMatchesUseStateForUnknown verifies
// that when a plan entry at an existing index has an Unknown id (the typical
// shape produced by the UseStateForUnknown plan modifier on `comment.id` when
// the framework hasn't resolved it yet), the modifier accepts it as long as
// the text matches state's. This guards against the modifier triggering a
// spurious replace during the first refresh after a provider upgrade.
func Test_commentsAppendOnly_AppendWithUnknownIdMatchesUseStateForUnknown(t *testing.T) {
	state := makeCommentsList(t, []struct{ id, comment string }{
		{"id-a", "first"},
	})
	// id-Unknown matches state's id-a positionally; text matches → not a
	// reorder/edit, just an in-flight plan value.
	plan := makeCommentsList(t, []struct{ id, comment string }{
		{"", "first"},
	})

	resp := runModifier(t, state, plan)

	require.False(t, resp.RequiresReplace)
	require.Empty(t, resp.Diagnostics)
}
