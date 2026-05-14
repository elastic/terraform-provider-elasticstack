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

package panelkit

import (
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

var panelModelTfsdkFieldIndex map[string]int

func init() {
	t := reflect.TypeFor[models.PanelModel]()
	panelModelTfsdkFieldIndex = make(map[string]int, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}
		if prev, dup := panelModelTfsdkFieldIndex[tag]; dup {
			panic(fmt.Sprintf(
				"dashboard panel PanelModel reflection: duplicate tfsdk tag %q on fields %s and %s",
				tag, t.Field(prev).Name, f.Name))
		}
		panelModelTfsdkFieldIndex[tag] = i
	}
}

func unknownPanelBlockPanic(blockName string) {
	panic(fmt.Sprintf("panelkit: unknown PanelModel block %q (no matching tfsdk tag on PanelModel)", blockName))
}

// HasPanelConfigBlock reports whether blockName is a tfsdk tag on models.PanelModel.
func HasPanelConfigBlock(blockName string) bool {
	_, ok := panelModelTfsdkFieldIndex[blockName]
	return ok
}

// MustPanelConfigBlockTagged panics if blockName is not a tfsdk tag on models.PanelModel.
func MustPanelConfigBlockTagged(blockName string) {
	if !HasPanelConfigBlock(blockName) {
		panic(fmt.Sprintf("dashboard panel registry: unknown panel config block %q — no PanelModel field with matching tfsdk tag", blockName))
	}
}

// HasConfig reports whether pm has non-nil typed state for blockName (e.g. "slo_burn_rate_config").
func HasConfig(pm *models.PanelModel, blockName string) bool {
	if pm == nil {
		return false
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		return false
	}
	rv := reflect.ValueOf(pm).Elem().Field(idx)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Interface, reflect.Map:
		return !rv.IsNil()
	default:
		return !rv.IsZero()
	}
}

// ClearConfig zeroes pm's typed field for blockName.
func ClearConfig(pm *models.PanelModel, blockName string) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		unknownPanelBlockPanic(blockName)
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() {
		return
	}
	fv.Set(reflect.Zero(fv.Type()))
}

// SetConfig assigns cfg to pm's typed field keyed by blockName. cfg must be assignable to that field type.
// A nil cfg clears the field when the field is pointer-like or otherwise zeroable.
func SetConfig(pm *models.PanelModel, blockName string, cfg any) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		unknownPanelBlockPanic(blockName)
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() {
		return
	}
	if cfg == nil {
		fv.Set(reflect.Zero(fv.Type()))
		return
	}
	cv := reflect.ValueOf(cfg)
	if cv.Type().AssignableTo(fv.Type()) {
		fv.Set(cv)
		return
	}
	if cv.Type().ConvertibleTo(fv.Type()) {
		fv.Set(cv.Convert(fv.Type()))
		return
	}
	panic(fmt.Sprintf("panelkit.SetConfig: value type %s not assignable to field %s (%s)", cv.Type(), blockName, fv.Type()))
}

// EnsureMutableTypedConfig allocates *T for pm's pointer-backed config block at blockName when it is nil.
func EnsureMutableTypedConfig(pm *models.PanelModel, blockName string) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		unknownPanelBlockPanic(blockName)
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() || fv.Kind() != reflect.Pointer {
		return
	}
	if fv.IsNil() {
		fv.Set(reflect.New(fv.Type().Elem()))
	}
}
