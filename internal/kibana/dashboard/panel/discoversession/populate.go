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

package discoversession

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// discoverSessionAPIConfigLooksByReference distinguishes inline vs linked Discover session configs.
// kbapi's Config union unmarshals successfully into both generated structs, so we key off JSON `ref_id`
// (present only on the by-reference branch).
func discoverSessionAPIConfigLooksByReference(apiCfg kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config) bool {
	raw, err := apiCfg.MarshalJSON()
	if err != nil {
		return false
	}
	var probe struct {
		RefID string `json:"ref_id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return strings.TrimSpace(probe.RefID) != ""
}

// discoverSessionPriorTFBranchMismatchesAPI reports out-of-band branch changes (e.g. Kibana flipped
// inline vs linked). Prior Terraform state used exclusively one branch while the API payload uses the other.
func discoverSessionPriorTFBranchMismatchesAPI(apiLooksByRef bool, prior *models.DiscoverSessionPanelConfigModel) bool {
	if prior == nil {
		return false
	}
	hasValue := prior.ByValue != nil
	hasRef := prior.ByReference != nil
	if apiLooksByRef && hasValue && !hasRef {
		return true
	}
	if !apiLooksByRef && hasRef && !hasValue {
		return true
	}
	return false
}

func populateDiscoverSessionPanelFromAPI(ctx context.Context, pm *models.PanelModel, tfPanel *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession) diag.Diagnostics {
	if tfPanel == nil {
		cfg, diags := discoverSessionPanelConfigFromAPIImport(ctx, apiPanel, nil)
		pm.DiscoverSessionConfig = cfg
		return diags
	}

	if pm.DiscoverSessionConfig == nil {
		cfg, diags := discoverSessionPanelConfigFromAPIImport(ctx, apiPanel, tfPanel)
		pm.DiscoverSessionConfig = cfg
		if tfPanel == nil {
			return diags
		}
	}

	existing := pm.DiscoverSessionConfig
	if existing == nil {
		return nil
	}

	prior := tfPanel.DiscoverSessionConfig
	apiByRef := discoverSessionAPIConfigLooksByReference(apiPanel.Config)

	if discoverSessionPriorTFBranchMismatchesAPI(apiByRef, prior) {
		// Drift import: replace typed config from API so the next plan surfaces the branch change.
		if apiByRef {
			cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1()
			if err != nil {
				return discoverSessionDecodeDiagnostics(err, "by-reference")
			}
			imported, imDiags := discoverSessionConfig1FromAPIImport(ctx, cfg1)
			if imported != nil {
				*existing = *imported
			}
			return imDiags
		}
		cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0()
		if err != nil {
			return discoverSessionDecodeDiagnostics(err, "by-value")
		}
		var priorTab *models.DiscoverSessionTabModel
		if tfPanel != nil && tfPanel.DiscoverSessionConfig != nil && tfPanel.DiscoverSessionConfig.ByValue != nil {
			priorTab = &tfPanel.DiscoverSessionConfig.ByValue.Tab
		}
		imported, tabDiags := discoverSessionConfig0FromAPIImport(ctx, cfg0, priorTab)
		if imported != nil {
			*existing = *imported
		}
		return tabDiags
	}

	if apiByRef {
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1()
		if err != nil {
			return discoverSessionDecodeDiagnostics(err, "by-reference")
		}
		return discoverSessionMergeConfig1FromAPI(ctx, existing, tfPanel, cfg1)
	}

	cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0()
	if err != nil {
		return discoverSessionDecodeDiagnostics(err, "by-value")
	}
	return discoverSessionMergeConfig0FromAPI(ctx, existing, tfPanel, cfg0)
}

// discoverSessionDecodeDiagnostics builds a diagnostic for failed kbapi union decoding so
// callers do not silently lose state during read/refresh.
func discoverSessionDecodeDiagnostics(err error, branch string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode discover_session API config",
		"Could not decode the API discover_session "+branch+" config: "+err.Error(),
	)
	return diags
}

func discoverSessionPanelConfigFromAPIImport(
	ctx context.Context,
	apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession,
	tfPanel *models.PanelModel,
) (*models.DiscoverSessionPanelConfigModel, diag.Diagnostics) {
	var priorTab *models.DiscoverSessionTabModel
	if tfPanel != nil && tfPanel.DiscoverSessionConfig != nil && tfPanel.DiscoverSessionConfig.ByValue != nil {
		priorTab = &tfPanel.DiscoverSessionConfig.ByValue.Tab
	}

	if discoverSessionAPIConfigLooksByReference(apiPanel.Config) {
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1()
		if err == nil {
			return discoverSessionConfig1FromAPIImport(ctx, cfg1)
		}
		return nil, nil
	}

	cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0()
	if err == nil {
		return discoverSessionConfig0FromAPIImport(ctx, cfg0, priorTab)
	}
	return nil, nil
}

func discoverSessionConfig0FromAPIImport(
	ctx context.Context,
	cfg0 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0,
	priorTab *models.DiscoverSessionTabModel,
) (*models.DiscoverSessionPanelConfigModel, diag.Diagnostics) {
	tab, tabDiags := discoverSessionTabFromAPIConfig0(ctx, cfg0.Tabs, priorTab)
	cfg := &models.DiscoverSessionPanelConfigModel{
		Title:       types.StringPointerValue(cfg0.Title),
		Description: types.StringPointerValue(cfg0.Description),
		HideTitle:   types.BoolPointerValue(cfg0.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg0.HideBorder),
		ByValue: &models.DiscoverSessionPanelByValueModel{
			TimeRange: discoverSessionTimeRangePtrFromAPI(cfg0.TimeRange),
			Tab:       tab,
		},
	}
	cfg.Drilldowns = readDiscoverSessionDrilldownsFromConfig0(cfg0.Drilldowns, nil)
	return cfg, tabDiags
}

func discoverSessionConfig1FromAPIImport(ctx context.Context, cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1) (*models.DiscoverSessionPanelConfigModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := &models.DiscoverSessionPanelConfigModel{
		Title:       types.StringPointerValue(cfg1.Title),
		Description: types.StringPointerValue(cfg1.Description),
		HideTitle:   types.BoolPointerValue(cfg1.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg1.HideBorder),
		ByReference: &models.DiscoverSessionPanelByRefModel{
			TimeRange: discoverSessionTimeRangePtrFromAPI(cfg1.TimeRange),
			RefID:     types.StringValue(cfg1.RefId),
		},
	}
	cfg.ByReference.SelectedTabID = typeutils.StringishPointerValue(cfg1.SelectedTabId)
	if cfg1.Overrides != nil {
		var ovDiags diag.Diagnostics
		cfg.ByReference.Overrides, ovDiags = discoverSessionOverridesFromAPI(ctx, *cfg1.Overrides)
		diags.Append(ovDiags...)
	}
	cfg.Drilldowns = readDiscoverSessionDrilldownsFromConfig1(cfg1.Drilldowns, nil)
	return cfg, diags
}

func discoverSessionTimeRangePtrFromAPI(api *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema) *models.TimeRangeModel {
	if api == nil {
		return nil
	}
	tr := &models.TimeRangeModel{
		From: types.StringValue(api.From),
		To:   types.StringValue(api.To),
	}
	if api.Mode != nil {
		tr.Mode = types.StringValue(string(*api.Mode))
	} else {
		tr.Mode = types.StringNull()
	}
	return tr
}

func discoverSessionTabFromAPIConfig0(
	ctx context.Context,
	tabs []kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item,
	prior *models.DiscoverSessionTabModel,
) (models.DiscoverSessionTabModel, diag.Diagnostics) {
	if len(tabs) == 0 {
		return models.DiscoverSessionTabModel{}, nil
	}
	tab := tabs[0]

	if prior != nil && prior.ESQL != nil {
		if esql, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil {
			m, d := discoverSessionESQLTabFromAPI(ctx, esql)
			return models.DiscoverSessionTabModel{ESQL: m}, d
		}
	}
	if prior != nil && prior.DSL != nil {
		if dsl, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil {
			m, d := discoverSessionDSLTabFromAPI(ctx, dsl)
			return models.DiscoverSessionTabModel{DSL: m}, d
		}
	}

	switch discoverSessionTabBranchFromAPIItem(tab) {
	case discoverSessionTabBranchESQL:
		if esql, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil {
			m, d := discoverSessionESQLTabFromAPI(ctx, esql)
			return models.DiscoverSessionTabModel{ESQL: m}, d
		}
	case discoverSessionTabBranchDSL:
		if dsl, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil {
			m, d := discoverSessionDSLTabFromAPI(ctx, dsl)
			return models.DiscoverSessionTabModel{DSL: m}, d
		}
	}

	if dsl, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil {
		m, d := discoverSessionDSLTabFromAPI(ctx, dsl)
		return models.DiscoverSessionTabModel{DSL: m}, d
	}
	if esql, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil {
		m, d := discoverSessionESQLTabFromAPI(ctx, esql)
		return models.DiscoverSessionTabModel{ESQL: m}, d
	}
	return models.DiscoverSessionTabModel{}, nil
}

type discoverSessionTabBranch int

const (
	discoverSessionTabBranchUnknown discoverSessionTabBranch = iota
	discoverSessionTabBranchDSL
	discoverSessionTabBranchESQL
)

func discoverSessionTabBranchFromAPIItem(tab kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item) discoverSessionTabBranch {
	raw, err := tab.MarshalJSON()
	if err != nil {
		return discoverSessionTabBranchUnknown
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return discoverSessionTabBranchUnknown
	}
	ds, _ := root["data_source"].(map[string]any)
	if dsType, _ := ds["type"].(string); dsType == "esql" {
		return discoverSessionTabBranchESQL
	}
	if _, ok := root["query"]; ok {
		return discoverSessionTabBranchDSL
	}
	return discoverSessionTabBranchUnknown
}

func discoverSessionDSLTabFromAPI(ctx context.Context, api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0) (*models.DiscoverSessionDSLTabModel, diag.Diagnostics) {
	m := &models.DiscoverSessionDSLTabModel{}
	var diags diag.Diagnostics

	if api.ColumnOrder != nil && len(*api.ColumnOrder) > 0 {
		m.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	m.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)

	if api.Sort != nil {
		m.Sort = discoverSessionSortSliceFromAPI0(*api.Sort)
	}

	if api.Density != nil {
		m.Density = types.StringValue(string(*api.Density))
	} else {
		m.Density = types.StringNull()
	}

	m.HeaderRowHeight = discoverSessionDSLHeaderRowHeightFromAPI(api.HeaderRowHeight)
	m.RowHeight = discoverSessionDSLRowHeightFromAPI(api.RowHeight)

	if api.RowsPerPage != nil {
		m.RowsPerPage = types.Int64Value(int64(*api.RowsPerPage))
	} else {
		m.RowsPerPage = types.Int64Null()
	}
	if api.SampleSize != nil {
		m.SampleSize = types.Int64Value(int64(*api.SampleSize))
	} else {
		m.SampleSize = types.Int64Null()
	}

	if api.ViewMode != nil {
		m.ViewMode = types.StringValue(string(*api.ViewMode))
	} else {
		m.ViewMode = types.StringNull()
	}

	q := discoverSessionQueryFromKbnAsCode(api.Query)
	m.Query = &q

	dsBytes, err := json.Marshal(api.DataSource)
	if err == nil {
		m.DataSourceJSON = jsontypes.NewNormalizedValue(string(dsBytes))
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		filters := make([]models.ChartFilterJSONModel, 0, len(*api.Filters))
		for _, item := range *api.Filters {
			fm := models.ChartFilterJSONModel{}
			diags.Append(lenscommon.ChartFilterJSONPopulateFromAPIItem(&fm, item)...)
			filters = append(filters, fm)
		}
		m.Filters = filters
	}

	return m, diags
}

func discoverSessionESQLTabFromAPI(ctx context.Context, api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1) (*models.DiscoverSessionESQLTabModel, diag.Diagnostics) {
	m := &models.DiscoverSessionESQLTabModel{}
	var diags diag.Diagnostics

	if api.ColumnOrder != nil && len(*api.ColumnOrder) > 0 {
		m.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	m.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)

	if api.Sort != nil {
		m.Sort = discoverSessionSortSliceFromAPI1(*api.Sort)
	}

	if api.Density != nil {
		m.Density = types.StringValue(string(*api.Density))
	} else {
		m.Density = types.StringNull()
	}

	m.HeaderRowHeight = discoverSessionESQLHeaderRowHeightFromAPI(api.HeaderRowHeight)
	m.RowHeight = discoverSessionESQLRowHeightFromAPI(api.RowHeight)

	dsBytes, err := json.Marshal(api.DataSource)
	if err == nil {
		m.DataSourceJSON = jsontypes.NewNormalizedValue(string(dsBytes))
	}

	return m, diags
}

func discoverSessionQueryFromKbnAsCode(q *kbapi.KibanaHTTPAPIsKbnAsCodeQuery) models.FilterSimpleModel {
	if q == nil {
		return models.FilterSimpleModel{}
	}
	return models.FilterSimpleModel{
		Expression: types.StringValue(q.Expression),
		Language:   types.StringValue(string(q.Language)),
	}
}

func discoverSessionSortSliceFromAPI0(api []struct {
	Direction kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0SortDirection `json:"direction"`
	Name      string                                                                            `json:"name"`
}) []models.DiscoverSessionSortModel {
	out := make([]models.DiscoverSessionSortModel, len(api))
	for i, s := range api {
		out[i].Name = types.StringValue(s.Name)
		out[i].Direction = types.StringValue(string(s.Direction))
	}
	return out
}

func discoverSessionSortSliceFromAPI1(api []struct {
	Direction kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1SortDirection `json:"direction"`
	Name      string                                                                            `json:"name"`
}) []models.DiscoverSessionSortModel {
	out := make([]models.DiscoverSessionSortModel, len(api))
	for i, s := range api {
		out[i].Name = types.StringValue(s.Name)
		out[i].Direction = types.StringValue(string(s.Direction))
	}
	return out
}

func discoverSessionOverridesFromAPI(ctx context.Context, api struct {
	ColumnOrder    *[]string `json:"column_order,omitempty"`
	ColumnSettings *map[string]struct {
		Width *float32 `json:"width,omitempty"`
	} `json:"column_settings,omitempty"`
	Density         *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
	HeaderRowHeight *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
	RowHeight       *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
	RowsPerPage     *float32                                                                                     `json:"rows_per_page,omitempty"`
	SampleSize      *float32                                                                                     `json:"sample_size,omitempty"`
	Sort            *[]struct {
		Direction kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                                `json:"name"`
	} `json:"sort,omitempty"`
}) (*models.DiscoverSessionOverridesModel, diag.Diagnostics) {
	m := &models.DiscoverSessionOverridesModel{}
	var diags diag.Diagnostics

	if api.ColumnOrder != nil && len(*api.ColumnOrder) > 0 {
		m.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	m.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)

	if api.Sort != nil {
		out := make([]models.DiscoverSessionSortModel, len(*api.Sort))
		for i, s := range *api.Sort {
			out[i].Name = types.StringValue(s.Name)
			out[i].Direction = types.StringValue(string(s.Direction))
		}
		m.Sort = out
	}

	if api.Density != nil {
		m.Density = types.StringValue(string(*api.Density))
	} else {
		m.Density = types.StringNull()
	}

	m.HeaderRowHeight = discoverSessionOverridesHeaderRowHeightFromAPI(api.HeaderRowHeight)
	m.RowHeight = discoverSessionOverridesRowHeightFromAPI(api.RowHeight)

	if api.RowsPerPage != nil {
		m.RowsPerPage = types.Int64Value(int64(*api.RowsPerPage))
	} else {
		m.RowsPerPage = types.Int64Null()
	}
	if api.SampleSize != nil {
		m.SampleSize = types.Int64Value(int64(*api.SampleSize))
	} else {
		m.SampleSize = types.Int64Null()
	}

	return m, diags
}

func discoverSessionColumnSettingsFromAPI(ctx context.Context, api *map[string]struct {
	Width *float32 `json:"width,omitempty"`
}, diags *diag.Diagnostics) types.Map {
	if api == nil || len(*api) == 0 {
		return types.MapNull(columnSettingObjectType())
	}
	elems := make(map[string]models.DiscoverSessionColumnSettingModel, len(*api))
	for k, v := range *api {
		var width types.Float64
		if v.Width != nil {
			width = types.Float64Value(float64(*v.Width))
		} else {
			width = types.Float64Null()
		}
		elems[k] = models.DiscoverSessionColumnSettingModel{Width: width}
	}
	return typeutils.MapValueFrom(ctx, elems, columnSettingObjectType(), path.Empty(), diags)
}

func rowHeightFromAPI(isNil bool, decodeInt func() (float32, error), decodeStr func() (string, error)) types.String {
	if isNil {
		return types.StringNull()
	}
	if v, err := decodeInt(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := decodeStr(); err == nil {
		return types.StringValue(v)
	}
	return types.StringNull()
}

func discoverSessionDSLHeaderRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_HeaderRowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight1()
			return string(v), err
		},
	)
}

func discoverSessionDSLRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_RowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight1()
			return string(v), err
		},
	)
}

func discoverSessionESQLHeaderRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_HeaderRowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight1()
			return string(v), err
		},
	)
}

func discoverSessionESQLRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_RowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight1()
			return string(v), err
		},
	)
}

func discoverSessionOverridesHeaderRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight1()
			return string(v), err
		},
	)
}

func discoverSessionOverridesRowHeightFromAPI(h *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight) types.String {
	return rowHeightFromAPI(
		h == nil,
		func() (float32, error) {
			return h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight0()
		},
		func() (string, error) {
			v, err := h.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight1()
			return string(v), err
		},
	)
}

func readDiscoverSessionDrilldownsFromConfig0(
	api *[]struct {
		EncodeUrl    *bool                                                                            `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                                           `json:"label"`
		OpenInNewTab *bool                                                                            `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsType    `json:"type"`
		Url          string                                                                           `json:"url"` //nolint:revive
	},
	prior []models.DiscoverSessionPanelDrilldown,
) []models.DiscoverSessionPanelDrilldown {
	return panelkit.MapDiscoverSessionDrilldownsFromAPI(api, func(d struct {
		EncodeUrl    *bool                                                                            `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                                           `json:"label"`
		OpenInNewTab *bool                                                                            `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsType    `json:"type"`
		Url          string                                                                           `json:"url"` //nolint:revive
	}) panelkit.URLDrilldownAPIItemData {
		return panelkit.URLDrilldownAPIItemData{URL: d.Url, Label: d.Label, EncodeUrl: d.EncodeUrl, OpenInNewTab: d.OpenInNewTab}
	}, prior)
}

func readDiscoverSessionDrilldownsFromConfig1(
	api *[]struct {
		EncodeUrl    *bool                                                                            `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                                           `json:"label"`
		OpenInNewTab *bool                                                                            `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsType    `json:"type"`
		Url          string                                                                           `json:"url"` //nolint:revive
	},
	prior []models.DiscoverSessionPanelDrilldown,
) []models.DiscoverSessionPanelDrilldown {
	return panelkit.MapDiscoverSessionDrilldownsFromAPI(api, func(d struct {
		EncodeUrl    *bool                                                                            `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                                           `json:"label"`
		OpenInNewTab *bool                                                                            `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsType    `json:"type"`
		Url          string                                                                           `json:"url"` //nolint:revive
	}) panelkit.URLDrilldownAPIItemData {
		return panelkit.URLDrilldownAPIItemData{URL: d.Url, Label: d.Label, EncodeUrl: d.EncodeUrl, OpenInNewTab: d.OpenInNewTab}
	}, prior)
}

func discoverSessionMergeConfig0FromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionPanelConfigModel,
	tfPanel *models.PanelModel,
	cfg0 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0,
) diag.Diagnostics {
	var diags diag.Diagnostics
	prior := tfPanel.DiscoverSessionConfig
	if prior == nil || prior.ByValue == nil {
		return diags
	}

	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(cfg0.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(cfg0.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(cfg0.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(cfg0.HideBorder)
	}

	existing.Drilldowns = readDiscoverSessionDrilldownsFromConfig0(cfg0.Drilldowns, existing.Drilldowns)

	if prior.ByValue.TimeRange == nil {
		// REQ-009 / REQ-042: practitioner omitted panel time_range — keep null even though API echoes inherited dashboard range.
		existing.ByValue.TimeRange = nil
	} else {
		existing.ByValue.TimeRange = discoverSessionMergeTimeRangeModel(prior.ByValue.TimeRange, cfg0.TimeRange)
	}

	if len(cfg0.Tabs) > 0 {
		diags.Append(discoverSessionMergeTabFromAPI(ctx, &existing.ByValue.Tab, prior.ByValue.Tab, cfg0.Tabs[0])...)
	}

	discoverSessionPreserveEnvelopeNullIntent(existing, prior)
	return diags
}

func discoverSessionPreserveEnvelopeNullIntent(existing, prior *models.DiscoverSessionPanelConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = prior.Title
	}
	if !typeutils.IsKnown(prior.Description) {
		existing.Description = prior.Description
	}
	if !typeutils.IsKnown(prior.HideTitle) {
		existing.HideTitle = prior.HideTitle
	}
	if !typeutils.IsKnown(prior.HideBorder) {
		existing.HideBorder = prior.HideBorder
	}
}

func discoverSessionMergeConfig1FromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionPanelConfigModel,
	tfPanel *models.PanelModel,
	cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1,
) diag.Diagnostics {
	var diags diag.Diagnostics
	prior := tfPanel.DiscoverSessionConfig
	if prior == nil || prior.ByReference == nil {
		return diags
	}

	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(cfg1.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(cfg1.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(cfg1.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(cfg1.HideBorder)
	}

	existing.Drilldowns = readDiscoverSessionDrilldownsFromConfig1(cfg1.Drilldowns, prior.Drilldowns)

	if prior.ByReference.TimeRange == nil {
		existing.ByReference.TimeRange = nil
	} else {
		existing.ByReference.TimeRange = discoverSessionMergeTimeRangeModel(prior.ByReference.TimeRange, cfg1.TimeRange)
	}

	if typeutils.IsKnown(existing.ByReference.RefID) {
		existing.ByReference.RefID = types.StringValue(cfg1.RefId)
	}

	switch {
	case typeutils.IsKnown(prior.ByReference.SelectedTabID) && !prior.ByReference.SelectedTabID.IsNull():
		existing.ByReference.SelectedTabID = prior.ByReference.SelectedTabID
	case cfg1.SelectedTabId != nil:
		existing.ByReference.SelectedTabID = types.StringValue(*cfg1.SelectedTabId)
	default:
		existing.ByReference.SelectedTabID = types.StringNull()
	}

	if existing.ByReference.Overrides != nil && cfg1.Overrides != nil {
		diags.Append(discoverSessionMergeOverridesFromAPI(ctx, existing.ByReference.Overrides, prior.ByReference.Overrides, *cfg1.Overrides)...)
	}
	if prior.ByReference.Overrides == nil {
		existing.ByReference.Overrides = nil
	} else if existing.ByReference.Overrides != nil {
		discoverSessionPreserveOverridesNullIntent(existing.ByReference.Overrides, prior.ByReference.Overrides)
	}

	discoverSessionPreserveEnvelopeNullIntent(existing, prior)
	return diags
}

func discoverSessionMergeTimeRangeModel(prior *models.TimeRangeModel, api *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema) *models.TimeRangeModel {
	if api == nil {
		return nil
	}
	tr := &models.TimeRangeModel{
		From: types.StringValue(api.From),
		To:   types.StringValue(api.To),
	}
	switch {
	case api.Mode != nil:
		tr.Mode = types.StringValue(string(*api.Mode))
	case prior != nil && typeutils.IsKnown(prior.Mode):
		tr.Mode = prior.Mode
	default:
		tr.Mode = types.StringNull()
	}
	return tr
}

func discoverSessionMergeTabFromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionTabModel,
	prior models.DiscoverSessionTabModel,
	tab kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item,
) diag.Diagnostics {
	if dsl, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil && existing.DSL != nil && prior.DSL != nil {
		return discoverSessionMergeDSLTabFromAPI(ctx, existing.DSL, *prior.DSL, dsl)
	}
	if esql, err := tab.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil && existing.ESQL != nil && prior.ESQL != nil {
		return discoverSessionMergeESQLTabFromAPI(ctx, existing.ESQL, *prior.ESQL, esql)
	}
	return nil
}

func discoverSessionMergeDSLTabFromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionDSLTabModel,
	prior models.DiscoverSessionDSLTabModel,
	api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if typeutils.IsKnown(prior.ColumnOrder) && api.ColumnOrder != nil {
		existing.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	if typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)
	}

	if len(prior.Sort) > 0 && api.Sort != nil && len(*api.Sort) > 0 {
		existing.Sort = discoverSessionSortSliceFromAPI0(*api.Sort)
	}

	if typeutils.IsKnown(prior.Density) {
		if api.Density != nil {
			existing.Density = types.StringValue(string(*api.Density))
		}
	}

	if typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = discoverSessionDSLHeaderRowHeightFromAPI(api.HeaderRowHeight)
	}
	if typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = discoverSessionDSLRowHeightFromAPI(api.RowHeight)
	}

	if typeutils.IsKnown(prior.RowsPerPage) && api.RowsPerPage != nil {
		existing.RowsPerPage = types.Int64Value(int64(*api.RowsPerPage))
	}
	if typeutils.IsKnown(prior.SampleSize) && api.SampleSize != nil {
		existing.SampleSize = types.Int64Value(int64(*api.SampleSize))
	}

	if typeutils.IsKnown(prior.ViewMode) {
		if api.ViewMode != nil {
			existing.ViewMode = types.StringValue(string(*api.ViewMode))
		}
	}

	if prior.Query != nil && typeutils.IsKnown(prior.Query.Expression) {
		q := discoverSessionQueryFromKbnAsCode(api.Query)
		existing.Query = &q
	}

	if typeutils.IsKnown(prior.DataSourceJSON) {
		dsBytes, err := json.Marshal(api.DataSource)
		if err == nil {
			existing.DataSourceJSON = jsontypes.NewNormalizedValue(string(dsBytes))
		}
	}

	if len(prior.Filters) > 0 && api.Filters != nil {
		filters := make([]models.ChartFilterJSONModel, 0, len(*api.Filters))
		for _, item := range *api.Filters {
			fm := models.ChartFilterJSONModel{}
			diags.Append(lenscommon.ChartFilterJSONPopulateFromAPIItem(&fm, item)...)
			filters = append(filters, fm)
		}
		existing.Filters = filters
	}

	discoverSessionPreserveDSLTabNullIntent(existing, prior)
	return diags
}

func discoverSessionPreserveDSLTabNullIntent(existing *models.DiscoverSessionDSLTabModel, prior models.DiscoverSessionDSLTabModel) {
	if !typeutils.IsKnown(prior.Density) {
		existing.Density = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RowsPerPage) {
		existing.RowsPerPage = types.Int64Null()
	}
	if !typeutils.IsKnown(prior.SampleSize) {
		existing.SampleSize = types.Int64Null()
	}
	if !typeutils.IsKnown(prior.ViewMode) {
		existing.ViewMode = types.StringNull()
	}
	if !typeutils.IsKnown(prior.ColumnOrder) {
		existing.ColumnOrder = prior.ColumnOrder
	}
	if !typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = prior.ColumnSettings
	}
	if prior.Sort == nil {
		existing.Sort = nil
	}
	if len(prior.Filters) == 0 {
		existing.Filters = nil
	}
}

func discoverSessionMergeESQLTabFromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionESQLTabModel,
	prior models.DiscoverSessionESQLTabModel,
	api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if typeutils.IsKnown(prior.ColumnOrder) && api.ColumnOrder != nil {
		existing.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	if typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)
	}

	if len(prior.Sort) > 0 && api.Sort != nil && len(*api.Sort) > 0 {
		existing.Sort = discoverSessionSortSliceFromAPI1(*api.Sort)
	}

	if typeutils.IsKnown(prior.Density) {
		if api.Density != nil {
			existing.Density = types.StringValue(string(*api.Density))
		}
	}

	if typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = discoverSessionESQLHeaderRowHeightFromAPI(api.HeaderRowHeight)
	}
	if typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = discoverSessionESQLRowHeightFromAPI(api.RowHeight)
	}

	if typeutils.IsKnown(prior.DataSourceJSON) {
		dsBytes, err := json.Marshal(api.DataSource)
		if err == nil {
			existing.DataSourceJSON = jsontypes.NewNormalizedValue(string(dsBytes))
		}
	}

	discoverSessionPreserveESQLTabNullIntent(existing, prior)
	return diags
}

func discoverSessionPreserveESQLTabNullIntent(existing *models.DiscoverSessionESQLTabModel, prior models.DiscoverSessionESQLTabModel) {
	if !typeutils.IsKnown(prior.Density) {
		existing.Density = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.ColumnOrder) {
		existing.ColumnOrder = prior.ColumnOrder
	}
	if !typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = prior.ColumnSettings
	}
	if prior.Sort == nil {
		existing.Sort = nil
	}
}

func discoverSessionMergeOverridesFromAPI(ctx context.Context, existing *models.DiscoverSessionOverridesModel, prior *models.DiscoverSessionOverridesModel, api struct {
	ColumnOrder    *[]string `json:"column_order,omitempty"`
	ColumnSettings *map[string]struct {
		Width *float32 `json:"width,omitempty"`
	} `json:"column_settings,omitempty"`
	Density         *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
	HeaderRowHeight *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
	RowHeight       *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
	RowsPerPage     *float32                                                                                     `json:"rows_per_page,omitempty"`
	SampleSize      *float32                                                                                     `json:"sample_size,omitempty"`
	Sort            *[]struct {
		Direction kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                                `json:"name"`
	} `json:"sort,omitempty"`
}) diag.Diagnostics {
	var diags diag.Diagnostics

	if prior != nil && typeutils.IsKnown(prior.ColumnOrder) && api.ColumnOrder != nil {
		existing.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	if prior != nil && typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)
	}

	if api.Sort != nil && prior != nil && len(prior.Sort) > 0 {
		out := make([]models.DiscoverSessionSortModel, len(*api.Sort))
		for i, s := range *api.Sort {
			out[i].Name = types.StringValue(s.Name)
			out[i].Direction = types.StringValue(string(s.Direction))
		}
		existing.Sort = out
	}

	if prior != nil && typeutils.IsKnown(prior.Density) {
		if api.Density != nil {
			existing.Density = types.StringValue(string(*api.Density))
		}
	}

	if prior != nil && typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = discoverSessionOverridesHeaderRowHeightFromAPI(api.HeaderRowHeight)
	}
	if prior != nil && typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = discoverSessionOverridesRowHeightFromAPI(api.RowHeight)
	}

	if prior != nil && typeutils.IsKnown(prior.RowsPerPage) && api.RowsPerPage != nil {
		existing.RowsPerPage = types.Int64Value(int64(*api.RowsPerPage))
	}
	if prior != nil && typeutils.IsKnown(prior.SampleSize) && api.SampleSize != nil {
		existing.SampleSize = types.Int64Value(int64(*api.SampleSize))
	}

	discoverSessionPreserveOverridesNullIntent(existing, prior)
	return diags
}

func discoverSessionPreserveOverridesNullIntent(existing, prior *models.DiscoverSessionOverridesModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.ColumnOrder) {
		existing.ColumnOrder = prior.ColumnOrder
	}
	if !typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = prior.ColumnSettings
	}
	if prior.Sort == nil {
		existing.Sort = nil
	}
	if !typeutils.IsKnown(prior.Density) {
		existing.Density = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HeaderRowHeight) {
		existing.HeaderRowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RowHeight) {
		existing.RowHeight = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RowsPerPage) {
		existing.RowsPerPage = types.Int64Null()
	}
	if !typeutils.IsKnown(prior.SampleSize) {
		existing.SampleSize = types.Int64Null()
	}
}
