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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

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
		lensDashboardAppByValueModel{},
		lensDashboardAPIGrid{},
		nil,
	)
	if !diags.HasError() {
		t.Fatal("expected error for unknown by_value.config_json")
	}
}

func TestLensDashboardAppByValueToAPI_sendsConfigAsAPIConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	raw := `{"type":"metric","title":"t",` +
		`"data_source":{"type":"data_view_spec","index_pattern":"metrics-*","time_field":"@timestamp"},` +
		`"filters":[],"metrics":[],"query":{"language":"kql","expression":""},` +
		`"styling":{"icon":{"name":"heart"}},"time_range":{"from":"now-15m","to":"now"}}`
	byValue := lensDashboardAppByValueModel{ConfigJSON: jsontypes.NewNormalizedValue(raw)}
	item, diags := lensDashboardAppByValueToAPI(byValue, lensDashboardAPIGrid{X: 1, Y: 2, W: float32ptr(8), H: float32ptr(9)}, new("pid"))
	require.False(t, diags.HasError())
	ld, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	require.NoError(t, err)
	var back map[string]any
	require.NoError(t, json.Unmarshal(mustJSON(t, ld.Config), &back))
	require.Equal(t, "metric", back["type"])
	require.Equal(t, "t", back["title"])

	// Read path: chart discriminator => by_value.config_json
	pm := &panelModel{}
	prior := &lensDashboardAppConfigModel{ByValue: &lensDashboardAppByValueModel{ConfigJSON: jsontypes.NewNormalizedValue(`{}`)}}
	diags = populateLensDashboardAppFromAPI(ctx, pm, &panelModel{LensDashboardAppConfig: prior}, ld)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByValue)
}

func TestLensDashboardAppByReferenceToAPI_mapsFields(t *testing.T) {
	t.Parallel()
	mode := "relative"
	byRef := lensDashboardAppByReferenceModel{
		RefID: types.StringValue("lensRef"),
		TimeRange: lensDashboardAppTimeRangeModel{
			From: types.StringValue("2024-01-01T00:00:00.000Z"),
			To:   types.StringValue("2024-01-01T01:00:00.000Z"),
			Mode: types.StringValue(mode),
		},
		ReferencesJSON: jsontypes.NewNormalizedValue(`[{"id":"abc","name":"lensRef","type":"lens"}]`),
		Title:          types.StringValue("T"),
		Description:    types.StringValue("D"),
		HideTitle:      types.BoolValue(true),
		HideBorder:     types.BoolValue(false),
		DrilldownsJSON: jsontypes.NewNormalizedValue(`[{"type":"dashboard_drilldown","trigger":"on_apply_filter","label":"x","dashboard_id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}]`),
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
	require.NotNil(t, cfg1.Drilldowns)
	require.Len(t, *cfg1.Drilldowns, 1)
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
	pm := &panelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, pm, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.LensDashboardAppConfig.ByReference)
	br := pm.LensDashboardAppConfig.ByReference
	require.Equal(t, "r1", br.RefID.ValueString())
	require.Equal(t, "absolute", br.TimeRange.Mode.ValueString())
	require.Equal(t, "T2", br.Title.ValueString())
}

func TestPopulateLensDashboardAppFromAPI_byValueOnAmbiguousNoPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Ambiguous: ref_id only, no chart type, incomplete time range
	var cfgUnion kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	require.NoError(t, cfgUnion.UnmarshalJSON([]byte(`{"ref_id":"only"}`)))
	api := kbapi.KbnDashboardPanelTypeLensDashboardApp{Config: cfgUnion}
	pm := &panelModel{}
	diags := populateLensDashboardAppFromAPI(ctx, pm, nil, api)
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
	pm := &panelModel{}
	prior := &lensDashboardAppConfigModel{
		ByReference: &lensDashboardAppByReferenceModel{RefID: types.StringValue("kept")},
	}
	tf := &panelModel{LensDashboardAppConfig: prior}
	diags := populateLensDashboardAppFromAPI(ctx, pm, tf, api)
	require.False(t, diags.HasError())
	// No population on ambiguous + prior by_reference: pm stays without new ByValue/ByReference from API
	require.Nil(t, pm.LensDashboardAppConfig)
}

func float32ptr(f float32) *float32 { return new(f) }

func mustJSON(t *testing.T, v kbapi.KbnDashboardPanelTypeLensDashboardApp_Config) []byte {
	t.Helper()
	b, err := v.MarshalJSON()
	require.NoError(t, err)
	return b
}
