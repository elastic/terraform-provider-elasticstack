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

// TODO(dashboard-panel-contract): improve navigateStructByTFSegments so raw API fixture.config shapes that flatten
// discriminated unions (for example slo_overview single/group fields under config.*) map to terraform single/groups nesting,
// eliminating the need for Config.OmitRequiredLeafPresence on those panels.

package contracttest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func navigateStructByTFSegments(root reflect.Value, segments []string) (reflect.Value, bool) {
	if len(segments) == 0 || !root.IsValid() {
		return reflect.Value{}, false
	}
	rv := root
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return reflect.Value{}, false
		}
		rv = rv.Elem()
	}
	for _, seg := range segments {
		if rv.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		idx, ok := fieldIndexByTfsdk(rv.Type(), seg)
		if !ok {
			return reflect.Value{}, false
		}
		rv = rv.Field(idx)
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return reflect.Value{}, false
			}
			rv = rv.Elem()
		}
	}
	return rv, rv.IsValid()
}

func fieldIndexByTfsdk(t reflect.Type, want string) (int, bool) {
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Tag.Get("tfsdk") == want {
			return i, true
		}
	}
	return 0, false
}

func allocPathParents(root reflect.Value, segments []string) (reflect.Value, bool) {
	if len(segments) == 0 || !root.IsValid() || !root.CanSet() || root.Kind() != reflect.Pointer {
		return reflect.Value{}, false
	}
	if root.IsNil() {
		root.Set(reflect.New(root.Type().Elem()))
	}
	cur := root.Elem()

	if len(segments) == 1 {
		return cur, cur.Kind() == reflect.Struct
	}
	for si := range len(segments) - 1 {
		seg := segments[si]
		if cur.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		idx, ok := fieldIndexByTfsdk(cur.Type(), seg)
		if !ok {
			return reflect.Value{}, false
		}
		fv := cur.Field(idx)
		if fv.Kind() == reflect.Pointer && fv.Type().Elem().Kind() == reflect.Struct && fv.CanSet() {
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			cur = fv.Elem()
			continue
		}
		if fv.Kind() == reflect.Struct {
			cur = fv
			continue
		}
		return reflect.Value{}, false
	}

	return cur, cur.Kind() == reflect.Struct
}

func modelLeafReflect(pm *models.PanelModel, block string, segments []string) (reflect.Value, bool) {
	if pm == nil {
		return reflect.Value{}, false
	}
	rt := reflect.ValueOf(pm).Elem()
	idx, ok := fieldIndexByTfsdk(rt.Type(), block)
	if !ok {
		return reflect.Value{}, false
	}
	fv := rt.Field(idx)
	got, navigated := navigateStructByTFSegments(fv, segments)
	return got, navigated && got.IsValid()
}

func writableModelLeaf(pm *models.PanelModel, block string, segments []string) (reflect.Value, bool) {
	rt := reflect.ValueOf(pm).Elem()
	idx, ok := fieldIndexByTfsdk(rt.Type(), block)
	if !ok {
		return reflect.Value{}, false
	}
	ptrField := rt.Field(idx)
	if ptrField.Kind() != reflect.Pointer || !ptrField.CanSet() {
		return reflect.Value{}, false
	}
	if ptrField.IsNil() {
		ptrField.Set(reflect.New(ptrField.Type().Elem()))
	}
	parentStruct, navigated := allocPathParents(ptrField, segments)
	if !navigated {
		return reflect.Value{}, false
	}
	leafSeg := segments[len(segments)-1]
	fidx, ok := fieldIndexByTfsdk(parentStruct.Type(), leafSeg)
	if !ok {
		return reflect.Value{}, false
	}
	dest := parentStruct.Field(fidx)
	return dest, dest.CanSet()
}

func reflectZeroModelLeaf(pm *models.PanelModel, block string, segments []string) bool {
	dest, ok := writableModelLeaf(pm, block, segments)
	if !ok || !dest.CanSet() {
		return false
	}
	dest.Set(reflect.Zero(dest.Type()))
	return true
}

func setStructLeaf(pm *models.PanelModel, blockName string, segments []string, val attr.Value) {
	rt := reflect.ValueOf(pm).Elem()
	idx, ok := fieldIndexByTfsdk(rt.Type(), blockName)
	if !ok {
		return
	}
	ptrField := rt.Field(idx)
	if ptrField.Kind() != reflect.Pointer || !ptrField.CanSet() {
		return
	}
	if ptrField.IsNil() {
		ptrField.Set(reflect.New(ptrField.Type().Elem()))
	}
	parentStruct, navigated := allocPathParents(ptrField, segments)
	if !navigated {
		return
	}
	leafSeg := segments[len(segments)-1]
	fidx, ok := fieldIndexByTfsdk(parentStruct.Type(), leafSeg)
	if !ok {
		return
	}
	dest := parentStruct.Field(fidx)
	if !dest.CanSet() {
		return
	}

	switch vv := val.(type) {
	case types.String:
		dest.Set(reflect.ValueOf(vv))
	case types.Bool:
		dest.Set(reflect.ValueOf(vv))
	case types.Float64:
		dest.Set(reflect.ValueOf(vv))
	case types.Int64:
		dest.Set(reflect.ValueOf(vv))
	default:
		if reflect.TypeOf(val).AssignableTo(dest.Type()) {
			dest.Set(reflect.ValueOf(val))
		}
	}
}

func fixtureHasTerraformNestedKey(cfg map[string]any, tfPath []string) bool {
	_, ok := navigateFixtureConfigByTerraformPath(cfg, tfPath)
	return ok
}

func navigateFixtureConfigByTerraformPath(cur any, tfPath []string) (any, bool) {
	if len(tfPath) == 0 {
		return cur, true
	}

	if arr, ok := cur.([]any); ok && len(arr) > 0 {
		if fm, ok := arr[0].(map[string]any); ok {
			cur = fm
		}
	}

	m, ok := cur.(map[string]any)
	if !ok {
		return nil, false
	}

	seg := tfPath[0]
	var next any
	for _, cand := range []string{tfAttrToAPICamel(seg), seg} {
		if v, exists := m[cand]; exists {
			next = v
			break
		}
	}
	if next == nil {
		return nil, false
	}
	return navigateFixtureConfigByTerraformPath(next, tfPath[1:])
}

func rawFixtureScalarAtConfig(cfg map[string]any, tfAttrName string) (any, bool) {
	for _, cand := range []string{tfAttrToAPICamel(tfAttrName), tfAttrName} {
		if raw, exists := cfg[cand]; exists {
			return raw, true
		}
	}
	return nil, false
}

func parseFixtureRoot(fixtureJSON string) (map[string]any, error) {
	var root map[string]any
	if err := json.Unmarshal([]byte(fixtureJSON), &root); err != nil {
		return nil, err
	}
	return root, nil
}

func parseFixtureConfig(fixtureJSON string) (map[string]any, error) {
	root, err := parseFixtureRoot(fixtureJSON)
	if err != nil {
		return nil, err
	}
	cfg, ok := root["config"].(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}
	return cfg, nil
}

func deleteSkipFields(root any, skips []string) {
	for _, s := range skips {
		deletePath(root, strings.Split(s, "."))
	}
}

func deletePath(root any, dotted []string) {
	if root == nil || len(dotted) == 0 {
		return
	}
	switch cur := root.(type) {
	case map[string]any:
		head := dotted[0]
		if len(dotted) == 1 {
			delete(cur, head)
			return
		}
		deletePath(cur[head], dotted[1:])
	case []any:
		idx, err := strconv.Atoi(dotted[0])
		if err != nil || idx < 0 || idx >= len(cur) {
			return
		}
		deletePath(cur[idx], dotted[1:])
	default:
		return
	}
}

func attrsComparableEqual(a, b attr.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	as, ok1 := a.(types.String)
	bs, ok2 := b.(types.String)
	if ok1 && ok2 {
		switch {
		case as.IsUnknown() || bs.IsUnknown():
			return true
		case as.IsNull() && bs.IsNull():
			return true
		}
		return as.Equal(bs)
	}
	ab, ok1 := a.(types.Bool)
	bb, ok2 := b.(types.Bool)
	if ok1 && ok2 {
		switch {
		case ab.IsUnknown() || bb.IsUnknown():
			return true
		case ab.IsNull() && bb.IsNull():
			return true
		}
		return ab.Equal(bb)
	}
	aF, ok1 := a.(types.Float64)
	bF, ok2 := b.(types.Float64)
	if ok1 && ok2 {
		switch {
		case aF.IsUnknown() || bF.IsUnknown():
			return true
		case aF.IsNull() && bF.IsNull():
			return true
		}
		return aF.Equal(bF)
	}
	aI, ok1 := a.(types.Int64)
	bI, ok2 := b.(types.Int64)
	if ok1 && ok2 {
		switch {
		case aI.IsUnknown() || bI.IsUnknown():
			return true
		case aI.IsNull() && bI.IsNull():
			return true
		}
		return aI.Equal(bI)
	}
	return a.Equal(b)
}

func clonePanel(pm *models.PanelModel) (*models.PanelModel, error) {
	if pm == nil {
		return nil, fmt.Errorf("nil panel")
	}
	b, err := json.Marshal(pm)
	if err != nil {
		return nil, err
	}
	var out models.PanelModel
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
