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

package mlanomalycharts_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/mlanomalycharts"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestSeverityThresholdItemValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	exactlyOne := validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{"severity", "min"},
		Summary:       "Invalid severity_threshold entry",
		MissingDetail: "Exactly one of `severity` or `min` must be set for each `severity_threshold` entry.",
		TooManyDetail: "Exactly one of `severity` or `min` must be set for each `severity_threshold` entry, not both.",
	})

	itemAttrs := map[string]attr.Type{
		"severity": types.StringType,
		"min":      types.Int64Type,
		"max":      types.Int64Type,
	}

	validateObject := func(t *testing.T, values map[string]attr.Value) {
		t.Helper()
		obj := types.ObjectValueMust(itemAttrs, values)
		resp := &validator.ObjectResponse{}
		req := validator.ObjectRequest{
			Path:        path.Root("severity_threshold").AtListIndex(0),
			ConfigValue: obj,
		}
		exactlyOne.ValidateObject(ctx, req, resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	}

	expectObjectError := func(t *testing.T, values map[string]attr.Value) {
		t.Helper()
		obj := types.ObjectValueMust(itemAttrs, values)
		resp := &validator.ObjectResponse{}
		req := validator.ObjectRequest{
			Path:        path.Root("severity_threshold").AtListIndex(0),
			ConfigValue: obj,
		}
		exactlyOne.ValidateObject(ctx, req, resp)
		require.True(t, resp.Diagnostics.HasError(), "expected validation error")
	}

	t.Run("accepts named severity", func(t *testing.T) {
		validateObject(t, map[string]attr.Value{
			"severity": types.StringValue("major"),
			"min":      types.Int64Null(),
			"max":      types.Int64Null(),
		})
	})

	t.Run("accepts raw range", func(t *testing.T) {
		validateObject(t, map[string]attr.Value{
			"severity": types.StringNull(),
			"min":      types.Int64Value(10),
			"max":      types.Int64Value(20),
		})
	})

	t.Run("rejects both severity and min", func(t *testing.T) {
		expectObjectError(t, map[string]attr.Value{
			"severity": types.StringValue("major"),
			"min":      types.Int64Value(50),
			"max":      types.Int64Null(),
		})
	})

	t.Run("rejects max without min or severity", func(t *testing.T) {
		expectObjectError(t, map[string]attr.Value{
			"severity": types.StringNull(),
			"min":      types.Int64Null(),
			"max":      types.Int64Value(75),
		})
	})

	t.Run("rejects severity with max", func(t *testing.T) {
		forbidMax := validators.ForbiddenIfDependentPathExpressionOneOf(
			path.MatchRelative().AtParent().AtName("severity"),
			[]string{"low", "warning", "minor", "major", "critical"},
		)

		itemSchema := schema.Schema{
			Attributes: map[string]schema.Attribute{
				"severity": schema.StringAttribute{Optional: true},
				"min":      schema.Int64Attribute{Optional: true},
				"max":      schema.Int64Attribute{Optional: true},
			},
		}
		rawConfig := tftypes.NewValue(
			tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"severity": tftypes.String,
				"min":      tftypes.Number,
				"max":      tftypes.Number,
			}},
			map[string]tftypes.Value{
				"severity": tftypes.NewValue(tftypes.String, "major"),
				"min":      tftypes.NewValue(tftypes.Number, nil),
				"max":      tftypes.NewValue(tftypes.Number, float64(75)),
			},
		)
		config := tfsdk.Config{Raw: rawConfig, Schema: itemSchema}

		resp := &validator.Int64Response{}
		forbidMax.ValidateInt64(ctx, validator.Int64Request{
			Path:        path.Root("max"),
			ConfigValue: types.Int64Value(75),
			Config:      config,
		}, resp)
		require.True(t, resp.Diagnostics.HasError(), "expected validation error when severity and max are both set")
	})

	t.Run("rejects neither severity nor min", func(t *testing.T) {
		expectObjectError(t, map[string]attr.Value{
			"severity": types.StringNull(),
			"min":      types.Int64Null(),
			"max":      types.Int64Null(),
		})
	})
}

func TestSchemaAttribute_registersPanelType(t *testing.T) {
	t.Parallel()
	attr := mlanomalycharts.SchemaAttribute()
	require.NotNil(t, attr)
}
