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

package slo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// kqlLegacyStringExclusiveWithObject enforces that the string arm of a KQL union and the
// parallel *_kql object form are not both configured.
type kqlLegacyStringExclusiveWithObject struct {
	parallelObjectAttr      string
	treatEmptyStringAsUnset bool
}

func (v kqlLegacyStringExclusiveWithObject) Description(_ context.Context) string {
	return fmt.Sprintf("mutually exclusive with %s: configure only one representation", v.parallelObjectAttr)
}

func (v kqlLegacyStringExclusiveWithObject) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v kqlLegacyStringExclusiveWithObject) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if v.treatEmptyStringAsUnset {
		if s == "" {
			return
		}
	} else if s == "" {
		return
	}

	var o types.Object
	merge := req.Path.Expression().Merge(path.MatchRelative().AtParent().AtName(v.parallelObjectAttr))
	matched, diags := req.Config.PathMatches(ctx, merge)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() || len(matched) == 0 {
		return
	}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, matched[0], &o)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if o.IsNull() || o.IsUnknown() {
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Configuration",
		fmt.Sprintf("Cannot set both this attribute and %s. Remove one of them.", v.parallelObjectAttr),
	)
}

// kqlObjectFormExclusiveWithString enforces the same rule from the object-form attribute.
type kqlObjectFormExclusiveWithString struct {
	parallelStringAttr      string
	treatEmptyStringAsUnset bool
}

func (v kqlObjectFormExclusiveWithString) Description(_ context.Context) string {
	return fmt.Sprintf("mutually exclusive with %s: configure only one representation", v.parallelStringAttr)
}

func (v kqlObjectFormExclusiveWithString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v kqlObjectFormExclusiveWithString) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	merge := req.Path.Expression().Merge(path.MatchRelative().AtParent().AtName(v.parallelStringAttr))
	matched, diags := req.Config.PathMatches(ctx, merge)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() || len(matched) == 0 {
		return
	}
	var s types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, matched[0], &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if s.IsNull() || s.IsUnknown() {
		return
	}
	if v.treatEmptyStringAsUnset && s.ValueString() == "" {
		return
	}
	if !v.treatEmptyStringAsUnset && s.ValueString() == "" {
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Configuration",
		fmt.Sprintf("Cannot set both %s and this block. Remove one of them.", v.parallelStringAttr),
	)
}
