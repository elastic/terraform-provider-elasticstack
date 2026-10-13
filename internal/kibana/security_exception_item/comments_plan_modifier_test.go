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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// The plan modifier reads required_comments and comments from the plan/state
// objects, so the test object type carries exactly those attributes.
var commentsPlanModifierObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"required_comments": tftypes.List{ElementType: tftypes.String},
		"comments": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"id":         tftypes.String,
			"comment":    tftypes.String,
			"created_at": tftypes.String,
			"created_by": tftypes.String,
		}}},
	},
}

var commentsPlanModifierTestSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"required_comments": schema.ListAttribute{Optional: true, ElementType: types.StringType},
		"comments": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id":         schema.StringAttribute{Computed: true},
					"comment":    schema.StringAttribute{Computed: true},
					"created_at": schema.StringAttribute{Computed: true},
					"created_by": schema.StringAttribute{Computed: true},
				},
			},
		},
	},
}

func storedCommentsList(t *testing.T, texts ...string) types.List {
	t.Helper()
	comments := make([]attr.Value, 0, len(texts))
	for _, text := range texts {
		obj, diags := types.ObjectValue(getCommentAttrTypes(), map[string]attr.Value{
			"id":         types.StringValue("id-" + text),
			"comment":    types.StringValue(text),
			"created_at": types.StringValue("2026-10-12T09:30:00.000Z"),
			"created_by": types.StringValue("analyst"),
		})
		require.False(t, diags.HasError())
		comments = append(comments, obj)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: getCommentAttrTypes()}, comments)
	require.False(t, diags.HasError())
	return list
}

func toTerraformValue(ctx context.Context, t *testing.T, v attr.Value) tftypes.Value {
	t.Helper()
	raw, err := v.ToTerraformValue(ctx)
	require.NoError(t, err)
	return raw
}

func runCommentsPlanModifier(t *testing.T, required types.List, storedComments types.List) *planmodifier.ListResponse {
	t.Helper()
	ctx := context.Background()

	rawAttrs := map[string]tftypes.Value{
		"required_comments": toTerraformValue(ctx, t, required),
		"comments":          toTerraformValue(ctx, t, storedComments),
	}
	plan := tfsdk.Plan{Schema: commentsPlanModifierTestSchema, Raw: tftypes.NewValue(commentsPlanModifierObjectType, rawAttrs)}
	state := tfsdk.State{Schema: commentsPlanModifierTestSchema, Raw: tftypes.NewValue(commentsPlanModifierObjectType, rawAttrs)}

	resp := &planmodifier.ListResponse{PlanValue: storedComments}
	for _, m := range commentsPlanModifiers() {
		m.PlanModifyList(ctx, planmodifier.ListRequest{
			Plan:        plan,
			State:       state,
			Config:      tfsdk.Config{Raw: plan.Raw},
			PlanValue:   storedComments,
			StateValue:  storedComments,
			ConfigValue: storedComments,
			Path:        path.Root("comments"),
		}, resp)
	}
	return resp
}

func requireCommentsUnknown(t *testing.T, required types.List, stored types.List, name string) {
	t.Helper()
	resp := runCommentsPlanModifier(t, required, stored)
	require.False(t, resp.Diagnostics.HasError(), "%s: %s", name, resp.Diagnostics)
	require.True(t, resp.PlanValue.IsUnknown(), "%s: comments must be planned as unknown", name)
}

func requireCommentsUnchanged(t *testing.T, required types.List, stored types.List, name string) {
	t.Helper()
	resp := runCommentsPlanModifier(t, required, stored)
	require.False(t, resp.Diagnostics.HasError(), "%s: %s", name, resp.Diagnostics)
	require.False(t, resp.PlanValue.IsUnknown(), "%s: comments must stay unchanged from prior state", name)
}

func TestCommentsPlanModifier_UnknownWhenRequiredTextWillBeAppended(t *testing.T) {
	t.Parallel()

	requireCommentsUnknown(t,
		types.ListValueMust(types.StringType, []attr.Value{types.StringValue("alpha"), types.StringValue("beta")}),
		storedCommentsList(t, "alpha"),
		"new required text not previously present",
	)
}

func TestCommentsPlanModifier_UnchangedWhenNothingWillBeAppended(t *testing.T) {
	t.Parallel()

	requireCommentsUnchanged(t,
		types.ListValueMust(types.StringType, []attr.Value{types.StringValue("alpha")}),
		storedCommentsList(t, "alpha", "beta"),
		"required list shrunk but every required text is stored",
	)
}

func TestCommentsPlanModifier_UnknownWhenRequiredCommentsUnknown(t *testing.T) {
	t.Parallel()

	requireCommentsUnknown(t,
		types.ListUnknown(types.StringType),
		storedCommentsList(t, "alpha"),
		"required_comments unknown at plan time",
	)
}

func TestCommentsPlanModifier_UnchangedWhenKnownRequiredTextsReordered(t *testing.T) {
	t.Parallel()

	requireCommentsUnchanged(t,
		types.ListValueMust(types.StringType, []attr.Value{types.StringValue("beta"), types.StringValue("alpha")}),
		storedCommentsList(t, "alpha", "beta"),
		"required texts reordered with nothing missing",
	)
}

func TestCommentsPlanModifier_KnownNullRequiredCommentsKeepsComments(t *testing.T) {
	t.Parallel()

	requireCommentsUnchanged(t,
		types.ListNull(types.StringType),
		storedCommentsList(t, "alpha"),
		"null required_comments names nothing missing",
	)
}
