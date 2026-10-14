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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func commentsPlanModifiers() []planmodifier.List {
	return []planmodifier.List{unknownWhenRequiredCommentMissing{}}
}

type unknownWhenRequiredCommentMissing struct{}

func (unknownWhenRequiredCommentMissing) Description(_ context.Context) string {
	return "marks comments unknown when an update will append a missing required comment"
}

func (unknownWhenRequiredCommentMissing) MarkdownDescription(ctx context.Context) string {
	return unknownWhenRequiredCommentMissing{}.Description(ctx)
}

func (m unknownWhenRequiredCommentMissing) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.State.Raw.IsNull() {
		return
	}

	var required types.List
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("required_comments"), &required)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if required.IsUnknown() {
		resp.PlanValue = types.ListUnknown(types.ObjectType{AttrTypes: getCommentAttrTypes()})
		return
	}

	requiredTexts, _ := requiredCommentTexts(required)
	if len(requiredTexts) == 0 {
		return
	}

	storedTexts := storedCommentTexts(ctx, req.StateValue, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(missingCommentTexts(requiredTexts, storedTexts)) > 0 {
		resp.PlanValue = types.ListUnknown(types.ObjectType{AttrTypes: getCommentAttrTypes()})
	}
}

func storedCommentTexts(ctx context.Context, comments types.List, diags *diag.Diagnostics) []string {
	models := typeutils.ListTypeAs[CommentModel](ctx, comments, path.Root("comments"), diags)
	if diags.HasError() {
		return nil
	}
	texts := make([]string, len(models))
	for i, comment := range models {
		texts[i] = comment.Comment.ValueString()
	}
	return texts
}
