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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func testPinnedDashboardModelMinimalQuery() *models.DashboardQueryModel {
	return &models.DashboardQueryModel{
		Language: types.StringValue("kql"),
		Text:     types.StringValue(""),
		JSON:     jsontypes.NewNormalizedNull(),
	}
}

func newPinnedDashboardModelBase(pinned []models.PinnedPanelModel) *models.DashboardModel {
	return &models.DashboardModel{
		Title: types.StringValue("pinned-mapping-test"),
		RefreshInterval: &models.RefreshIntervalModel{
			Pause: types.BoolValue(true),
			Value: types.Int64Value(0),
		},
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-15m"),
			To:   types.StringValue("now"),
		},
		Query:        testPinnedDashboardModelMinimalQuery(),
		PinnedPanels: pinned,
	}
}

func pinnedFixtureOptionsList(field string) models.PinnedPanelModel {
	return models.PinnedPanelModel{
		Type: types.StringValue(panelTypeOptionsListControl),
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID: types.StringValue("dv"),
			FieldName:  types.StringValue(field),
		},
	}
}

func pinnedFixtureRangeSlider(minVal, maxVal string, step float32) models.PinnedPanelModel {
	return models.PinnedPanelModel{
		Type: types.StringValue(panelTypeRangeSlider),
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			DataViewID: types.StringValue("dv"),
			FieldName:  types.StringValue("source.bytes"),
			Value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(minVal),
				types.StringValue(maxVal),
			}),
			Step: types.Float32Value(step),
		},
	}
}

func mustAPIPinnedItems(t *testing.T, dm *models.DashboardModel) *kbapi.DashboardPinnedPanels {
	t.Helper()
	items, diags := dashboardPinnedPanelsToAPICreateItems(dm)
	require.False(t, diags.HasError(), "%s", diags)
	return items
}

func Test_dashboardModel_mapPinnedPanelsFromAPI_unsetVsEmptyAndDrift(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("prior nil + API nil yields nil", func(t *testing.T) {
		t.Parallel()
		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, nil, nil)
		require.False(t, diags.HasError())
		require.Nil(t, out)
	})

	t.Run("prior nil + API empty slice yields nil", func(t *testing.T) {
		t.Parallel()
		api := []kbapi.DashboardPinnedPanels_Item{}
		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, nil, &api)
		require.False(t, diags.HasError())
		require.Nil(t, out)
	})

	t.Run("prior empty slice + API empty slice yields empty slice", func(t *testing.T) {
		t.Parallel()
		api := []kbapi.DashboardPinnedPanels_Item{}
		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, []models.PinnedPanelModel{}, &api)
		require.False(t, diags.HasError())
		require.NotNil(t, out)
		require.Empty(t, out)
	})

	t.Run("prior nil + API one entry yields one populated entry", func(t *testing.T) {
		t.Parallel()
		src := newPinnedDashboardModelBase([]models.PinnedPanelModel{pinnedFixtureOptionsList("status")})
		api := mustAPIPinnedItems(t, src)
		require.NotNil(t, api)

		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, nil, api)
		require.False(t, diags.HasError())
		require.Len(t, out, 1)
		require.Equal(t, panelTypeOptionsListControl, out[0].Type.ValueString())
		require.Nil(t, out[0].RangeSliderControlConfig)
		require.NotNil(t, out[0].OptionsListControlConfig)
		require.Equal(t, "status", out[0].OptionsListControlConfig.FieldName.ValueString())
	})

	t.Run("prior populated + API populated same indices and types reuses prior pointers", func(t *testing.T) {
		t.Parallel()

		ol := &models.OptionsListControlConfigModel{
			DataViewID: types.StringValue("dv"),
			FieldName:  types.StringValue("status"),
		}
		rs := &models.RangeSliderControlConfigModel{
			DataViewID: types.StringValue("dv"),
			FieldName:  types.StringValue("source.bytes"),
			Value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("100"),
				types.StringValue("500"),
			}),
			Step: types.Float32Value(10),
		}

		prior := []models.PinnedPanelModel{
			{Type: types.StringValue(panelTypeOptionsListControl), OptionsListControlConfig: ol},
			{Type: types.StringValue(panelTypeRangeSlider), RangeSliderControlConfig: rs},
		}

		api := mustAPIPinnedItems(t, newPinnedDashboardModelBase([]models.PinnedPanelModel{prior[0], prior[1]}))
		require.NotNil(t, api)

		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, api)
		require.False(t, diags.HasError())
		require.Len(t, out, 2)
		require.Same(t, ol, out[0].OptionsListControlConfig)
		require.Same(t, rs, out[1].RangeSliderControlConfig)
	})

	t.Run("prior populated + API type drift clears mismatched typed configs", func(t *testing.T) {
		t.Parallel()

		priorOL := &models.OptionsListControlConfigModel{
			DataViewID: types.StringValue("dv"),
			FieldName:  types.StringValue("status"),
		}
		prior := []models.PinnedPanelModel{
			{Type: types.StringValue(panelTypeOptionsListControl), OptionsListControlConfig: priorOL},
		}

		apiModel := newPinnedDashboardModelBase([]models.PinnedPanelModel{pinnedFixtureRangeSlider("1", "2", 1)})
		api := mustAPIPinnedItems(t, apiModel)
		require.NotNil(t, api)

		out, diags := dashboardMapPinnedPanelsFromAPI(ctx, prior, api)
		require.False(t, diags.HasError())
		require.Len(t, out, 1)
		require.Equal(t, panelTypeRangeSlider, out[0].Type.ValueString())
		require.Nil(t, out[0].OptionsListControlConfig)
		require.NotNil(t, out[0].RangeSliderControlConfig)
	})
}

func Test_dashboardModel_toAPIRequests_pinnedPanelsJSONShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("create omits pinned_panels key when unset", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase(nil)
		req := dashboardToAPICreateRequest(ctx, m, diags)
		require.False(t, diags.HasError())

		raw, err := json.Marshal(req)
		require.NoError(t, err)

		var decoded map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &decoded))
		_, ok := decoded["pinned_panels"]
		require.False(t, ok)
	})

	t.Run("create includes explicit empty pinned_panels array", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase([]models.PinnedPanelModel{})
		req := dashboardToAPICreateRequest(ctx, m, diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.PinnedPanels)
		require.Empty(t, *req.PinnedPanels)
	})

	t.Run("update omits pinned_panels field assignment when unset", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase(nil)
		req := dashboardToAPIUpdateRequest(ctx, m, diags)
		require.False(t, diags.HasError())
		require.Nil(t, req.PinnedPanels)
	})

	t.Run("update assigns explicit empty pinned_panels array", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase([]models.PinnedPanelModel{})
		req := dashboardToAPIUpdateRequest(ctx, m, diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.PinnedPanels)
		require.Empty(t, *req.PinnedPanels)
	})

	t.Run("update assigns non-empty pinned_panels", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase([]models.PinnedPanelModel{
			pinnedFixtureOptionsList("status"),
		})
		req := dashboardToAPIUpdateRequest(ctx, m, diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.PinnedPanels)
		require.Len(t, *req.PinnedPanels, 1)

		raw, err := json.Marshal((*req.PinnedPanels)[0])
		require.NoError(t, err)
		require.NotContains(t, string(raw), `"grid"`)
	})

	t.Run("pinned panel payload JSON does not include grid", func(t *testing.T) {
		t.Parallel()
		diags := &diag.Diagnostics{}
		m := newPinnedDashboardModelBase([]models.PinnedPanelModel{
			pinnedFixtureOptionsList("status"),
		})
		req := dashboardToAPICreateRequest(ctx, m, diags)
		require.False(t, diags.HasError())
		require.NotNil(t, req.PinnedPanels)
		require.Len(t, *req.PinnedPanels, 1)

		raw, err := json.Marshal((*req.PinnedPanels)[0])
		require.NoError(t, err)
		require.NotContains(t, string(raw), `"grid"`)
	})
}
