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

package osquery

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	PlatformLinux   = "linux"
	PlatformDarwin  = "darwin"
	PlatformWindows = "windows"

	AttrECSMappingField  = "field"
	AttrECSMappingValue  = "value"
	AttrECSMappingValues = "values"
)

var (
	PlatformValues = []string{PlatformLinux, PlatformDarwin, PlatformWindows}

	ECSMappingAttrTypes = map[string]attr.Type{
		AttrECSMappingField:  types.StringType,
		AttrECSMappingValue:  types.StringType,
		AttrECSMappingValues: types.SetType{ElemType: types.StringType},
	}
)

type ECSMapping struct {
	Field  types.String `tfsdk:"field"`
	Value  types.String `tfsdk:"value"`
	Values types.Set    `tfsdk:"values"`
}

func ECSMappingElemType() attr.Type {
	return types.ObjectType{AttrTypes: ECSMappingAttrTypes}
}

func PlatformSetFromAPI(platform *kbapi.SecurityOsqueryAPIPlatform) types.Set {
	if platform == nil || strings.TrimSpace(*platform) == "" {
		return types.SetNull(types.StringType)
	}

	parts := strings.Split(*platform, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}

	sort.Strings(values)
	return StringSetValue(values)
}

func PlatformToAPI(ctx context.Context, platform types.Set) (*kbapi.SecurityOsqueryAPIPlatform, diag.Diagnostics) {
	if !typeutils.IsKnown(platform) || platform.IsNull() {
		return nil, nil
	}

	var values []string
	diags := platform.ElementsAs(ctx, &values, false)
	if diags.HasError() {
		return nil, diags
	}

	if len(values) == 0 {
		return nil, diags
	}

	sort.Strings(values)
	joined := strings.Join(values, ",")
	return &joined, diags
}

func ECSMappingMapFromAPI(api *kbapi.SecurityOsqueryAPIECSMapping) (types.Map, diag.Diagnostics) {
	if api == nil || len(*api) == 0 {
		return types.MapNull(ECSMappingElemType()), nil
	}

	elems := make(map[string]attr.Value, len(*api))
	var diags diag.Diagnostics
	for key, item := range *api {
		mapping, mappingDiags := ECSMappingFromAPIType(key, item)
		diags.Append(mappingDiags...)
		if diags.HasError() {
			return types.MapNull(ECSMappingElemType()), diags
		}

		obj, objDiags := types.ObjectValue(ECSMappingAttrTypes, map[string]attr.Value{
			AttrECSMappingField:  mapping.Field,
			AttrECSMappingValue:  mapping.Value,
			AttrECSMappingValues: mapping.Values,
		})
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.MapNull(ECSMappingElemType()), diags
		}
		elems[key] = obj
	}

	mapping, mapDiags := types.MapValue(ECSMappingElemType(), elems)
	diags.Append(mapDiags...)
	return mapping, diags
}

func ECSMappingMapToAPI(ctx context.Context, mapping types.Map) (*kbapi.SecurityOsqueryAPIECSMapping, diag.Diagnostics) {
	if !typeutils.IsKnown(mapping) || mapping.IsNull() {
		return nil, nil
	}

	var diags diag.Diagnostics
	elems := make(kbapi.SecurityOsqueryAPIECSMapping, len(mapping.Elements()))
	for key, av := range mapping.Elements() {
		var m ECSMapping
		d := av.(basetypes.ObjectValue).As(ctx, &m, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		item, d := m.ToAPIType()
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		elems[key] = item
	}

	if len(elems) == 0 {
		return nil, nil
	}

	return &elems, diags
}

func ECSMappingFromAPIType(key string, item kbapi.SecurityOsqueryAPIECSMappingItem) (ECSMapping, diag.Diagnostics) {
	result := ECSMapping{
		Field:  types.StringNull(),
		Value:  types.StringNull(),
		Values: types.SetNull(types.StringType),
	}

	if item.Value != nil {
		if item.Field != nil {
			return result, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid ECS mapping",
					fmt.Sprintf("ecs_mapping[%q]: API returned both field and value", key),
				),
			}
		}

		if scalar, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue0(); err == nil {
			result.Value = types.StringValue(scalar)
			return result, nil
		}

		if values, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue1(); err == nil {
			sorted := append([]string(nil), values...)
			sort.Strings(sorted)
			result.Values = StringSetValue(sorted)
			return result, nil
		}

		return result, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid ECS mapping value",
				fmt.Sprintf("ecs_mapping[%q]: API value is not a string or string array", key),
			),
		}
	}

	if item.Field != nil {
		result.Field = types.StringValue(*item.Field)
	}

	return result, nil
}

func (m ECSMapping) ToAPIType() (kbapi.SecurityOsqueryAPIECSMappingItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	fieldSet := typeutils.IsKnown(m.Field)
	valueSet := typeutils.IsKnown(m.Value)
	valuesSet := typeutils.IsKnown(m.Values)

	setCount := 0
	if fieldSet {
		setCount++
	}
	if valueSet {
		setCount++
	}
	if valuesSet {
		setCount++
	}

	if setCount != 1 {
		diags.AddError(
			"Invalid ecs_mapping element",
			"Exactly one of field, value, or values must be set per ecs_mapping element.",
		)
		return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
	}

	item := kbapi.SecurityOsqueryAPIECSMappingItem{}
	switch {
	case fieldSet:
		item.Field = m.Field.ValueStringPointer()
	case valueSet:
		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := value.FromSecurityOsqueryAPIECSMappingItemValue0(m.Value.ValueString()); err != nil {
			diags.AddError("Invalid ecs_mapping element", fmt.Sprintf("Failed to encode scalar value: %s", err))
			return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
		}
		item.Value = &value
	case valuesSet:
		var values []string
		for _, element := range m.Values.Elements() {
			if str, ok := element.(types.String); ok && typeutils.IsKnown(str) {
				values = append(values, str.ValueString())
			}
		}
		sort.Strings(values)

		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := value.FromSecurityOsqueryAPIECSMappingItemValue1(values); err != nil {
			diags.AddError("Invalid ecs_mapping element", fmt.Sprintf("Failed to encode array values: %s", err))
			return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
		}
		item.Value = &value
	}

	return item, diags
}

func StringSetValue(values []string) types.Set {
	if len(values) == 0 {
		return types.SetNull(types.StringType)
	}

	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}

	return types.SetValueMust(types.StringType, elements)
}
