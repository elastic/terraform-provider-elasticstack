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

package rule

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func validateArtifactData(t *testing.T, value types.String) validator.StringResponse {
	t.Helper()
	resp := validator.StringResponse{}
	artifactDataValidator().ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("artifacts").AtListIndex(0).AtName("data"),
		ConfigValue: value,
	}, &resp)
	return resp
}

func jsonObjectWithKeys(count int, keyPrefix string) string {
	keys := make([]string, 0, count)
	for i := range count {
		keys = append(keys, `"`+keyPrefix+strconv.Itoa(i)+`": 1`)
	}
	return "{" + strings.Join(keys, ",") + "}"
}

func TestArtifactDataValidator_rejectsNonObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"array":    `["a"]`,
		"string":   `"a"`,
		"not json": "not json",
		"null":     "null",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateArtifactData(t, types.StringValue(value))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Summary()+" "+resp.Diagnostics[0].Detail(), "JSON object")
		})
	}
}

func TestArtifactDataValidator_enforcesTopLevelKeyCount(t *testing.T) {
	t.Parallel()

	t.Run("accepts exactly 32 keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(jsonObjectWithKeys(32, "k")))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	t.Run("rejects 33 keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(jsonObjectWithKeys(33, "k")))
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics[0].Detail(), "at most 32 top-level keys")
	})
}

func TestArtifactDataValidator_enforcesKeyLength(t *testing.T) {
	t.Parallel()

	t.Run("accepts a 256-character key and unbounded nested keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(`{"`+strings.Repeat("k", 256)+`": 1, "a": {"`+strings.Repeat("n", 300)+`": true}}`))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	for name, value := range map[string]string{
		"empty key":    `{"": 1}`,
		"too long key": `{"` + strings.Repeat("k", 257) + `": 1}`,
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			resp := validateArtifactData(t, types.StringValue(value))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "1-256 characters")
		})
	}

	t.Run("unknown values are skipped", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringUnknown())
		require.False(t, resp.Diagnostics.HasError())
	})
}

func artifactsList(t *testing.T, ids ...string) types.List {
	t.Helper()
	objectType := getSchema(context.Background()).Attributes["artifacts"].GetType().(types.ListType).ElemType.(types.ObjectType)
	elements := make([]attr.Value, 0, len(ids))
	for _, id := range ids {
		elements = append(elements, types.ObjectValueMust(objectType.AttrTypes, map[string]attr.Value{
			"id":   types.StringValue(id),
			"type": types.StringValue("runbook"),
			"data": jsontypes.NewNormalizedValue(`{"content": "# Steps"}`),
		}))
	}
	return types.ListValueMust(objectType, elements)
}

func validateArtifactIDs(t *testing.T, value types.List) validator.ListResponse {
	t.Helper()
	resp := validator.ListResponse{}
	artifactIDsUniqueValidator().ValidateList(context.Background(), validator.ListRequest{
		Path:        path.Root("artifacts"),
		ConfigValue: value,
	}, &resp)
	return resp
}

func TestArtifactIDsUniqueValidator(t *testing.T) {
	t.Parallel()

	t.Run("accepts distinct ids", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactIDs(t, artifactsList(t, "rb", "db"))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	t.Run("rejects a repeated id at the repeated element", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactIDs(t, artifactsList(t, "rb", "db", "rb"))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, path.Root("artifacts").AtListIndex(2).AtName("id").String(), resp.Diagnostics[0].(diag.DiagnosticWithPath).Path().String())
		require.Contains(t, resp.Diagnostics[0].Detail(), `"rb"`)
	})

	t.Run("unknown list is skipped", func(t *testing.T) {
		t.Parallel()
		objectType := getSchema(context.Background()).Attributes["artifacts"].GetType().(types.ListType).ElemType
		resp := validateArtifactIDs(t, types.ListUnknown(objectType))
		require.False(t, resp.Diagnostics.HasError())
	})
}

func TestNonEmptyListValidator(t *testing.T) {
	t.Parallel()

	validate := func(t *testing.T, value types.List) validator.ListResponse {
		t.Helper()
		resp := validator.ListResponse{}
		nonEmptyListValidator().ValidateList(context.Background(), validator.ListRequest{
			Path:        path.Root("metadata").AtName("tags"),
			ConfigValue: value,
		}, &resp)
		return resp
	}

	t.Run("accepts one item", func(t *testing.T) {
		t.Parallel()
		resp := validate(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("prod")}))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	t.Run("rejects an empty list and says to omit the attribute", func(t *testing.T) {
		t.Parallel()
		resp := validate(t, types.ListValueMust(types.StringType, []attr.Value{}))
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics[0].Detail(), "at least one item")
		require.Contains(t, resp.Diagnostics[0].Detail(), "omit")
	})

	t.Run("null and unknown lists are skipped", func(t *testing.T) {
		t.Parallel()
		require.False(t, validate(t, types.ListNull(types.StringType)).Diagnostics.HasError())
		require.False(t, validate(t, types.ListUnknown(types.StringType)).Diagnostics.HasError())
	})
}

func TestNoDataStrategyValidator_rejectsAlert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	strategy := getSchema(ctx).Attributes["no_data"].(schema.SingleNestedAttribute).Attributes["strategy"].(schema.StringAttribute)
	resp := validator.StringResponse{}
	for _, v := range strategy.Validators {
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("no_data").AtName("strategy"),
			ConfigValue: types.StringValue(noDataStrategyAlert),
		}, &resp)
	}

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Detail(), `["ignore" "keep_last" "resolve"]`)
}
