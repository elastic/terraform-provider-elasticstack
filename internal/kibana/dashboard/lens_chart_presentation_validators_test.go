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

package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_lensChartDrilldown_urlTrigger_oneOf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	urlBlock := lensChartDrilldownListItemAttributes()["url_drilldown"].(schema.SingleNestedAttribute)
	trigger, ok := urlBlock.Attributes["trigger"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, trigger.Validators)

	req := validator.StringRequest{
		Path:        path.Root("trigger"),
		ConfigValue: types.StringValue("totally_invalid"),
	}

	var resp validator.StringResponse
	for _, v := range trigger.Validators {
		v.ValidateString(ctx, req, &resp)
	}
	require.True(t, resp.Diagnostics.HasError(), "invalid url_drilldown.trigger should be rejected")
}

func Test_lensChartDrilldown_dashboardTrigger_isComputedSchema(t *testing.T) {
	t.Parallel()

	dashBlock := lensChartDrilldownListItemAttributes()["dashboard_drilldown"].(schema.SingleNestedAttribute)
	trigger, ok := dashBlock.Attributes["trigger"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, trigger.Computed, "dashboard_drilldown.trigger is computed from Kibana and must not be configurable")
}

func Test_drilldownListItemVariantsValidator_zeroVariants(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	attrs := lensChartDrilldownListItemAttributes()
	itemTypes := map[string]attr.Type{
		"dashboard_drilldown": attrs["dashboard_drilldown"].(schema.SingleNestedAttribute).GetType(),
		"discover_drilldown":  attrs["discover_drilldown"].(schema.SingleNestedAttribute).GetType(),
		"url_drilldown":       attrs["url_drilldown"].(schema.SingleNestedAttribute).GetType(),
	}

	nullAll := map[string]attr.Value{
		"dashboard_drilldown": types.ObjectNull(itemTypes["dashboard_drilldown"].(types.ObjectType).AttrTypes),
		"discover_drilldown":  types.ObjectNull(itemTypes["discover_drilldown"].(types.ObjectType).AttrTypes),
		"url_drilldown":       types.ObjectNull(itemTypes["url_drilldown"].(types.ObjectType).AttrTypes),
	}
	ov := types.ObjectValueMust(itemTypes, nullAll)

	var resp validator.ObjectResponse
	(lenscommon.DrilldownListItemVariantsValidator{}).ValidateObject(ctx, validator.ObjectRequest{
		Path:        path.Root("drilldowns").AtListIndex(0),
		ConfigValue: ov,
	}, &resp)

	require.True(t, resp.Diagnostics.HasError(), "empty drilldown item (no variants) must be rejected")
}
