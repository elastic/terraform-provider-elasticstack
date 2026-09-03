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

package fieldstatstable_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/fieldstatstable"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func panelModelBase() models.PanelModel {
	return models.PanelModel{
		Grid: models.PanelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(24),
			H: types.Int64Value(15),
		},
		ID: types.StringValue("panel-id"),
	}
}

func stringVal(s string) types.String { return types.StringValue(s) }
func boolVal(b bool) types.Bool       { return types.BoolValue(b) }
func stringNull() types.String        { return types.StringNull() }
func boolNull() types.Bool            { return types.BoolNull() }

func configJSONSet(s string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(s, func(m map[string]any) map[string]any { return m })
}

// deepCopyFieldStatsTableConfig clones a FieldStatsTableConfigModel so that a test's "prior" and
// "pm" models don't alias the same nested pointers. FromAPI mutates the target model's nested
// structs in place; if prior and pm share those pointers, null-preservation logic ends up
// comparing a mutated value against itself instead of the true prior state.
func deepCopyFieldStatsTableConfig(cfg *models.FieldStatsTableConfigModel) *models.FieldStatsTableConfigModel {
	if cfg == nil {
		return nil
	}
	out := &models.FieldStatsTableConfigModel{}
	if cfg.ByDataview != nil {
		dv := *cfg.ByDataview
		dv.TimeRange = deepCopyTimeRange(cfg.ByDataview.TimeRange)
		out.ByDataview = &dv
	}
	if cfg.ByEsql != nil {
		esql := *cfg.ByEsql
		esql.TimeRange = deepCopyTimeRange(cfg.ByEsql.TimeRange)
		out.ByEsql = &esql
	}
	return out
}

func deepCopyTimeRange(tr *models.TimeRangeModel) *models.TimeRangeModel {
	if tr == nil {
		return nil
	}
	out := *tr
	return &out
}

func configMap(t *testing.T, item kbapi.DashboardPanelItem) map[string]any {
	t.Helper()
	raw, err := json.Marshal(item)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	cfg, ok := m["config"].(map[string]any)
	require.True(t, ok, "config should be object")
	return cfg
}

func diagSummary(diags diag.Diagnostics) string {
	if diags == nil {
		return ""
	}
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Summary())
		if dt := d.Detail(); dt != "" {
			b.WriteString(" — ")
			b.WriteString(dt)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func TestHandler_roundTrip_byDataview(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pm := panelModelBase()
	pm.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID:        stringVal("logs-view"),
			ShowDistributions: boolVal(true),
			Title:             stringVal("Field statistics — logs view"),
			Description:       stringVal("Field stats table panel (dataview)"),
			HideTitle:         boolVal(false),
			HideBorder:        boolVal(true),
			TimeRange: &models.TimeRangeModel{
				From: stringVal("now-24h"),
				To:   stringVal("now"),
				Mode: stringVal("relative"),
			},
		},
	}

	item, diags := fieldstatstable.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := configMap(t, item)
	require.Equal(t, "dataview", cfg["view_type"])
	require.Equal(t, "logs-view", cfg["data_view_id"])
	require.Equal(t, true, cfg["show_distributions"])
	require.Equal(t, "Field stats table panel (dataview)", cfg["description"])
	require.Equal(t, false, cfg["hide_title"])
	require.Equal(t, true, cfg["hide_border"])

	next := pm
	next.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(pm.FieldStatsTableConfig)
	d2 := fieldstatstable.Handler{}.FromAPI(ctx, &next, &pm, item)
	require.False(t, d2.HasError(), "%s", d2)

	require.Nil(t, next.FieldStatsTableConfig.ByEsql)
	require.NotNil(t, next.FieldStatsTableConfig.ByDataview)
	dv := next.FieldStatsTableConfig.ByDataview
	assert.Equal(t, "logs-view", dv.DataViewID.ValueString())
	assert.True(t, dv.ShowDistributions.ValueBool())
	assert.Equal(t, "Field statistics — logs view", dv.Title.ValueString())
	assert.Equal(t, "Field stats table panel (dataview)", dv.Description.ValueString())
	assert.False(t, dv.HideTitle.ValueBool())
	assert.True(t, dv.HideBorder.ValueBool())
	require.NotNil(t, dv.TimeRange)
	assert.Equal(t, "now-24h", dv.TimeRange.From.ValueString())
	assert.Equal(t, "now", dv.TimeRange.To.ValueString())
	assert.Equal(t, "relative", dv.TimeRange.Mode.ValueString())
}

func TestHandler_roundTrip_byEsql(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pm := panelModelBase()
	pm.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByEsql: &models.FieldStatsTableByEsqlModel{
			Query:             stringVal("FROM logs | STATS count = COUNT(*) BY service.name"),
			ShowDistributions: boolVal(false),
			Title:             stringVal("Field statistics — logs by service"),
			Description:       stringVal("Field stats table panel (esql)"),
			HideTitle:         boolVal(true),
			HideBorder:        boolVal(false),
			TimeRange: &models.TimeRangeModel{
				From: stringVal("now-24h"),
				To:   stringVal("now"),
				Mode: stringVal("relative"),
			},
		},
	}

	item, diags := fieldstatstable.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := configMap(t, item)
	require.Equal(t, "esql", cfg["view_type"])
	query, ok := cfg["query"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "FROM logs | STATS count = COUNT(*) BY service.name", query["esql"])
	assert.Equal(t, false, cfg["show_distributions"])
	require.Equal(t, "Field stats table panel (esql)", cfg["description"])
	require.Equal(t, true, cfg["hide_title"])
	require.Equal(t, false, cfg["hide_border"])

	next := pm
	next.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(pm.FieldStatsTableConfig)
	d2 := fieldstatstable.Handler{}.FromAPI(ctx, &next, &pm, item)
	require.False(t, d2.HasError(), "%s", d2)

	require.Nil(t, next.FieldStatsTableConfig.ByDataview)
	require.NotNil(t, next.FieldStatsTableConfig.ByEsql)
	esql := next.FieldStatsTableConfig.ByEsql
	assert.Equal(t, "FROM logs | STATS count = COUNT(*) BY service.name", esql.Query.ValueString())
	assert.False(t, esql.ShowDistributions.ValueBool())
	assert.Equal(t, "Field stats table panel (esql)", esql.Description.ValueString())
	assert.True(t, esql.HideTitle.ValueBool())
	assert.False(t, esql.HideBorder.ValueBool())
	require.NotNil(t, esql.TimeRange)
	assert.Equal(t, "relative", esql.TimeRange.Mode.ValueString())
}

func TestHandler_toAPI_omitsNullOptionalFields(t *testing.T) {
	t.Parallel()

	pm := panelModelBase()
	pm.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID:        stringVal("logs-view"),
			ShowDistributions: boolNull(),
			Title:             stringNull(),
			Description:       stringNull(),
			HideTitle:         boolNull(),
			HideBorder:        boolNull(),
		},
	}

	item, diags := fieldstatstable.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	cfg := configMap(t, item)
	require.Equal(t, "logs-view", cfg["data_view_id"])
	_, hasShow := cfg["show_distributions"]
	assert.False(t, hasShow)
	_, hasTitle := cfg["title"]
	assert.False(t, hasTitle)
	_, hasDescription := cfg["description"]
	assert.False(t, hasDescription)
	_, hasHideTitle := cfg["hide_title"]
	assert.False(t, hasHideTitle)
	_, hasHideBorder := cfg["hide_border"]
	assert.False(t, hasHideBorder)
	_, hasTR := cfg["time_range"]
	assert.False(t, hasTR)
}

func TestHandler_fromAPI_nullPreservation_timeRange(t *testing.T) {
	t.Parallel()

	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0{
		DataViewId: "logs-view",
		ViewType:   kbapi.Dataview,
	}
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: "now-24h",
		To:   "now",
		Mode: &mode,
	}
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats0(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID: stringVal("logs-view"),
			TimeRange:  nil,
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.Nil(t, pm.FieldStatsTableConfig.ByDataview.TimeRange)
}

func TestHandler_fromAPI_branchMismatch_priorDataviewAPIEsql(t *testing.T) {
	t.Parallel()

	show := true
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1{
		ViewType:          kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1ViewTypeEsql,
		ShowDistributions: &show,
	}
	apiCfg.Query.Esql = "FROM logs"
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats1(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID: stringVal("logs-view"),
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.Nil(t, pm.FieldStatsTableConfig.ByDataview)
	require.NotNil(t, pm.FieldStatsTableConfig.ByEsql)
	assert.Equal(t, "FROM logs", pm.FieldStatsTableConfig.ByEsql.Query.ValueString())
}

func TestHandler_fromAPI_branchMismatch_priorEsqlAPIDataview(t *testing.T) {
	t.Parallel()

	show := true
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0{
		DataViewId:        "logs-view",
		ViewType:          kbapi.Dataview,
		ShowDistributions: &show,
	}
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats0(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByEsql: &models.FieldStatsTableByEsqlModel{
			Query: stringVal("FROM logs"),
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.Nil(t, pm.FieldStatsTableConfig.ByEsql)
	require.NotNil(t, pm.FieldStatsTableConfig.ByDataview)
	assert.Equal(t, "logs-view", pm.FieldStatsTableConfig.ByDataview.DataViewID.ValueString())
}

func TestHandler_fromAPI_invalidViewType(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"view_type":"unknown","data_view_id":"logs-view"}`)
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, json.Unmarshal(raw, &cfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	pm := panelModelBase()
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, nil, item)
	require.True(t, diags.HasError())
	require.Contains(t, diagSummary(diags), "view_type")
}

func TestHandler_fromAPI_import_byDataview(t *testing.T) {
	t.Parallel()

	show := true
	title := "Field statistics — logs view"
	description := "Field stats table panel (dataview)"
	hideTitle := false
	hideBorder := true
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0{
		DataViewId:        "logs-view",
		ViewType:          kbapi.Dataview,
		ShowDistributions: &show,
		Title:             &title,
		Description:       &description,
		HideTitle:         &hideTitle,
		HideBorder:        &hideBorder,
		TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: "now-24h",
			To:   "now",
			Mode: &mode,
		},
	}
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats0(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	pm := panelModelBase()
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.FieldStatsTableConfig)
	require.Nil(t, pm.FieldStatsTableConfig.ByEsql)
	dv := pm.FieldStatsTableConfig.ByDataview
	require.NotNil(t, dv)
	assert.Equal(t, "logs-view", dv.DataViewID.ValueString())
	assert.True(t, dv.ShowDistributions.ValueBool())
	assert.Equal(t, title, dv.Title.ValueString())
	assert.Equal(t, description, dv.Description.ValueString())
	assert.False(t, dv.HideTitle.ValueBool())
	assert.True(t, dv.HideBorder.ValueBool())
	require.NotNil(t, dv.TimeRange)
	assert.Equal(t, "now-24h", dv.TimeRange.From.ValueString())
	assert.Equal(t, "now", dv.TimeRange.To.ValueString())
	assert.Equal(t, "relative", dv.TimeRange.Mode.ValueString())
}

func TestHandler_fromAPI_import_byEsql(t *testing.T) {
	t.Parallel()

	show := false
	title := "Field statistics — logs by service"
	description := "Field stats table panel (esql)"
	hideTitle := true
	hideBorder := false
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1{
		ViewType:          kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1ViewTypeEsql,
		ShowDistributions: &show,
		Title:             &title,
		Description:       &description,
		HideTitle:         &hideTitle,
		HideBorder:        &hideBorder,
		TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: "now-24h",
			To:   "now",
			Mode: &mode,
		},
	}
	apiCfg.Query.Esql = "FROM logs-* | STATS count = COUNT(*) BY service.name"
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats1(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	pm := panelModelBase()
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, nil, item)
	require.False(t, diags.HasError(), "%s", diags)

	require.NotNil(t, pm.FieldStatsTableConfig)
	require.Nil(t, pm.FieldStatsTableConfig.ByDataview)
	esql := pm.FieldStatsTableConfig.ByEsql
	require.NotNil(t, esql)
	assert.Equal(t, "FROM logs-* | STATS count = COUNT(*) BY service.name", esql.Query.ValueString())
	assert.False(t, esql.ShowDistributions.ValueBool())
	assert.Equal(t, title, esql.Title.ValueString())
	assert.Equal(t, description, esql.Description.ValueString())
	assert.True(t, esql.HideTitle.ValueBool())
	assert.False(t, esql.HideBorder.ValueBool())
	require.NotNil(t, esql.TimeRange)
	assert.Equal(t, "now-24h", esql.TimeRange.From.ValueString())
	assert.Equal(t, "now", esql.TimeRange.To.ValueString())
	assert.Equal(t, "relative", esql.TimeRange.Mode.ValueString())
}

func TestHandler_fromAPI_nullPreservation_esqlTimeRange(t *testing.T) {
	t.Parallel()

	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1{
		ViewType: kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1ViewTypeEsql,
	}
	apiCfg.Query.Esql = "FROM logs"
	mode := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaModeRelative
	apiCfg.TimeRange = &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: "now-24h",
		To:   "now",
		Mode: &mode,
	}
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats1(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByEsql: &models.FieldStatsTableByEsqlModel{
			Query:     stringVal("FROM logs"),
			TimeRange: nil,
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	require.Nil(t, pm.FieldStatsTableConfig.ByEsql.TimeRange)
}

func TestHandler_fromAPI_nullPreservation_showDistributions(t *testing.T) {
	t.Parallel()

	show := true
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1{
		ViewType:          kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1ViewTypeEsql,
		ShowDistributions: &show,
	}
	apiCfg.Query.Esql = "FROM logs"
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats1(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByEsql: &models.FieldStatsTableByEsqlModel{
			Query:             stringVal("FROM logs"),
			ShowDistributions: boolNull(),
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	assert.True(t, pm.FieldStatsTableConfig.ByEsql.ShowDistributions.IsNull())
}

func TestHandler_fromAPI_nullPreservation_dataviewShowDistributions(t *testing.T) {
	t.Parallel()

	show := true
	apiCfg := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0{
		DataViewId:        "logs-view",
		ViewType:          kbapi.Dataview,
		ShowDistributions: &show,
	}
	var cfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats
	require.NoError(t, cfg.FromKibanaHTTPAPIsDataVisualizerFieldStats0(apiCfg))

	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
		Type:   kbapi.FieldStatsTable,
		Config: cfg,
	}
	var item kbapi.DashboardPanelItem
	require.NoError(t, item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel))

	prior := panelModelBase()
	prior.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID:        stringVal("logs-view"),
			ShowDistributions: boolNull(),
		},
	}

	pm := prior
	pm.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(prior.FieldStatsTableConfig)
	diags := fieldstatstable.Handler{}.FromAPI(context.Background(), &pm, &prior, item)
	require.False(t, diags.HasError(), "%s", diags)
	assert.True(t, pm.FieldStatsTableConfig.ByDataview.ShowDistributions.IsNull())
}

func TestHandler_toAPI_rejectsConfigJSON(t *testing.T) {
	t.Parallel()

	pm := panelModelBase()
	pm.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByDataview: &models.FieldStatsTableByDataviewModel{
			DataViewID: stringVal("logs-view"),
		},
	}
	pm.ConfigJSON = configJSONSet("{}")

	_, diags := fieldstatstable.Handler{}.ToAPI(pm, nil)
	require.True(t, diags.HasError())
	require.Contains(t, diagSummary(diags), "config_json")
}

func TestHandler_roundTrip_jsonStable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pm := panelModelBase()
	pm.FieldStatsTableConfig = &models.FieldStatsTableConfigModel{
		ByEsql: &models.FieldStatsTableByEsqlModel{
			Query:             stringVal("FROM logs-* | LIMIT 10"),
			ShowDistributions: boolVal(false),
		},
	}

	item1, diags := fieldstatstable.Handler{}.ToAPI(pm, nil)
	require.False(t, diags.HasError(), "%s", diags)

	next := pm
	next.FieldStatsTableConfig = deepCopyFieldStatsTableConfig(pm.FieldStatsTableConfig)
	require.False(t, fieldstatstable.Handler{}.FromAPI(ctx, &next, &pm, item1).HasError())

	item2, diags2 := fieldstatstable.Handler{}.ToAPI(next, nil)
	require.False(t, diags2.HasError(), "%s", diags2)

	raw1, err := item1.MarshalJSON()
	require.NoError(t, err)
	raw2, err := item2.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(raw1), string(raw2))
}
