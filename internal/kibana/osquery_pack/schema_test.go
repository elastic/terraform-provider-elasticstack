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

package osquerypack

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

	packIDAttr, ok := s.Attributes["pack_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, packIDAttr.IsComputed())
	assert.False(t, packIDAttr.IsRequired())
	assert.False(t, packIDAttr.IsOptional())
	assertEmptyStringPlanModifiers(t, packIDAttr.PlanModifiers)

	spaceIDAttr, ok := s.Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.True(t, spaceIDAttr.IsComputed())
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "useStateForUnknown")
	assertHasStringPlanModifier(t, spaceIDAttr.PlanModifiers, "requiresReplace")

	nameAttr, ok := s.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, nameAttr.IsRequired())

	descriptionAttr, ok := s.Attributes["description"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, descriptionAttr.IsOptional())
	assert.False(t, descriptionAttr.IsRequired())
	assert.False(t, descriptionAttr.IsComputed())

	enabledAttr, ok := s.Attributes["enabled"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, enabledAttr.IsOptional())
	assert.False(t, enabledAttr.IsRequired())
	assert.False(t, enabledAttr.IsComputed())

	policyIDsAttr, ok := s.Attributes["policy_ids"].(schema.ListAttribute)
	require.True(t, ok)
	assert.True(t, policyIDsAttr.IsOptional())
	assert.Equal(t, types.StringType, policyIDsAttr.ElementType)

	shardsAttr, ok := s.Attributes["shards"].(schema.MapAttribute)
	require.True(t, ok)
	assert.True(t, shardsAttr.IsOptional())
	assert.Equal(t, types.Float64Type, shardsAttr.ElementType)
	require.NotEmpty(t, shardsAttr.Validators)

	queriesAttr, ok := s.Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, queriesAttr.IsRequired())
	require.NotEmpty(t, queriesAttr.Validators)

	queryAttr, ok := queriesAttr.NestedObject.Attributes["query"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, queryAttr.IsRequired())

	platformAttr, ok := queriesAttr.NestedObject.Attributes["platform"].(schema.SetAttribute)
	require.True(t, ok)
	assert.True(t, platformAttr.IsOptional())
	assert.Equal(t, types.StringType, platformAttr.ElementType)
	require.NotEmpty(t, platformAttr.Validators)

	versionAttr, ok := queriesAttr.NestedObject.Attributes["version"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, versionAttr.IsOptional())

	savedQueryIDAttr, ok := queriesAttr.NestedObject.Attributes["saved_query_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, savedQueryIDAttr.IsOptional())

	ecsMappingAttr, ok := queriesAttr.NestedObject.Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok)
	assert.True(t, ecsMappingAttr.IsOptional())

	snapshotAttr, ok := queriesAttr.NestedObject.Attributes["snapshot"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, snapshotAttr.IsOptional())
	assert.True(t, snapshotAttr.IsComputed())
	assertHasBoolPlanModifier(t, snapshotAttr.BoolPlanModifiers(), "useStateForUnknown")

	removedAttr, ok := queriesAttr.NestedObject.Attributes["removed"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, removedAttr.IsOptional())
	assert.True(t, removedAttr.IsComputed())
	assertHasBoolPlanModifier(t, removedAttr.BoolPlanModifiers(), "useStateForUnknown")
}

func TestSchema_queriesMapValidators(t *testing.T) {
	t.Parallel()

	queriesAttr, ok := getSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.NotEmpty(t, queriesAttr.Validators)

	t.Run("rejects empty map", func(t *testing.T) {
		t.Parallel()
		diags := validateMapValidators(
			context.Background(),
			queriesAttr.Validators,
			types.MapValueMust(queryMapElemType(), map[string]attr.Value{}),
			path.Root("queries"),
		)
		require.True(t, diags.HasError())
	})
}

func TestSchema_shardsMapValidators(t *testing.T) {
	t.Parallel()

	shardsAttr, ok := getSchema(context.Background()).Attributes["shards"].(schema.MapAttribute)
	require.True(t, ok)
	require.NotEmpty(t, shardsAttr.Validators)

	t.Run("accepts value in range", func(t *testing.T) {
		t.Parallel()
		diags := validateMapValidators(
			context.Background(),
			shardsAttr.Validators,
			types.MapValueMust(types.Float64Type, map[string]attr.Value{
				"policy-abc": types.Float64Value(75),
			}),
			path.Root("shards"),
		)
		require.False(t, diags.HasError(), "%s", diags)
	})

	t.Run("rejects value below range", func(t *testing.T) {
		t.Parallel()
		diags := validateMapValidators(
			context.Background(),
			shardsAttr.Validators,
			types.MapValueMust(types.Float64Type, map[string]attr.Value{
				"policy-abc": types.Float64Value(0),
			}),
			path.Root("shards"),
		)
		require.True(t, diags.HasError())
	})

	t.Run("rejects value above range", func(t *testing.T) {
		t.Parallel()
		diags := validateMapValidators(
			context.Background(),
			shardsAttr.Validators,
			types.MapValueMust(types.Float64Type, map[string]attr.Value{
				"policy-abc": types.Float64Value(101),
			}),
			path.Root("shards"),
		)
		require.True(t, diags.HasError())
	})
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

	queriesAttr, ok := getSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)

	platformAttr, ok := queriesAttr.NestedObject.Attributes["platform"].(schema.SetAttribute)
	require.True(t, ok)
	require.NotEmpty(t, platformAttr.Validators)

	for _, value := range osqueryPlatformValues {
		t.Run("valid/"+value, func(t *testing.T) {
			t.Parallel()
			diags := validateSetValidators(context.Background(), platformAttr.Validators, types.SetValueMust(types.StringType, []attr.Value{types.StringValue(value)}), path.Root("queries").AtMapKey("find_procs").AtName("platform"))
			assert.False(t, diags.HasError(), "expected %q to be valid: %v", value, diags)
		})
	}

	t.Run("invalid/ios", func(t *testing.T) {
		t.Parallel()
		diags := validateSetValidators(context.Background(), platformAttr.Validators, types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ios")}), path.Root("queries").AtMapKey("find_procs").AtName("platform"))
		assert.True(t, diags.HasError())
	})
}

func TestQueryElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	queriesAttr, ok := getSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok, "expected queries to be MapNestedAttribute")

	schemaElem := schemaNestedObjectElemType(queriesAttr.NestedObject)
	require.Equal(t, schemaElem, queryMapElemType(),
		"queryMapElemType() drifted from the schema's queries nested object; update both together")
}

func TestEcsMappingElemType_matchesSchema(t *testing.T) {
	t.Parallel()

	queriesAttr, ok := getSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)

	ecsMappingAttr, ok := queriesAttr.NestedObject.Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok, "expected queries.ecs_mapping to be MapNestedAttribute")

	schemaElem := schemaNestedObjectElemType(ecsMappingAttr.NestedObject)
	require.Equal(t, schemaElem, ecsMappingMapElemType(),
		"ecsMappingMapElemType() drifted from schema ecs_mapping nested object; update both together")
}

func TestSchema_ecsMappingNestedObjectValidatorsWired(t *testing.T) {
	t.Parallel()

	fromQueries := requireQueriesNestedObjectFromSchema(t)
	ecsMappingAttr, ok := fromQueries.Attributes["ecs_mapping"].(schema.MapNestedAttribute)
	require.True(t, ok)
	require.NotEmpty(t, ecsMappingAttr.NestedObject.Validators, "ecs_mapping from getSchema must attach nested object validators")

	fromHelper := ecsMappingSchema()
	require.NotEmpty(t, fromHelper.NestedObject.Validators, "ecsMappingSchema must attach nested object validators")
}

func TestSchema_ecsMappingNestedObjectValidators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nestedObject := requireEcsMappingNestedObjectFromSchema(t)
	elemPath := path.Root("queries").AtMapKey("find_procs").AtName("ecs_mapping").AtMapKey("process.name")

	fieldOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
			attrEcsMappingField:  types.StringValue("cmdline"),
			attrEcsMappingValue:  types.StringNull(),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
	}
	valueOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringValue("process"),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
	}
	valuesOnly := func() types.Object {
		return types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringNull(),
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
		obj := types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
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
		obj := types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
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
		obj := types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
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

	t.Run("rejects value and values", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringValue("literal"),
			attrEcsMappingValues: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("process"),
			}),
		})
		diags := validateObjectValidators(ctx, nestedObject.Validators, obj, elemPath)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "not more than one")
	})

	t.Run("rejects all three", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes(), map[string]attr.Value{
			attrEcsMappingField: types.StringValue("cmdline"),
			attrEcsMappingValue: types.StringValue("literal"),
			attrEcsMappingValues: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("process"),
			}),
		})
		diags := validateObjectValidators(ctx, nestedObject.Validators, obj, elemPath)
		require.True(t, diags.HasError())
		require.Contains(t, diags.Errors()[0].Detail(), "not more than one")
	})
}

func TestSchema_ecsMappingValuesSetValidators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nestedObject := requireEcsMappingNestedObjectFromSchema(t)
	valuesAttr, ok := nestedObject.Attributes[attrEcsMappingValues].(schema.SetAttribute)
	require.True(t, ok)
	require.NotEmpty(t, valuesAttr.Validators)

	t.Run("rejects empty values set", func(t *testing.T) {
		t.Parallel()
		diags := validateSetValidators(
			ctx,
			valuesAttr.Validators,
			types.SetValueMust(types.StringType, []attr.Value{}),
			path.Root("queries").AtMapKey("find_procs").AtName("ecs_mapping").AtMapKey("process.name").AtName(attrEcsMappingValues),
		)
		require.True(t, diags.HasError())
	})
}

func requireQueriesNestedObjectFromSchema(t *testing.T) schema.NestedAttributeObject {
	t.Helper()

	queriesAttr, ok := getSchema(context.Background()).Attributes["queries"].(schema.MapNestedAttribute)
	require.True(t, ok)
	return queriesAttr.NestedObject
}

func requireEcsMappingNestedObjectFromSchema(t *testing.T) schema.NestedAttributeObject {
	t.Helper()

	queriesAttr := requireQueriesNestedObjectFromSchema(t)
	ecsMappingAttr, ok := queriesAttr.Attributes["ecs_mapping"].(schema.MapNestedAttribute)
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

func validateMapValidators(ctx context.Context, validators []validator.Map, value types.Map, p path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	req := validator.MapRequest{ConfigValue: value, Path: p}
	for _, v := range validators {
		var resp validator.MapResponse
		v.ValidateMap(ctx, req, &resp)
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

func assertEmptyStringPlanModifiers(t *testing.T, modifiers []planmodifier.String) {
	t.Helper()
	require.Empty(t, modifiers)
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
