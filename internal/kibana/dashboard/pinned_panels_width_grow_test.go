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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func pinnedFixtureTimeSlider() models.PinnedPanelModel {
	return models.PinnedPanelModel{
		Type: types.StringValue(panelTypeTimeSlider),
		TimeSliderControlConfig: &models.TimeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.1),
			EndPercentageOfTimeRange:   types.Float32Value(0.9),
			IsAnchored:                 types.BoolValue(false),
		},
	}
}

func pinnedFixtureEsql() models.PinnedPanelModel {
	return models.PinnedPanelModel{
		Type: types.StringValue(panelTypeEsqlControl),
		EsqlControlConfig: &models.EsqlControlConfigModel{
			SelectedOptions:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			VariableName:     types.StringValue("field"),
			VariableType:     types.StringValue("fields"),
			EsqlQuery:        types.StringValue("FROM logs | KEEP @timestamp"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			AvailableOptions: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
		},
	}
}

func withSizing(ppm models.PinnedPanelModel, width types.String, grow types.Bool) models.PinnedPanelModel {
	ppm.Width = width
	ppm.Grow = grow
	return ppm
}

type pinnedSizingCase struct {
	name    string
	fixture models.PinnedPanelModel
}

func pinnedSizingCases() []pinnedSizingCase {
	return []pinnedSizingCase{
		{panelTypeOptionsListControl, pinnedFixtureOptionsList("status")},
		{panelTypeRangeSlider, pinnedFixtureRangeSlider("1", "2", 1)},
		{panelTypeEsqlControl, pinnedFixtureEsql()},
		{panelTypeTimeSlider, pinnedFixtureTimeSlider()},
	}
}

func pinnedPayloadSizing(t *testing.T, ppm models.PinnedPanelModel) map[string]any {
	t.Helper()
	items := mustAPIPinnedItems(t, newPinnedDashboardModelBase([]models.PinnedPanelModel{ppm}))
	require.Len(t, *items, 1)
	raw, err := json.Marshal((*items)[0])
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return decoded
}

func Test_pinnedPanels_widthGrow_toAPI(t *testing.T) {
	t.Parallel()
	for _, tc := range pinnedSizingCases() {
		t.Run(tc.name+" includes configured width and grow", func(t *testing.T) {
			t.Parallel()
			payload := pinnedPayloadSizing(t, withSizing(tc.fixture, types.StringValue("small"), types.BoolValue(true)))
			require.Equal(t, "small", payload["width"])
			require.Equal(t, true, payload["grow"])
		})
		t.Run(tc.name+" includes explicit false grow", func(t *testing.T) {
			t.Parallel()
			payload := pinnedPayloadSizing(t, withSizing(tc.fixture, types.StringValue("large"), types.BoolValue(false)))
			require.Equal(t, "large", payload["width"])
			require.Equal(t, false, payload["grow"])
		})
		t.Run(tc.name+" omits null width and grow", func(t *testing.T) {
			t.Parallel()
			payload := pinnedPayloadSizing(t, withSizing(tc.fixture, types.StringNull(), types.BoolNull()))
			require.NotContains(t, payload, "width")
			require.NotContains(t, payload, "grow")
		})
	}
}

func Test_pinnedPanels_widthGrow_fromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, tc := range pinnedSizingCases() {
		apiWith := func(t *testing.T, width types.String, grow types.Bool) *kbapi.DashboardPinnedPanels {
			t.Helper()
			items := mustAPIPinnedItems(t, newPinnedDashboardModelBase([]models.PinnedPanelModel{withSizing(tc.fixture, width, grow)}))
			out := *items
			return &out
		}

		t.Run(tc.name+" no prior reads API values directly", func(t *testing.T) {
			t.Parallel()
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, nil, apiWith(t, types.StringValue("small"), types.BoolValue(false)))
			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, types.StringValue("small"), out[0].Width)
			require.Equal(t, types.BoolValue(false), out[0].Grow)
		})
		t.Run(tc.name+" no prior with absent API values stays null", func(t *testing.T) {
			t.Parallel()
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, nil, apiWith(t, types.StringNull(), types.BoolNull()))
			require.False(t, diags.HasError(), "%s", diags)
			require.True(t, out[0].Width.IsNull())
			require.True(t, out[0].Grow.IsNull())
		})
		t.Run(tc.name+" prior null preserves null against API defaults", func(t *testing.T) {
			t.Parallel()
			prior := []models.PinnedPanelModel{withSizing(tc.fixture, types.StringNull(), types.BoolNull())}
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, apiWith(t, types.StringValue("medium"), types.BoolValue(false)))
			require.False(t, diags.HasError(), "%s", diags)
			require.True(t, out[0].Width.IsNull())
			require.True(t, out[0].Grow.IsNull())
		})
		t.Run(tc.name+" configured prior round-trips", func(t *testing.T) {
			t.Parallel()
			prior := []models.PinnedPanelModel{withSizing(tc.fixture, types.StringValue("large"), types.BoolValue(true))}
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, apiWith(t, types.StringValue("large"), types.BoolValue(true)))
			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, types.StringValue("large"), out[0].Width)
			require.Equal(t, types.BoolValue(true), out[0].Grow)
		})
		t.Run(tc.name+" configured prior detects drift", func(t *testing.T) {
			t.Parallel()
			prior := []models.PinnedPanelModel{withSizing(tc.fixture, types.StringValue("large"), types.BoolValue(true))}
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, apiWith(t, types.StringValue("small"), types.BoolValue(false)))
			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, types.StringValue("small"), out[0].Width)
			require.Equal(t, types.BoolValue(false), out[0].Grow)
		})
		t.Run(tc.name+" different-type prior reads API values directly", func(t *testing.T) {
			t.Parallel()
			other := pinnedFixtureOptionsList("status")
			if tc.name == panelTypeOptionsListControl {
				other = pinnedFixtureTimeSlider()
			}
			prior := []models.PinnedPanelModel{withSizing(other, types.StringNull(), types.BoolNull())}
			out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, apiWith(t, types.StringValue("medium"), types.BoolValue(true)))
			require.False(t, diags.HasError(), "%s", diags)
			require.Equal(t, types.StringValue("medium"), out[0].Width)
			require.Equal(t, types.BoolValue(true), out[0].Grow)
		})
	}
}
