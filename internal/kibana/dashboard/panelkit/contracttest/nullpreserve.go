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

package contracttest

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appendNullPreserveIssues(ctx context.Context, handler iface.Handler, fixture string, skipFields []string, issues *[]string) {
	block := handler.PanelType() + "_config"
	if !panelkit.HasPanelConfigBlock(block) {
		return
	}
	sna, ok := handler.SchemaAttribute().(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	lp := collectLeafPaths(sna)

	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[NullPreserve] parse: %v", err))
		return
	}

	var baseline models.PanelModel
	if diags := handler.FromAPI(ctx, &baseline, nil, item0); diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[NullPreserve] baseline FromAPI: %s", summarizeDiags(diags)))
		return
	}
	if !panelkit.HasConfig(&baseline, block) {
		return
	}

	for _, leaf := range lp.optional {
		dotted := strings.Join(leaf, ".")
		if slices.Contains(skipFields, dotted) || skipHasPrefix(skipFields, dotted) {
			continue
		}
		schemaLeaf, schOK := schemaAttributeAt(sna, leaf)
		if !schOK {
			*issues = append(*issues, fmt.Sprintf("[NullPreserve] dotted=%s: schema path not found", dotted))
			continue
		}
		blVal, leafOK := modelLeafReflect(&baseline, block, leaf)
		if !leafOK {
			*issues = append(*issues, fmt.Sprintf("[NullPreserve] dotted=%s: model path not readable on baseline state", dotted))
			continue
		}

		// Fresh import versus baseline.
		label := "[NullPreserve] fresh_import/" + dotted
		assertFromAPI(ctx, handler, fixture, issues, label, nil,
			func(out *models.PanelModel) bool {
				got, ok := modelLeafReflect(out, block, leaf)
				if !ok {
					return false
				}
				return modelLeafDeepEqual(blVal, got)
			})

		// prior_null: plant typed null/zero at leaf.
		nullPrior := clonePanelBaselineWithNullLeaf(&baseline, block, leaf, schemaLeaf)
		if nullPrior != nil {
			assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] prior_null/"+dotted, nullPrior,
				func(out *models.PanelModel) bool {
					got, ok := modelLeafReflect(out, block, leaf)
					if !ok {
						return false
					}
					switch schemaLeaf.(type) {
					case schema.StringAttribute, schema.BoolAttribute, schema.Float64Attribute, schema.Int64Attribute:
						if av, conv := attrFromReflectLeaf(got); conv {
							if s, ok := av.(types.String); ok {
								return s.IsNull()
							}
							if b, ok := av.(types.Bool); ok {
								return b.IsNull()
							}
							if f, ok := av.(types.Float64); ok {
								return f.IsNull()
							}
							if i, ok := av.(types.Int64); ok {
								return i.IsNull()
							}
						}
						return reflect.ValueOf(reflect.Zero(got.Type()).Interface()).IsZero()
					default:
						return sliceOrMapLooksUnset(got)
					}
				})
		}

		if priorKnownApplicable(schemaLeaf, leaf) {
			switch schemaLeaf.(type) {
			case schema.StringAttribute, schema.BoolAttribute, schema.Float64Attribute, schema.Int64Attribute:
				if p := clonePanelStaleScalarLeaf(&baseline, block, leaf, schemaLeaf); p != nil {
					assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] prior_known/"+dotted, p,
						func(out *models.PanelModel) bool {
							got, ok := modelLeafReflect(out, block, leaf)
							if !ok {
								return false
							}
							gav, gc := attrFromReflectLeaf(got)
							bav, bc := attrFromReflectLeaf(blVal)
							if gc && bc {
								return attrsComparableEqual(bav, gav)
							}
							return modelLeafDeepEqual(blVal, got)
						})
				}
			case schema.ListNestedAttribute, schema.ListAttribute, schema.MapAttribute:
				if p := clonePanelMirroringBaselineSliceLeaf(&baseline, block, leaf, schemaLeaf); p != nil {
					assertFromAPI(ctx, handler, fixture, issues, "[NullPreserve] prior_known/"+dotted, p,
						func(out *models.PanelModel) bool {
							got, ok := modelLeafReflect(out, block, leaf)
							if !ok {
								return false
							}
							return modelLeafDeepEqual(blVal, got)
						})
				}
			}
		}
	}
}

func sliceOrMapLooksUnset(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return sliceOrMapLooksUnset(v.Elem())
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

func priorKnownApplicable(sch schema.Attribute, leaf []string) bool {
	switch sch.(type) {
	case schema.StringAttribute, schema.BoolAttribute, schema.Float64Attribute, schema.Int64Attribute:
		return true
	case schema.ListNestedAttribute, schema.ListAttribute, schema.MapAttribute:
		return len(leaf) > 0
	default:
		_ = leaf
		return false
	}
}

func clonePanelBaselineWithNullLeaf(baseline *models.PanelModel, block string, leaf []string, sch schema.Attribute) *models.PanelModel {
	p, err := clonePanel(baseline)
	if err != nil || p == nil {
		return nil
	}
	p.Type = baseline.Type
	panelkit.EnsureMutableTypedConfig(p, block)
	switch sch.(type) {
	case schema.StringAttribute:
		setStructLeaf(p, block, leaf, types.StringNull())
	case schema.BoolAttribute:
		setStructLeaf(p, block, leaf, types.BoolNull())
	case schema.Float64Attribute:
		setStructLeaf(p, block, leaf, types.Float64Null())
	case schema.Int64Attribute:
		setStructLeaf(p, block, leaf, types.Int64Null())
	case schema.ListNestedAttribute, schema.ListAttribute, schema.MapAttribute:
		if ok := reflectZeroModelLeaf(p, block, leaf); !ok {
			return nil
		}
	default:
		return nil
	}
	return p
}

func clonePanelStaleScalarLeaf(baseline *models.PanelModel, block string, leaf []string, sch schema.Attribute) *models.PanelModel {
	p, err := clonePanel(baseline)
	if err != nil || p == nil {
		return nil
	}
	panelkit.EnsureMutableTypedConfig(p, block)
	p.Type = baseline.Type
	bl, ok := modelLeafReflect(baseline, block, leaf)
	if !ok {
		return nil
	}
	switch sch.(type) {
	case schema.StringAttribute:
		setStructLeaf(p, block, leaf, types.StringValue("stale-prior-contracttest"))
		return p
	case schema.BoolAttribute:
		if b, ok := attrFromReflectLeaf(bl); ok {
			if bv, ok := b.(types.Bool); ok && typeKnownBoolLike(bv) {
				setStructLeaf(p, block, leaf, invertBoolTf(bv))
				return p
			}
		}
		setStructLeaf(p, block, leaf, types.BoolValue(true))
		return p
	case schema.Float64Attribute:
		setStructLeaf(p, block, leaf, types.Float64Value(-9.87654321))
		return p
	case schema.Int64Attribute:
		setStructLeaf(p, block, leaf, types.Int64Value(99988777))
		return p
	default:
		return nil
	}
}

func typeKnownBoolLike(b types.Bool) bool {
	return !b.IsUnknown() && !b.IsNull()
}

func invertBoolTf(b types.Bool) types.Bool {
	if b.IsUnknown() || b.IsNull() {
		return types.BoolNull()
	}
	return types.BoolValue(!b.ValueBool())
}

func clonePanelMirroringBaselineSliceLeaf(baseline *models.PanelModel, block string, leaf []string, sch schema.Attribute) *models.PanelModel {
	switch sch.(type) {
	case schema.ListNestedAttribute, schema.ListAttribute, schema.MapAttribute:
	default:
		return nil
	}
	src, ok := modelLeafReflect(baseline, block, leaf)
	if !ok || !src.IsValid() {
		return nil
	}
	p, err := clonePanel(baseline)
	if err != nil || p == nil {
		return nil
	}
	panelkit.EnsureMutableTypedConfig(p, block)
	p.Type = baseline.Type
	dst, ok := writableModelLeaf(p, block, leaf)
	if !ok {
		return nil
	}
	if src.Kind() == reflect.Slice && dst.CanSet() {
		dup := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		reflect.Copy(dup, src)
		dst.Set(dup)
		return p
	}
	if src.Kind() == reflect.Map && dst.CanSet() {
		dup := reflect.MakeMap(src.Type())
		iter := src.MapRange()
		for iter.Next() {
			dup.SetMapIndex(iter.Key(), iter.Value())
		}
		dst.Set(dup)
		return p
	}
	return nil
}

func assertFromAPI(
	ctx context.Context,
	handler iface.Handler,
	fixture string,
	issues *[]string,
	label string,
	prior *models.PanelModel,
	ok func(*models.PanelModel) bool,
) {
	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("%s: parse: %v", label, err))
		return
	}
	var pm models.PanelModel
	if prior != nil {
		pm = *prior
	}
	diags := handler.FromAPI(ctx, &pm, prior, item0)
	if diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("%s: FromAPI: %s", label, summarizeDiags(diags)))
		return
	}
	if !ok(&pm) {
		*issues = append(*issues, fmt.Sprintf("%s: post-condition failed", label))
	}
}

func skipHasPrefix(skip []string, sk string) bool {
	for _, s := range skip {
		if strings.HasPrefix(sk, s+".") {
			return true
		}
	}
	return false
}

func attrFromReflectLeaf(rv reflect.Value) (attr.Value, bool) {
	if !rv.IsValid() {
		return nil, false
	}
	x := rv.Interface()
	switch v := x.(type) {
	case types.String:
		return v, true
	case types.Bool:
		return v, true
	case types.Float64:
		return v, true
	case types.Int64:
		return v, true
	default:
		return nil, false
	}
}

func modelLeafDeepEqual(a, b reflect.Value) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}
	ax, ay := attrFromReflectLeaf(a)
	bx, by := attrFromReflectLeaf(b)
	if ay && by {
		return attrsComparableEqual(ax, bx)
	}
	return reflect.DeepEqual(a.Interface(), b.Interface())
}
