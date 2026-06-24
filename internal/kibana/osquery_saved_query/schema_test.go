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

package osquerysavedquery

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_attributeMetadata(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	idAttr, ok := s.Attributes["id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, idAttr.IsComputed())
	assert.False(t, idAttr.IsRequired())
	assertHasStringPlanModifier(t, idAttr.PlanModifiers, "useStateForUnknown")

	savedQueryIDAttr, ok := s.Attributes["saved_query_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, savedQueryIDAttr.IsRequired())
	assert.False(t, savedQueryIDAttr.IsComputed())
	assertHasStringPlanModifier(t, savedQueryIDAttr.PlanModifiers, "requiresReplace")

	spaceIDAttr, ok := s.Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.True(t, spaceIDAttr.IsComputed())
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "useStateForUnknown")
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "requiresReplace")

	queryAttr, ok := s.Attributes["query"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, queryAttr.IsRequired())

	snapshotAttr, ok := s.Attributes["snapshot"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, snapshotAttr.IsOptional())
	assert.True(t, snapshotAttr.IsComputed())
	assertHasBoolPlanModifier(t, snapshotAttr.BoolPlanModifiers(), "useStateForUnknown")

	removedAttr, ok := s.Attributes["removed"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, removedAttr.IsOptional())
	assert.True(t, removedAttr.IsComputed())
	assertHasBoolPlanModifier(t, removedAttr.BoolPlanModifiers(), "useStateForUnknown")
}

func TestSchema_spaceIDDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	spaceIDAttr, ok := getSchema(ctx).Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)

	defaultHandler := spaceIDAttr.StringDefaultValue()
	require.NotNil(t, defaultHandler)

	var resp defaults.StringResponse
	defaultHandler.DefaultString(ctx, defaults.StringRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, clients.DefaultSpaceID, resp.PlanValue.ValueString())
}

func TestSchema_platformAllowedValues(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	platformAttr, ok := s.Attributes["platform"].(schema.SetAttribute)
	require.True(t, ok)
	require.NotEmpty(t, platformAttr.Validators)

	for _, value := range osqueryPlatformValues {
		t.Run("valid/"+value, func(t *testing.T) {
			t.Parallel()
			diags := validateSetValidators(context.Background(), platformAttr.Validators, types.SetValueMust(types.StringType, []attr.Value{types.StringValue(value)}), path.Root("platform"))
			assert.False(t, diags.HasError(), "expected %q to be valid: %v", value, diags)
		})
	}

	t.Run("invalid/ios", func(t *testing.T) {
		t.Parallel()
		diags := validateSetValidators(context.Background(), platformAttr.Validators, types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ios")}), path.Root("platform"))
		assert.True(t, diags.HasError())
	})
}

func TestEcsMappingElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	ecsMappingAttr, ok := getSchema(context.Background()).Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok, "expected ecs_mapping to be MapNestedAttribute")

	schemaElem := schemaNestedObjectElemType(ecsMappingAttr.NestedObject)
	require.Equal(t, schemaElem, getEcsMappingElemType(),
		"getEcsMappingElemType() drifted from schema ecs_mapping nested object; update both together")
}

func TestSchema_ecsMappingNestedObjectValidatorsWired(t *testing.T) {
	t.Parallel()

	fromGetSchema, ok := getSchema(context.Background()).Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.NotEmpty(t, fromGetSchema.NestedObject.Validators, "ecs_mapping from getSchema must attach nested object validators")

	fromHelper := ecsMappingSchema()
	require.NotEmpty(t, fromHelper.NestedObject.Validators, "ecsMappingSchema must attach nested object validators")
}

func TestSchema_ecsMappingNestedObjectValidators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nestedObject := requireEcsMappingNestedObjectFromSchema(t)
	elemPath := path.Root("ecs_mapping").AtMapKey("process.name")

	fieldOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringValue("cmdline"),
			attrEcsMappingValue:  types.StringNull(),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
	}
	valueOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringValue("process"),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
	}
	valuesOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField: types.StringNull(),
			attrEcsMappingValue: types.StringNull(),
			attrEcsMappingValues: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("process"),
				types.StringValue("network"),
			}),
		})
	}

	t.Run("accepts field only", func(t *testing.T) {
		t.Parallel()
		diags := validateObjectValidators(ctx, nestedObject.Validators, fieldOnly(), elemPath)
		require.False(t, diags.HasError(), "%s", diags)
	})

	t.Run("accepts value only", func(t *testing.T) {
		t.Parallel()
		diags := validateObjectValidators(ctx, nestedObject.Validators, valueOnly(), elemPath)
		require.False(t, diags.HasError(), "%s", diags)
	})

	t.Run("accepts values only", func(t *testing.T) {
		t.Parallel()
		diags := validateObjectValidators(ctx, nestedObject.Validators, valuesOnly(), elemPath)
		require.False(t, diags.HasError(), "%s", diags)
	})

	t.Run("rejects empty element", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringNull(),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
		diags := validateObjectValidators(ctx, nestedObject.Validators, obj, elemPath)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "Exactly one of")
	})

	t.Run("rejects field and value", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringValue("cmdline"),
			attrEcsMappingValue:  types.StringValue("literal"),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
		diags := validateObjectValidators(ctx, nestedObject.Validators, obj, elemPath)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "not more than one")
	})

	t.Run("rejects field and values", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField: types.StringValue("cmdline"),
			attrEcsMappingValue: types.StringNull(),
			attrEcsMappingValues: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("process"),
			}),
		})
		diags := validateObjectValidators(ctx, nestedObject.Validators, obj, elemPath)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "not more than one")
	})
}

func TestSchema_ecsMappingNestedAttributeValidators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nestedObject := requireEcsMappingNestedObjectFromSchema(t)

	fieldAttr, ok := nestedObject.Attributes[attrEcsMappingField].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, fieldAttr.Validators)

	valueAttr, ok := nestedObject.Attributes[attrEcsMappingValue].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, valueAttr.Validators)

	valuesAttr, ok := nestedObject.Attributes[attrEcsMappingValues].(schema.SetAttribute)
	require.True(t, ok)
	require.NotEmpty(t, valuesAttr.Validators)

	t.Run("rejects empty field", func(t *testing.T) {
		t.Parallel()
		diags := validateStringValidators(ctx, fieldAttr.Validators, types.StringValue(""), path.Root("ecs_mapping").AtMapKey("k").AtName(attrEcsMappingField))
		require.True(t, diags.HasError())
	})

	t.Run("rejects empty value", func(t *testing.T) {
		t.Parallel()
		diags := validateStringValidators(ctx, valueAttr.Validators, types.StringValue(""), path.Root("ecs_mapping").AtMapKey("k").AtName(attrEcsMappingValue))
		require.True(t, diags.HasError())
	})

	t.Run("rejects empty values set", func(t *testing.T) {
		t.Parallel()
		diags := validateSetValidators(ctx, valuesAttr.Validators, types.SetValueMust(types.StringType, []attr.Value{}), path.Root("ecs_mapping").AtMapKey("k").AtName(attrEcsMappingValues))
		require.True(t, diags.HasError())
	})
}

func requireEcsMappingNestedObjectFromSchema(t *testing.T) schema.NestedAttributeObject {
	t.Helper()

	ecsMappingAttr, ok := getSchema(context.Background()).Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.NotEmpty(t, ecsMappingAttr.NestedObject.Validators)
	return ecsMappingAttr.NestedObject
}

func validateObjectValidators(ctx context.Context, validators []validator.Object, obj types.Object, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	req := validator.ObjectRequest{ConfigValue: obj, Path: p}
	for _, v := range validators {
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, req, &resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func validateStringValidators(ctx context.Context, validators []validator.String, value types.String, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	req := validator.StringRequest{ConfigValue: value, Path: p}
	for _, v := range validators {
		var resp validator.StringResponse
		v.ValidateString(ctx, req, &resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func validateSetValidators(ctx context.Context, validators []validator.Set, value types.Set, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	req := validator.SetRequest{ConfigValue: value, Path: p}
	for _, v := range validators {
		var resp validator.SetResponse
		v.ValidateSet(ctx, req, &resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func schemaNestedObjectElemType(no schema.NestedAttributeObject) attr.Type {
	attrTypes := make(map[string]attr.Type, len(no.Attributes))
	for name, a := range no.Attributes {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}

func assertHasStringPlanModifier(t *testing.T, modifiers []planmodifier.String, suffix string) {
	t.Helper()
	assertHasPlanModifierType(t, len(modifiers), func(i int) string {
		return reflect.TypeOf(modifiers[i]).String()
	}, suffix)
}

func assertHasBoolPlanModifier(t *testing.T, modifiers []planmodifier.Bool, suffix string) {
	t.Helper()
	assertHasPlanModifierType(t, len(modifiers), func(i int) string {
		return reflect.TypeOf(modifiers[i]).String()
	}, suffix)
}

func assertHasPlanModifierType(t *testing.T, count int, typeAt func(int) string, suffix string) {
	t.Helper()
	for i := range count {
		if strings.Contains(typeAt(i), suffix) {
			return
		}
	}
	t.Fatalf("expected plan modifier containing %q among %d modifiers", suffix, count)
}
