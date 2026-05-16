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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestJsonValuePriorEmbeddedInExpandedCurrent(t *testing.T) {
	t.Parallel()
	prior := `{"type":"metric","title":"t"}`
	current := `{"type":"metric","title":"t","_k":["injected"]}`
	ok, err := jsonValuePriorEmbeddedInExpandedCurrent(prior, current)
	require.NoError(t, err)
	require.True(t, ok)

	changed := `{"type":"metric","title":"other"}`
	ok, err = jsonValuePriorEmbeddedInExpandedCurrent(prior, changed)
	require.NoError(t, err)
	require.False(t, ok)

	emptyType := `{"type":"","title":"x"}`
	current2 := `{"type":"","title":"x","k":1}`
	ok, err = jsonValuePriorEmbeddedInExpandedCurrent(emptyType, current2)
	require.NoError(t, err)
	require.False(t, ok)

	emptyNoChart := `{"k":1}`
	current3 := `{"k":1,"type":"x"}`
	ok, err = jsonValuePriorEmbeddedInExpandedCurrent(emptyNoChart, current3)
	require.NoError(t, err)
	require.False(t, ok)

	stylingA := `{"type":"metric","title":"t","styling":{"icon":{"name":"heart"}},"metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}]}`
	stylingB := `{"type":"metric","title":"t","styling":{"primary":{}},"metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}]}`
	ok, err = jsonValuePriorEmbeddedInExpandedCurrent(stylingA, stylingB)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPreservePriorLensByValueConfigJSON_enrichment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := `{"type":"metric","a":1}`
	enriched := `{"type":"metric","a":1,"b":2}`
	priorN := jsontypes.NewNormalizedValue(prior)
	fromAPI := jsontypes.NewNormalizedValue(enriched)
	var diags diag.Diagnostics
	out := preservePriorLensByValueConfigJSON(ctx, priorN, fromAPI, &diags)
	require.False(t, diags.HasError())
	require.Equal(t, prior, out.ValueString())
}

func TestJsonValuePriorEmbedded_filtersOmittedOrNull(t *testing.T) {
	t.Parallel()
	base := `{"type":"metric","title":"x","metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}]`
	priorEmpty := base + `,"filters":[]}`
	currentOmit := base + `}`
	currentNull := base + `,"filters":null}`
	ok, err := jsonValuePriorEmbeddedInExpandedCurrent(priorEmpty, currentOmit)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = jsonValuePriorEmbeddedInExpandedCurrent(priorEmpty, currentNull)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestJsonValuePriorEmbedded_defaultKqlQueryOmittedOnRead(t *testing.T) {
	t.Parallel()
	prior := `{"type":"metric","title":"t","metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}],` +
		`"query":{"language":"kql","expression":""}}`
	current := `{"type":"metric","title":"t","metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}]}`
	ok, err := jsonValuePriorEmbeddedInExpandedCurrent(prior, current)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestJsonValuePriorEmbedded_nonEmptyArrayWhenPriorEmptyRejects(t *testing.T) {
	t.Parallel()
	// User-authored `metrics: []` vs API with metrics — not a value-subset; must not preserve an empty list that would strip API data.
	prior := `{"type":"metric","title":"t","metrics":[]}`
	current := `{"type":"metric","title":"t","metrics":[{"type":"primary","operation":"count","format":{"type":"number"}}]}`
	ok, err := jsonValuePriorEmbeddedInExpandedCurrent(prior, current)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestIsOmissibleDefaultKqlQuery(t *testing.T) {
	t.Parallel()
	require.True(t, isOmissibleDefaultKqlQuery(nil))
	require.True(t, isOmissibleDefaultKqlQuery(map[string]any{}))
	require.True(t, isOmissibleDefaultKqlQuery(map[string]any{"language": "kql"}))
	require.True(t, isOmissibleDefaultKqlQuery(map[string]any{"language": "kql", "expression": ""}))
	require.False(t, isOmissibleDefaultKqlQuery(map[string]any{"language": "lucene", "expression": ""}))
	require.False(t, isOmissibleDefaultKqlQuery(map[string]any{"language": "kql", "expression": "host:*"}))
}

func TestPopulateLensDashboardAppFromAPI_byValuePreservesPractitionerEnriched(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorStr := `{"data_source":{"index_pattern":"metrics-*","time_field":"@timestamp","type":"data_view_spec"},` +
		`"filters":[],"metrics":[{"format":{"type":"number"},"operation":"count","type":"primary"}],` +
		`"query":{"expression":"","language":"kql"},"styling":{"icon":{"name":"heart"}},` +
		`"time_range":{"from":"now-15m","to":"now"},"title":"Acc by-value","type":"metric"}`
	apiStr := `{"time_range":{"from":"now-15m","to":"now"},"title":"Acc by-value",` +
		`"data_source":{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"},` +
		`"type":"metric","sampling":1,"ignore_global_filters":false,` +
		`"metrics":[{"type":"primary","operation":"count","empty_as_null":false,` +
		`"format":{"type":"number","decimals":2,"compact":false},"color":{"type":"auto"}}],` +
		`"styling":{"primary":{"position":"bottom","labels":{"alignment":"left"},` +
		`"value":{"sizing":"auto","alignment":"right"}}}}`
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON([]byte(apiStr)))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	tfPanel := &models.PanelModel{
		LensDashboardAppConfig: &models.LensDashboardAppConfigModel{
			ByValue: &models.LensDashboardAppByValueModel{
				ConfigJSON: jsontypes.NewNormalizedValue(priorStr),
			},
		},
	}
	pm := &models.PanelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, nil, pm, tfPanel, api)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig)
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.True(t, typeutils.IsKnown(pm.LensDashboardAppConfig.ByValue.ConfigJSON))
	require.Equal(t, priorStr, pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString())
}

func TestClassifyLensDashboardAppConfigFromRoot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		root map[string]any
		want lensConfigClass
	}{
		{
			name: "by_value chart with ref_id and time_range still chart",
			root: map[string]any{
				"type":   "xyChartESQL",
				"ref_id": "panel_0",
				"time_range": map[string]any{
					"from": "now-7d", "to": "now",
				},
			},
			want: lensConfigClassByValueChart,
		},
		{
			name: "by_reference without top-level chart type",
			root: map[string]any{
				"ref_id": "panel_0",
				"time_range": map[string]any{
					"from": "now-7d", "to": "now",
				},
			},
			want: lensConfigClassByReference,
		},
		{
			name: "ambiguous incomplete",
			root: map[string]any{"ref_id": "x"},
			want: lensConfigClassAmbiguous,
		},
		{
			name: "by_value empty type string not chart",
			root: map[string]any{"type": ""},
			want: lensConfigClassAmbiguous,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := classifyLensDashboardAppConfigFromRoot(tc.root)
			if got != tc.want {
				t.Fatalf("classify: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestLensDashboardAppByValueToAPI_UnknownConfigJSON(t *testing.T) {
	t.Parallel()
	_, diags := lensDashboardAppByValueToAPI(
		models.LensDashboardAppByValueModel{},
		lensDashboardAPIGrid{},
		nil,
		nil,
	)
	if !diags.HasError() {
		t.Fatal("expected error for unknown by_value.config_json")
	}
}

func TestLensDashboardAppByValueToAPI_sendsConfigAsAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	raw := `{"type":"metric","title":"t",` +
		`"data_source":{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"},` +
		`"filters":[],"metrics":[],"query":{"language":"kql","expression":""},` +
		`"styling":{"icon":{"name":"heart"}},"time_range":{"from":"now-15m","to":"now"}}`
	byValue := models.LensDashboardAppByValueModel{ConfigJSON: jsontypes.NewNormalizedValue(raw)}
	item, diags := lensDashboardAppByValueToAPI(byValue, lensDashboardAPIGrid{X: 1, Y: 2, W: float32ptr(8), H: float32ptr(9)}, new("pid"), nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	var back map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &back))
	require.Equal(t, "metric", back["type"])
	require.Equal(t, "t", back["title"])

	// Read path: chart discriminator => by_value.config_json
	pm := &models.PanelModel{}
	prior := &models.LensDashboardAppConfigModel{ByValue: &models.LensDashboardAppByValueModel{ConfigJSON: jsontypes.NewNormalizedValue(`{}`)}}
	diags = populateLensDashboardAppFromAPI(ctx, nil, pm, &models.PanelModel{LensDashboardAppConfig: prior}, ld)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	var readRoot map[string]any
	require.NoError(t, json.Unmarshal([]byte(pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString()), &readRoot))
	require.Equal(t, "metric", readRoot["type"])
	require.Equal(t, "t", readRoot["title"])
}

func TestLensDashboardAppByReferenceToAPI_mapsFields(t *testing.T) {
	t.Parallel()
	mode := "relative"
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
			Mode: types.StringValue(mode),
		},
		ReferencesJSON: jsontypes.NewNormalizedValue(`[{"id":"abc","name":"lensRef","type":"lens"}]`),
		Title:          types.StringValue("T"),
		Description:    types.StringValue("D"),
		HideTitle:      types.BoolValue(true),
		HideBorder:     types.BoolValue(false),
		Drilldowns: models.DrilldownsModel{
			{
				Dashboard: &models.DrilldownDashboardBlockModel{
					DashboardID: types.StringValue("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
					Label:       types.StringValue("x"),
				},
			},
		},
	}
	item, diags := lensDashboardAppByReferenceToAPI(byRef, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	require.Equal(t, "lensRef", cfg1.RefId)
	require.Equal(t, "2024-01-01T00:00:00.000Z", cfg1.TimeRange.From)
	require.Equal(t, "2024-01-01T01:00:00.000Z", cfg1.TimeRange.To)
	require.NotNil(t, cfg1.TimeRange.Mode)
	require.Equal(t, kbapi.KbnEsQueryServerTimeRangeSchemaModeRelative, *cfg1.TimeRange.Mode)
	require.Len(t, *cfg1.References, 1)
	require.Equal(t, "abc", (*cfg1.References)[0].Id)
	require.NotNil(t, cfg1.Title)
	require.Equal(t, "T", *cfg1.Title)
	require.NotNil(t, cfg1.Description)
	require.Equal(t, "D", *cfg1.Description)
	require.NotNil(t, cfg1.HideTitle)
	require.True(t, *cfg1.HideTitle)
	require.NotNil(t, cfg1.HideBorder)
	require.False(t, *cfg1.HideBorder)
	require.NotNil(t, cfg1.Drilldowns)
	require.Len(t, *cfg1.Drilldowns, 1)
}

func TestLensDashboardAppByReferenceToAPI_emptyStructuredDrilldowns_sendsEmptyArray(t *testing.T) {
	t.Parallel()
	// Mirrors Terraform `drilldowns = []`: framework reflects a known-empty nested list attribute as non-nil slice (len 0).
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
		},
		Drilldowns: explicitEmptyDrilldowns(),
	}
	item, diags := lensDashboardAppByReferenceToAPI(byRef, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	require.NotNil(t, cfg1.Drilldowns, "explicit [] must set Drilldowns so the API can clear prior drilldowns")
	require.Empty(t, *cfg1.Drilldowns)
	var wire map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &wire))
	require.Contains(t, wire, "drilldowns")
	require.Equal(t, []any{}, wire["drilldowns"])
}

func TestLensDashboardAppByReferenceToAPI_omittedStructuredDrilldowns_nilSliceSkipsAPIField(t *testing.T) {
	t.Parallel()
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
		},
	}
	item, diags := lensDashboardAppByReferenceToAPI(byRef, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	require.Nil(t, cfg1.Drilldowns)
	var wire map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &wire))
	_, has := wire["drilldowns"]
	require.False(t, has, "omitted drills should omit API drilldowns key/json field where possible")
}

func TestLensDashboardAppByReferenceToAPI_emptyReferencesJSON_sendsEmptyArray(t *testing.T) {
	t.Parallel()
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
		},
		ReferencesJSON: jsontypes.NewNormalizedValue(`[]`),
	}
	item, diags := lensDashboardAppByReferenceToAPI(byRef, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	require.NotNil(t, cfg1.References)
	require.Empty(t, *cfg1.References)
	var wireRefs map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &wireRefs))
	require.Contains(t, wireRefs, "references")
	require.Equal(t, []any{}, wireRefs["references"])
}

func TestPopulateLensDashboardAppFromAPI_byReferencePath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := kbapi.KbnEsQueryServerTimeRangeSchemaModeAbsolute
	cfg1 := kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1{
		RefId: "r1",
		TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{
			From: "a",
			To:   "b",
			Mode: &m,
		},
		Title:       new("T2"),
		Description: new("D2"),
	}
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.FromKbnDashboardPanelTypeLensDashboardAppConfig1(cfg1))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	pm := &models.PanelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, nil, pm, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByReference)
	br := pm.LensDashboardAppConfig.ByReference
	require.Equal(t, "r1", br.RefID.ValueString())
	require.Equal(t, "a", br.TimeRange.From.ValueString())
	require.Equal(t, "b", br.TimeRange.To.ValueString())
	require.Equal(t, "absolute", br.TimeRange.Mode.ValueString())
	require.Equal(t, "T2", br.Title.ValueString())
	require.Equal(t, "D2", br.Description.ValueString())
	require.Nil(t, br.Drilldowns)
}

func TestPopulateLensDashboardAppFromAPI_byReference_keepsPriorDrilldownsWhenAPIOmits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorWant := models.DrilldownsModel{
		{
			Dashboard: &models.DrilldownDashboardBlockModel{
				DashboardID: types.StringValue("dddddddd-dddd-dddd-dddd-dddddddddddd"),
				Label:       types.StringValue("kept-drill"),
			},
		},
	}
	tf := models.PanelModel{
		LensDashboardAppConfig: &models.LensDashboardAppConfigModel{
			ByReference: &models.LensDashboardAppByReferenceModel{
				RefID:      types.StringValue("r1"),
				TimeRange:  models.LensDashboardAppTimeRangeModel{From: types.StringValue("a"), To: types.StringValue("b")},
				Drilldowns: priorWant,
			},
		},
	}
	pm := tf
	apiWire := []byte(`{"ref_id":"r1","time_range":{"from":"a","to":"b"}}`)
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON(apiWire))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	diags := populateLensDashboardAppFromAPI(ctx, nil, &pm, &tf, api)
	require.False(t, diags.HasError())
	got := pm.LensDashboardAppConfig.ByReference.Drilldowns
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Dashboard)
	assertDashboardBlocksEqual(t, priorWant[0].Dashboard, got[0].Dashboard)
}

func TestLensDashboardAppByReferenceToAPI_discoverAndURLKinds(t *testing.T) {
	t.Parallel()
	byRef := models.LensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: models.LensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
		},
		Drilldowns: models.DrilldownsModel{
			{
				Discover: &models.DrilldownDiscoverBlockModel{
					Label:        types.StringValue("Open Discover"),
					OpenInNewTab: types.BoolValue(false),
				},
			},
			{
				URL: &models.DrilldownURLBlockModel{
					URL:       types.StringValue("https://example.com/{{event.field}}"),
					Label:     types.StringValue("Open URL"),
					Trigger:   types.StringValue("on_click_value"),
					EncodeURL: types.BoolValue(true),
				},
			},
		},
	}
	item, diags := lensDashboardAppByReferenceToAPI(byRef, lensDashboardAPIGrid{}, nil)
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	cfg1, err := ld.Config.AsKbnDashboardPanelTypeLensDashboardAppConfig1()
	require.NoError(t, err)
	require.Len(t, *cfg1.Drilldowns, 2)
}

func TestPopulateLensDashboardAppFromAPI_byReferenceRead_drilldowns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	wire := `{
		"ref_id": "r1",
		"time_range": {"from": "x1", "to": "x2"},
		"drilldowns": [
			{
				"type": "dashboard_drilldown",
				"trigger": "on_apply_filter",
				"label": "Drill label",
				"dashboard_id": "dddddddd-dddd-dddd-dddd-dddddddddddd"
			},
			{
				"type": "url_drilldown",
				"url": "https://example.com/",
				"label": "U",
				"trigger": "on_click_value"
			},
			{
				"type": "discover_drilldown",
				"trigger": "on_apply_filter",
				"label": "Discover me"
			}
		]
	}`
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON([]byte(wire)))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	pm := &models.PanelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, nil, pm, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByReference)
	dd := pm.LensDashboardAppConfig.ByReference.Drilldowns
	require.Len(t, dd, 3)
	require.NotNil(t, dd[0].Dashboard)
	require.Equal(t, "dddddddd-dddd-dddd-dddd-dddddddddddd", dd[0].Dashboard.DashboardID.ValueString())
	require.Equal(t, "Drill label", dd[0].Dashboard.Label.ValueString())
	require.NotNil(t, dd[1].URL)
	require.Equal(t, "https://example.com/", dd[1].URL.URL.ValueString())
	require.Equal(t, "U", dd[1].URL.Label.ValueString())
	require.Equal(t, "on_click_value", dd[1].URL.Trigger.ValueString())
	require.NotNil(t, dd[2].Discover)
	require.Equal(t, "Discover me", dd[2].Discover.Label.ValueString())
}

func TestPopulateLensDashboardAppFromAPI_byValueOnAmbiguousNoPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Ambiguous: ref_id only, no chart type, incomplete time range
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON([]byte(`{"ref_id":"only"}`)))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	pm := &models.PanelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, nil, pm, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
	require.Contains(t, pm.LensDashboardAppConfig.ByValue.ConfigJSON.ValueString(), `"ref_id"`)
}

func TestPopulateLensDashboardAppFromAPI_ambiguousPreservesPriorByReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON([]byte(`{"ref_id":"only"}`)))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	prior := &models.LensDashboardAppConfigModel{
		ByReference: &models.LensDashboardAppByReferenceModel{RefID: types.StringValue("kept")},
	}
	tf := &models.PanelModel{LensDashboardAppConfig: prior}
	// Match `mapPanelFromAPI`: seed `pm` from the prior plan/state panel before converters run.
	pm := *tf
	diags := populateLensDashboardAppFromAPI(ctx, nil, &pm, tf, api)
	require.False(t, diags.HasError())
	// Ambiguous API + prior by_reference: no rewrite from API; prior block stays in the seeded panel (REQ-009).
	require.NotNil(t, pm.LensDashboardAppConfig)
	require.NotNil(t, pm.LensDashboardAppConfig.ByReference)
	require.Equal(t, "kept", pm.LensDashboardAppConfig.ByReference.RefID.ValueString())
	require.Nil(t, pm.LensDashboardAppConfig.ByValue)
}

func float32ptr(f float32) *float32 { return new(f) }

func mustJSON(t *testing.T, v kbapi.KbnDashboardPanelTypeLensDashboardApp_Config) []byte {
	t.Helper()
	b, err := v.MarshalJSON()
	require.NoError(t, err)
	return b
}
