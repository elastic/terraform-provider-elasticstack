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

package ilm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func phaseListToExpandMap(ctx context.Context, phaseList types.List) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if phaseList.IsNull() || phaseList.IsUnknown() || len(phaseList.Elements()) == 0 {
		return nil, diags
	}
	var objs []types.Object
	diags.Append(phaseList.ElementsAs(ctx, &objs, false)...)
	if diags.HasError() || len(objs) == 0 {
		return nil, diags
	}
	m, d := objectToExpandMap(ctx, objs[0])
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	applyAllocateJSONDefaults(m)
	return m, diags
}

// applyAllocateJSONDefaults mirrors SDK defaults for allocate JSON strings ({}).
func applyAllocateJSONDefaults(phase map[string]any) {
	allocRaw, ok := phase["allocate"]
	if !ok {
		return
	}
	allocList, ok := allocRaw.([]any)
	if !ok || len(allocList) == 0 {
		return
	}
	am, ok := allocList[0].(map[string]any)
	if !ok {
		return
	}
	for _, k := range []string{"include", "exclude", "require"} {
		if _, has := am[k]; !has {
			am[k] = "{}"
		}
	}
}

func objectToExpandMap(ctx context.Context, obj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(map[string]any)
	for k, v := range obj.Attributes() {
		if v.IsNull() || v.IsUnknown() {
			continue
		}
		raw, d := attrValueToExpandRaw(ctx, v)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if raw != nil {
			out[k] = raw
		}
	}
	return out, diags
}

func attrValueToExpandRaw(ctx context.Context, v attr.Value) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch tv := v.(type) {
	case types.String:
		return tv.ValueString(), diags
	case types.Int64:
		return tv.ValueInt64(), diags
	case types.Bool:
		return tv.ValueBool(), diags
	case jsontypes.Normalized:
		if tv.IsNull() {
			return nil, diags
		}
		return tv.ValueString(), diags
	case types.List:
		if len(tv.Elements()) == 0 {
			return nil, diags
		}
		elem := tv.Elements()[0]
		innerObj, ok := elem.(types.Object)
		if !ok {
			diags.AddError("Internal error", "expected object inside list")
			return nil, diags
		}
		m, d := objectToExpandMap(ctx, innerObj)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return []any{m}, diags
	default:
		return nil, diags
	}
}
