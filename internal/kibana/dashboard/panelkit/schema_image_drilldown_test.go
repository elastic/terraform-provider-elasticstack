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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_imageDrilldownEntryValidator(t *testing.T) {
	ctx := context.Background()
	v := ExactlyOneOfNestedAttrsValidator(ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{"dashboard_drilldown", "url_drilldown"},
		Summary:       "Invalid drilldown entry",
		MissingDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set.",
		TooManyDetail: "Exactly one of `dashboard_drilldown` or `url_drilldown` must be set, not both.",
	})

	dashType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"dashboard_id": types.StringType,
		"label":        types.StringType,
		"trigger":      types.StringType,
	}}
	urlType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"url": types.StringType, "label": types.StringType, "trigger": types.StringType,
	}}
	entryType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"dashboard_drilldown": dashType,
		"url_drilldown":       urlType,
	}}

	dashObj := types.ObjectValueMust(dashType.AttrTypes, map[string]attr.Value{
		"dashboard_id": types.StringValue("d"),
		"label":        types.StringValue("l"),
		"trigger":      types.StringValue("on_click_image"),
	})
	urlObj := types.ObjectValueMust(urlType.AttrTypes, map[string]attr.Value{
		"url":     types.StringValue("https://x"),
		"label":   types.StringValue("u"),
		"trigger": types.StringValue("on_open_panel_menu"),
	})

	t.Run("rejects both drilldown kinds", func(t *testing.T) {
		ov := types.ObjectValueMust(entryType.AttrTypes, map[string]attr.Value{
			"dashboard_drilldown": dashObj,
			"url_drilldown":       urlObj,
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("drilldowns")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("rejects neither", func(t *testing.T) {
		ov := types.ObjectValueMust(entryType.AttrTypes, map[string]attr.Value{
			"dashboard_drilldown": types.ObjectNull(dashType.AttrTypes),
			"url_drilldown":       types.ObjectNull(urlType.AttrTypes),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("drilldowns")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}
