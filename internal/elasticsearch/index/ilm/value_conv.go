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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func phaseMapToListValue(ctx context.Context, phaseName string, data map[string]any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	ot := phaseObjectType(phaseName)
	attrs, d := phaseDataToObjectAttrs(ctx, ot, data)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListUnknown(ot), diags
	}
	obj, d := types.ObjectValue(ot.AttrTypes, attrs)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListUnknown(ot), diags
	}
	listVal, d := types.ListValueFrom(ctx, ot, []attr.Value{obj})
	diags.Append(d...)
	return listVal, diags
}

func phaseDataToObjectAttrs(ctx context.Context, ot types.ObjectType, data map[string]any) (map[string]attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := make(map[string]attr.Value)
	for k, elemT := range ot.AttrTypes {
		raw, ok := data[k]
		if !ok || raw == nil {
			attrs[k] = nullValueForType(elemT)
			continue
		}
		v, d := anyToAttr(ctx, elemT, raw)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		attrs[k] = v
	}
	return attrs, diags
}

func nullValueForType(t attr.Type) attr.Value {
	if lt, ok := t.(types.ListType); ok {
		return types.ListNull(lt.ElemType)
	}
	if t.Equal(types.StringType) {
		return types.StringNull()
	}
	if t.Equal(types.Int64Type) {
		return types.Int64Null()
	}
	if t.Equal(types.BoolType) {
		return types.BoolNull()
	}
	if _, ok := t.(jsontypes.NormalizedType); ok {
		return jsontypes.NewNormalizedNull()
	}
	if ot, ok := t.(types.ObjectType); ok {
		return types.ObjectNull(ot.AttrTypes)
	}
	return types.StringNull()
}

func anyToAttr(ctx context.Context, t attr.Type, raw any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	if t.Equal(types.StringType) {
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected string, got %T", raw))
			return types.StringUnknown(), diags
		}
		return types.StringValue(s), diags
	}
	if t.Equal(types.Int64Type) {
		n, ok := coerceInt64(raw)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected number, got %T", raw))
			return types.Int64Unknown(), diags
		}
		return types.Int64Value(n), diags
	}
	if t.Equal(types.BoolType) {
		b, ok := raw.(bool)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected bool, got %T", raw))
			return types.BoolUnknown(), diags
		}
		return types.BoolValue(b), diags
	}
	if ty, ok := t.(types.ListType); ok {
		slice, ok := raw.([]any)
		if !ok || len(slice) == 0 {
			return types.ListNull(ty.ElemType), diags
		}
		elemOT, ok := ty.ElemType.(types.ObjectType)
		if !ok {
			diags.AddError("Internal error", "list element must be object")
			return types.ListUnknown(ty.ElemType), diags
		}
		m, ok := slice[0].(map[string]any)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected object map, got %T", slice[0]))
			return types.ListUnknown(ty.ElemType), diags
		}
		innerAttrs, d := phaseDataToObjectAttrs(ctx, elemOT, m)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListUnknown(ty.ElemType), diags
		}
		obj, d := types.ObjectValue(elemOT.AttrTypes, innerAttrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListUnknown(ty.ElemType), diags
		}
		lv, d := types.ListValueFrom(ctx, ty.ElemType, []attr.Value{obj})
		diags.Append(d...)
		return lv, diags
	}
	if _, ok := t.(jsontypes.NormalizedType); ok {
		s, ok := raw.(string)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected JSON string, got %T", raw))
			return jsontypes.NewNormalizedUnknown(), diags
		}
		return jsontypes.NewNormalizedValue(s), diags
	}
	if ty, ok := t.(types.ObjectType); ok {
		m, ok := raw.(map[string]any)
		if !ok {
			diags.AddError("Type mismatch", fmt.Sprintf("expected map, got %T", raw))
			return types.ObjectUnknown(ty.AttrTypes), diags
		}
		innerAttrs, d := phaseDataToObjectAttrs(ctx, ty, m)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(ty.AttrTypes), diags
		}
		ov, d := types.ObjectValue(ty.AttrTypes, innerAttrs)
		diags.Append(d...)
		return ov, diags
	}
	diags.AddError("Internal error", fmt.Sprintf("unsupported attr type %T", t))
	return types.StringUnknown(), diags
}

func coerceInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}
