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

package links

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func linksConfigAttrTypes() map[string]attr.Type {
	return models.LinksPanelConfigModel{}.AttrTypes()
}

func byValueAttrTypes() map[string]attr.Type {
	return models.LinksPanelByValueModel{}.AttrTypes()
}

func byReferenceAttrTypes() map[string]attr.Type {
	return models.LinksPanelByReferenceModel{}.AttrTypes()
}

func linkItemAttrTypes() map[string]attr.Type {
	return models.LinkItemModel{}.AttrTypes()
}

func newLinkItemObject(values map[string]attr.Value) types.Object {
	return types.ObjectValueMust(linkItemAttrTypes(), values)
}

func newByValueObject(layout string, items ...types.Object) types.Object {
	listElemType := types.ObjectType{AttrTypes: linkItemAttrTypes()}
	linksList := types.ListValueMust(listElemType, func() []attr.Value {
		vals := make([]attr.Value, len(items))
		for i, item := range items {
			vals[i] = item
		}
		return vals
	}())

	return types.ObjectValueMust(byValueAttrTypes(), map[string]attr.Value{
		"layout":      types.StringValue(layout),
		"title":       types.StringNull(),
		"description": types.StringNull(),
		"hide_title":  types.BoolNull(),
		"hide_border": types.BoolNull(),
		"links":       linksList,
	})
}

func newByReferenceObject(refID string) types.Object {
	return types.ObjectValueMust(byReferenceAttrTypes(), map[string]attr.Value{
		"ref_id":      types.StringValue(refID),
		"title":       types.StringNull(),
		"description": types.StringNull(),
		"hide_title":  types.BoolNull(),
		"hide_border": types.BoolNull(),
	})
}

func newLinksConfigObject(byValue, byReference attr.Value) types.Object {
	return types.ObjectValueMust(linksConfigAttrTypes(), map[string]attr.Value{
		"by_value":     byValue,
		"by_reference": byReference,
	})
}

func TestLinksConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := linksConfigModeValidator{}

	validDashboardLink := newLinkItemObject(map[string]attr.Value{
		"type":            types.StringValue("dashboard"),
		"destination":     types.StringValue("dashboard-1"),
		"label":           types.StringValue("Dashboard"),
		"open_in_new_tab": types.BoolValue(true),
		"use_filters":     types.BoolValue(true),
		"use_time_range":  types.BoolValue(false),
		"encode_url":      types.BoolNull(),
	})

	t.Run("both branches set -> error", func(t *testing.T) {
		t.Parallel()
		ov := newLinksConfigObject(
			newByValueObject("vertical", validDashboardLink),
			newByReferenceObject("ref-1"),
		)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("neither branch set -> error", func(t *testing.T) {
		t.Parallel()
		ov := newLinksConfigObject(
			types.ObjectNull(byValueAttrTypes()),
			types.ObjectNull(byReferenceAttrTypes()),
		)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Exactly one")
	})

	t.Run("only by_value -> ok", func(t *testing.T) {
		t.Parallel()
		ov := newLinksConfigObject(
			newByValueObject("vertical", validDashboardLink),
			types.ObjectNull(byReferenceAttrTypes()),
		)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("only by_reference -> ok", func(t *testing.T) {
		t.Parallel()
		ov := newLinksConfigObject(
			types.ObjectNull(byValueAttrTypes()),
			newByReferenceObject("ref-1"),
		)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})
}

func TestLinksItemTypeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := linksItemTypeValidator{}

	t.Run("dashboard link with encode_url set -> error", func(t *testing.T) {
		t.Parallel()
		ov := newLinkItemObject(map[string]attr.Value{
			"type":            types.StringValue("dashboard"),
			"destination":     types.StringValue("dashboard-1"),
			"label":           types.StringValue("Dashboard"),
			"open_in_new_tab": types.BoolValue(true),
			"use_filters":     types.BoolValue(true),
			"use_time_range":  types.BoolValue(false),
			"encode_url":      types.BoolValue(true),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links").AtListIndex(0)}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "encode_url")
	})

	t.Run("external link with use_filters set -> error", func(t *testing.T) {
		t.Parallel()
		ov := newLinkItemObject(map[string]attr.Value{
			"type":            types.StringValue("external"),
			"destination":     types.StringValue("https://example.com"),
			"label":           types.StringValue("Example"),
			"open_in_new_tab": types.BoolValue(false),
			"use_filters":     types.BoolValue(true),
			"use_time_range":  types.BoolNull(),
			"encode_url":      types.BoolValue(true),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links").AtListIndex(0)}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "use_filters")
	})

	t.Run("external link with use_time_range set -> error", func(t *testing.T) {
		t.Parallel()
		ov := newLinkItemObject(map[string]attr.Value{
			"type":            types.StringValue("external"),
			"destination":     types.StringValue("https://example.com"),
			"label":           types.StringValue("Example"),
			"open_in_new_tab": types.BoolValue(false),
			"use_filters":     types.BoolNull(),
			"use_time_range":  types.BoolValue(true),
			"encode_url":      types.BoolValue(true),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links").AtListIndex(0)}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "use_time_range")
	})

	t.Run("valid dashboard link -> ok", func(t *testing.T) {
		t.Parallel()
		ov := newLinkItemObject(map[string]attr.Value{
			"type":            types.StringValue("dashboard"),
			"destination":     types.StringValue("dashboard-1"),
			"label":           types.StringValue("Dashboard"),
			"open_in_new_tab": types.BoolValue(true),
			"use_filters":     types.BoolValue(true),
			"use_time_range":  types.BoolValue(false),
			"encode_url":      types.BoolNull(),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links").AtListIndex(0)}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("valid external link -> ok", func(t *testing.T) {
		t.Parallel()
		ov := newLinkItemObject(map[string]attr.Value{
			"type":            types.StringValue("external"),
			"destination":     types.StringValue("https://example.com"),
			"label":           types.StringValue("Example"),
			"open_in_new_tab": types.BoolValue(false),
			"use_filters":     types.BoolNull(),
			"use_time_range":  types.BoolNull(),
			"encode_url":      types.BoolValue(true),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("links").AtListIndex(0)}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})
}
