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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	assertHasPlanModifier(t, idAttr.PlanModifiers, "useStateForUnknown")

	savedQueryIDAttr, ok := s.Attributes["saved_query_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, savedQueryIDAttr.IsRequired())
	assert.False(t, savedQueryIDAttr.IsComputed())
	assertHasPlanModifier(t, savedQueryIDAttr.PlanModifiers, "requiresReplace")

	spaceIDAttr, ok := s.Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, spaceIDAttr.IsOptional())
	assert.True(t, spaceIDAttr.IsComputed())
	assertHasPlanModifier(t, spaceIDAttr.PlanModifiers, "useStateForUnknown")
	assertHasPlanModifier(t, spaceIDAttr.PlanModifiers, "requiresReplace")

	queryAttr, ok := s.Attributes["query"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, queryAttr.IsRequired())

	snapshotAttr, ok := s.Attributes["snapshot"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, snapshotAttr.IsOptional())
	assert.True(t, snapshotAttr.IsComputed())

	removedAttr, ok := s.Attributes["removed"].(schema.BoolAttribute)
	require.True(t, ok)
	assert.True(t, removedAttr.IsOptional())
	assert.True(t, removedAttr.IsComputed())
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
			req := validator.SetRequest{
				ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue(value)}),
			}
			var resp validator.SetResponse
			for _, val := range platformAttr.Validators {
				val.ValidateSet(context.Background(), req, &resp)
			}
			assert.False(t, resp.Diagnostics.HasError(), "expected %q to be valid: %v", value, resp.Diagnostics)
		})
	}

	t.Run("invalid/ios", func(t *testing.T) {
		t.Parallel()
		req := validator.SetRequest{
			ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ios")}),
		}
		var resp validator.SetResponse
		for _, val := range platformAttr.Validators {
			val.ValidateSet(context.Background(), req, &resp)
		}
		assert.True(t, resp.Diagnostics.HasError())
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

func Test_ecsMappingExactlyOneOfValidator(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ecsMappingExactlyOneOfValidator()

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
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: fieldOnly(), Path: path.Root("ecs_mapping").AtMapKey("process.name")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts value only", func(t *testing.T) {
		t.Parallel()
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: valueOnly(), Path: path.Root("ecs_mapping").AtMapKey("event.category")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts values only", func(t *testing.T) {
		t.Parallel()
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: valuesOnly(), Path: path.Root("ecs_mapping").AtMapKey("event.category")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects field and value", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringValue("cmdline"),
			attrEcsMappingValue:  types.StringValue("literal"),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: obj, Path: path.Root("ecs_mapping").AtMapKey("k")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not more than one")
	})

	t.Run("rejects empty element", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  types.StringNull(),
			attrEcsMappingValue:  types.StringNull(),
			attrEcsMappingValues: types.SetNull(types.StringType),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: obj, Path: path.Root("ecs_mapping").AtMapKey("k")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Exactly one of")
	})
}

func schemaNestedObjectElemType(no schema.NestedAttributeObject) attr.Type {
	attrTypes := make(map[string]attr.Type, len(no.Attributes))
	for name, a := range no.Attributes {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}

func assertHasPlanModifier(t *testing.T, modifiers []planmodifier.String, suffix string) {
	t.Helper()
	for _, m := range modifiers {
		if strings.Contains(reflect.TypeOf(m).String(), suffix) {
			return
		}
	}
	t.Fatalf("expected plan modifier containing %q among %d modifiers", suffix, len(modifiers))
}
