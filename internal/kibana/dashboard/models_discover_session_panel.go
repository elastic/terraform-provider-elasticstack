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
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func discoverSessionColumnSettingObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"width": types.Float64Type,
		},
	}
}

func discoverSessionPanelToAPI(ctx context.Context, pm models.PanelModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, panelID *string, dashTR *models.TimeRangeModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.DiscoverSessionConfig
	if cfg == nil {
		diags.AddError("Missing discover_session panel configuration", "Discover session panels require `discover_session_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	out := kbapi.KbnDashboardPanelTypeDiscoverSession{
		Grid: grid,
		Id:   panelID,
		Type: kbapi.DiscoverSession,
	}

	var apiCfg kbapi.KbnDashboardPanelTypeDiscoverSession_Config
	switch {
	case cfg.ByValue != nil:
		built, d := discoverSessionByValueToAPI(ctx, cfg, dashTR)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := apiCfg.FromKbnDashboardPanelTypeDiscoverSessionConfig0(built); err != nil {
			diags.AddError("Failed to build discover_session config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	case cfg.ByReference != nil:
		built, d := discoverSessionByReferenceToAPI(ctx, cfg, dashTR)
		diags.Append(d...)
		if diags.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		if err := apiCfg.FromKbnDashboardPanelTypeDiscoverSessionConfig1(built); err != nil {
			diags.AddError("Failed to build discover_session config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	default:
		diags.AddError("Invalid discover_session_config", "Exactly one of `by_value` or `by_reference` must be set.")
		return kbapi.DashboardPanelItem{}, diags
	}

	out.Config = apiCfg

	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeDiscoverSession(out); err != nil {
		diags.AddError("Failed to create discover_session panel", err.Error())
	}
	return panelItem, diags
}

func discoverSessionByValueToAPI(
	ctx context.Context,
	cfg *models.DiscoverSessionPanelConfigModel,
	dashTR *models.TimeRangeModel,
) (kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0{}

	tr, d := discoverSessionResolveTimeRange(cfg.ByValue.TimeRange, dashTR)
	diags.Append(d...)
	if diags.HasError() {
		return api, diags
	}
	api.TimeRange = tr

	discoverSessionApplyEnvelopeToConfig0(cfg, &api)
	tabItem, d := discoverSessionTabToAPI(ctx, cfg.ByValue.Tab)
	diags.Append(d...)
	if diags.HasError() {
		return api, diags
	}
	api.Tabs = []kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item{tabItem}
	return api, diags
}

func discoverSessionByReferenceToAPI(
	ctx context.Context,
	cfg *models.DiscoverSessionPanelConfigModel,
	dashTR *models.TimeRangeModel,
) (kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1{}

	tr, d := discoverSessionResolveTimeRange(cfg.ByReference.TimeRange, dashTR)
	diags.Append(d...)
	if diags.HasError() {
		return api, diags
	}
	api.TimeRange = tr

	api.RefId = cfg.ByReference.RefID.ValueString()

	if typeutils.IsKnown(cfg.ByReference.SelectedTabID) && !cfg.ByReference.SelectedTabID.IsNull() {
		s := cfg.ByReference.SelectedTabID.ValueString()
		api.SelectedTabId = &s
	}

	if cfg.ByReference.Overrides != nil {
		o, d := discoverSessionOverridesToAPI(ctx, *cfg.ByReference.Overrides)
		diags.Append(d...)
		if diags.HasError() {
			return api, diags
		}
		api.Overrides = &o
	}

	discoverSessionApplyEnvelopeToConfig1(cfg, &api)
	return api, diags
}

func discoverSessionApplyEnvelopeToConfig0(cfg *models.DiscoverSessionPanelConfigModel, api *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0) {
	if typeutils.IsKnown(cfg.Title) {
		api.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		api.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		api.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		api.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}
	if len(cfg.Drilldowns) > 0 {
		dd := make([]struct {
			EncodeUrl    *bool                                                              `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                                             `json:"label"`
			OpenInNewTab *bool                                                              `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTrigger `json:"trigger"`
			Type         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsType    `json:"type"`
			Url          string                                                             `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))
		for i, x := range cfg.Drilldowns {
			dd[i].Url = x.URL.ValueString()
			dd[i].Label = x.Label.ValueString()
			dd[i].Trigger = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTriggerOnOpenPanelMenu
			dd[i].Type = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTypeUrlDrilldown
			if typeutils.IsKnown(x.EncodeURL) {
				dd[i].EncodeUrl = x.EncodeURL.ValueBoolPointer()
			}
			if typeutils.IsKnown(x.OpenInNewTab) {
				dd[i].OpenInNewTab = x.OpenInNewTab.ValueBoolPointer()
			}
		}
		api.Drilldowns = &dd
	}
}

func discoverSessionApplyEnvelopeToConfig1(cfg *models.DiscoverSessionPanelConfigModel, api *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1) {
	if typeutils.IsKnown(cfg.Title) {
		api.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		api.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		api.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		api.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}
	if len(cfg.Drilldowns) > 0 {
		dd := make([]struct {
			EncodeUrl    *bool                                                              `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                                             `json:"label"`
			OpenInNewTab *bool                                                              `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTrigger `json:"trigger"`
			Type         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsType    `json:"type"`
			Url          string                                                             `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))
		for i, x := range cfg.Drilldowns {
			dd[i].Url = x.URL.ValueString()
			dd[i].Label = x.Label.ValueString()
			dd[i].Trigger = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTriggerOnOpenPanelMenu
			dd[i].Type = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTypeUrlDrilldown
			if typeutils.IsKnown(x.EncodeURL) {
				dd[i].EncodeUrl = x.EncodeURL.ValueBoolPointer()
			}
			if typeutils.IsKnown(x.OpenInNewTab) {
				dd[i].OpenInNewTab = x.OpenInNewTab.ValueBoolPointer()
			}
		}
		api.Drilldowns = &dd
	}
}

func discoverSessionResolveTimeRange(panelTR *models.TimeRangeModel, dashTR *models.TimeRangeModel) (kbapi.KbnEsQueryServerTimeRangeSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	source := panelTR
	if source == nil {
		source = dashTR
	}
	if source == nil {
		diags.AddError(
			"Missing time range for discover_session panel",
			"Set `discover_session_config.by_value.time_range` or `discover_session_config.by_reference.time_range`, or configure the dashboard root `time_range` so the panel can inherit it.",
		)
		return kbapi.KbnEsQueryServerTimeRangeSchema{}, diags
	}
	out := kbapi.KbnEsQueryServerTimeRangeSchema{
		From: source.From.ValueString(),
		To:   source.To.ValueString(),
	}
	if typeutils.IsKnown(source.Mode) {
		m := kbapi.KbnEsQueryServerTimeRangeSchemaMode(source.Mode.ValueString())
		out.Mode = &m
	}
	return out, diags
}

func discoverSessionTabToAPI(ctx context.Context, tab models.DiscoverSessionTabModel) (kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	var item kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item
	switch {
	case tab.DSL != nil:
		dsl, d := discoverSessionDSLTabToAPI(ctx, *tab.DSL)
		diags.Append(d...)
		if diags.HasError() {
			return item, diags
		}
		if err := item.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(dsl); err != nil {
			diags.AddError("Failed to marshal discover_session dsl tab", err.Error())
		}
	case tab.ESQL != nil:
		esql, d := discoverSessionESQLTabToAPI(ctx, *tab.ESQL)
		diags.Append(d...)
		if diags.HasError() {
			return item, diags
		}
		if err := item.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(esql); err != nil {
			diags.AddError("Failed to marshal discover_session esql tab", err.Error())
		}
	default:
		diags.AddError("Invalid discover_session tab", "Exactly one of `dsl` or `esql` must be set.")
	}
	return item, diags
}

func discoverSessionDSLTabToAPI(ctx context.Context, m models.DiscoverSessionDSLTabModel) (kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0{}

	if typeutils.IsKnown(m.ColumnOrder) && !m.ColumnOrder.IsNull() {
		co := typeutils.ListTypeToSliceString(ctx, m.ColumnOrder, path.Empty(), &diags)
		if len(co) > 0 {
			api.ColumnOrder = &co
		}
	}

	api.ColumnSettings = discoverSessionColumnSettingsToAPI(ctx, m.ColumnSettings, &diags)

	if len(m.Sort) > 0 {
		s := discoverSessionSortToAPI0(m.Sort)
		api.Sort = &s
	}

	if typeutils.IsKnown(m.Density) {
		d := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0Density(m.Density.ValueString())
		api.Density = &d
	}

	hdr, dHdr := discoverSessionDSLHeaderRowHeightToAPI(m.HeaderRowHeight)
	diags.Append(dHdr...)
	if hdr != nil {
		api.HeaderRowHeight = hdr
	}

	rh, dRh := discoverSessionDSLRowHeightToAPI(m.RowHeight)
	diags.Append(dRh...)
	if rh != nil {
		api.RowHeight = rh
	}

	if typeutils.IsKnown(m.RowsPerPage) {
		v := float32(m.RowsPerPage.ValueInt64())
		api.RowsPerPage = &v
	}
	if typeutils.IsKnown(m.SampleSize) {
		v := float32(m.SampleSize.ValueInt64())
		api.SampleSize = &v
	}

	if typeutils.IsKnown(m.ViewMode) {
		vm := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0ViewMode(m.ViewMode.ValueString())
		api.ViewMode = &vm
	}

	if m.Query != nil {
		api.Query = discoverSessionQueryToKbnAsCode(*m.Query)
	}

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal discover_session tab.dsl.data_source_json", err.Error())
			return api, diags
		}
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_Filters_Item, 0, len(m.Filters))
		for _, f := range m.Filters {
			var item kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_Filters_Item
			fd := decodeChartFilterJSON(f.FilterJSON, &item)
			diags.Append(fd...)
			if fd.HasError() {
				return api, diags
			}
			filters = append(filters, item)
		}
		api.Filters = &filters
	}

	return api, diags
}

func discoverSessionESQLTabToAPI(ctx context.Context, m models.DiscoverSessionESQLTabModel) (kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1{}

	if typeutils.IsKnown(m.ColumnOrder) && !m.ColumnOrder.IsNull() {
		co := typeutils.ListTypeToSliceString(ctx, m.ColumnOrder, path.Empty(), &diags)
		if len(co) > 0 {
			api.ColumnOrder = &co
		}
	}

	api.ColumnSettings = discoverSessionColumnSettingsToAPI(ctx, m.ColumnSettings, &diags)

	if len(m.Sort) > 0 {
		s := discoverSessionSortToAPI1(m.Sort)
		api.Sort = &s
	}

	if typeutils.IsKnown(m.Density) {
		d := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1Density(m.Density.ValueString())
		api.Density = &d
	}

	hdr, dHdr := discoverSessionESQLHeaderRowHeightToAPI(m.HeaderRowHeight)
	diags.Append(dHdr...)
	if hdr != nil {
		api.HeaderRowHeight = hdr
	}

	rh, dRh := discoverSessionESQLRowHeightToAPI(m.RowHeight)
	diags.Append(dRh...)
	if rh != nil {
		api.RowHeight = rh
	}

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal discover_session tab.esql.data_source_json", err.Error())
			return api, diags
		}
	}

	return api, diags
}

func discoverSessionOverridesToAPI(ctx context.Context, m models.DiscoverSessionOverridesModel) (struct {
	ColumnOrder    *[]string `json:"column_order,omitempty"`
	ColumnSettings *map[string]struct {
		Width *float32 `json:"width,omitempty"`
	} `json:"column_settings,omitempty"`
	Density         *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
	HeaderRowHeight *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
	RowHeight       *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
	RowsPerPage     *float32                                                                       `json:"rows_per_page,omitempty"`
	SampleSize      *float32                                                                       `json:"sample_size,omitempty"`
	Sort            *[]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                  `json:"name"`
	} `json:"sort,omitempty"`
}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api struct {
		ColumnOrder    *[]string `json:"column_order,omitempty"`
		ColumnSettings *map[string]struct {
			Width *float32 `json:"width,omitempty"`
		} `json:"column_settings,omitempty"`
		Density         *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
		HeaderRowHeight *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
		RowHeight       *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
		RowsPerPage     *float32                                                                       `json:"rows_per_page,omitempty"`
		SampleSize      *float32                                                                       `json:"sample_size,omitempty"`
		Sort            *[]struct {
			Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
			Name      string                                                                  `json:"name"`
		} `json:"sort,omitempty"`
	}

	if typeutils.IsKnown(m.ColumnOrder) && !m.ColumnOrder.IsNull() {
		co := typeutils.ListTypeToSliceString(ctx, m.ColumnOrder, path.Empty(), &diags)
		if len(co) > 0 {
			api.ColumnOrder = &co
		}
	}

	api.ColumnSettings = discoverSessionColumnSettingsToAPI(ctx, m.ColumnSettings, &diags)

	if len(m.Sort) > 0 {
		s := discoverSessionOverridesSortToAPI(m.Sort)
		api.Sort = &s
	}

	if typeutils.IsKnown(m.Density) {
		d := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity(m.Density.ValueString())
		api.Density = &d
	}

	hdr, dHdr := discoverSessionOverridesHeaderRowHeightToAPI(m.HeaderRowHeight)
	diags.Append(dHdr...)
	if hdr != nil {
		api.HeaderRowHeight = hdr
	}

	rh, dRh := discoverSessionOverridesRowHeightToAPI(m.RowHeight)
	diags.Append(dRh...)
	if rh != nil {
		api.RowHeight = rh
	}

	if typeutils.IsKnown(m.RowsPerPage) {
		v := float32(m.RowsPerPage.ValueInt64())
		api.RowsPerPage = &v
	}
	if typeutils.IsKnown(m.SampleSize) {
		v := float32(m.SampleSize.ValueInt64())
		api.SampleSize = &v
	}

	return api, diags
}

func discoverSessionSortToAPI0(sort []models.DiscoverSessionSortModel) []struct {
	Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0SortDirection `json:"direction"`
	Name      string                                                              `json:"name"`
} {
	out := make([]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0SortDirection `json:"direction"`
		Name      string                                                              `json:"name"`
	}, len(sort))
	for i, s := range sort {
		out[i].Name = s.Name.ValueString()
		out[i].Direction = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0SortDirection(s.Direction.ValueString())
	}
	return out
}

func discoverSessionSortToAPI1(sort []models.DiscoverSessionSortModel) []struct {
	Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1SortDirection `json:"direction"`
	Name      string                                                              `json:"name"`
} {
	out := make([]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1SortDirection `json:"direction"`
		Name      string                                                              `json:"name"`
	}, len(sort))
	for i, s := range sort {
		out[i].Name = s.Name.ValueString()
		out[i].Direction = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1SortDirection(s.Direction.ValueString())
	}
	return out
}

func discoverSessionOverridesSortToAPI(sort []models.DiscoverSessionSortModel) []struct {
	Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
	Name      string                                                                  `json:"name"`
} {
	out := make([]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                  `json:"name"`
	}, len(sort))
	for i, s := range sort {
		out[i].Name = s.Name.ValueString()
		out[i].Direction = kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection(s.Direction.ValueString())
	}
	return out
}

func discoverSessionQueryToKbnAsCode(m models.FilterSimpleModel) kbapi.KbnAsCodeQuery {
	q := kbapi.KbnAsCodeQuery{
		Expression: m.Expression.ValueString(),
	}
	if typeutils.IsKnown(m.Language) {
		q.Language = kbapi.KbnAsCodeQueryLanguage(m.Language.ValueString())
	} else {
		q.Language = kbapi.Kql
	}
	return q
}

func discoverSessionColumnSettingsToAPI(ctx context.Context, m types.Map, diags *diag.Diagnostics) *map[string]struct {
	Width *float32 `json:"width,omitempty"`
} {
	if !typeutils.IsKnown(m) || m.IsNull() {
		return nil
	}
	raw := typeutils.MapTypeAs[models.DiscoverSessionColumnSettingModel](ctx, m, path.Empty(), diags)
	if diags.HasError() {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]struct {
		Width *float32 `json:"width,omitempty"`
	}, len(raw))
	for k, v := range raw {
		if typeutils.IsKnown(v.Width) {
			f := float32(v.Width.ValueFloat64())
			out[k] = struct {
				Width *float32 `json:"width,omitempty"`
			}{Width: &f}
		} else {
			out[k] = struct {
				Width *float32 `json:"width,omitempty"`
			}{}
		}
	}
	return &out
}

func discoverSessionDSLHeaderRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_HeaderRowHeight, diag.Diagnostics) {
	return discoverSessionNumericOrAutoUnionToDSLHeaderHeight(s)
}

func discoverSessionDSLRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_RowHeight, diag.Diagnostics) {
	return discoverSessionNumericOrAutoUnionToDSLRowHeight(s)
}

func discoverSessionESQLHeaderRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_HeaderRowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_HeaderRowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight1Auto); err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight0(f); err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
		}
	}
	return &out, diags
}

func discoverSessionESQLRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_RowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_RowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight1Auto); err != nil {
			diags.AddError("Invalid row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight0(f); err != nil {
			diags.AddError("Invalid row_height", err.Error())
		}
	}
	return &out, diags
}

func discoverSessionOverridesHeaderRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight1Auto); err != nil {
			diags.AddError("Invalid overrides.header_row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid overrides.header_row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight0(f); err != nil {
			diags.AddError("Invalid overrides.header_row_height", err.Error())
		}
	}
	return &out, diags
}

func discoverSessionOverridesRowHeightToAPI(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight1Auto); err != nil {
			diags.AddError("Invalid overrides.row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid overrides.row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight0(f); err != nil {
			diags.AddError("Invalid overrides.row_height", err.Error())
		}
	}
	return &out, diags
}

func discoverSessionNumericOrAutoUnionToDSLHeaderHeight(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_HeaderRowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_HeaderRowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight1Auto); err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight0(f); err != nil {
			diags.AddError("Invalid header_row_height", err.Error())
		}
	}
	return &out, diags
}

func discoverSessionNumericOrAutoUnionToDSLRowHeight(s types.String) (*kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_RowHeight, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(s) || s.IsNull() {
		return nil, diags
	}
	v := s.ValueString()
	var out kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_RowHeight
	switch v {
	case dashboardValueAuto:
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight1(kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight1Auto); err != nil {
			diags.AddError("Invalid row_height", err.Error())
		}
	default:
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddError("Invalid row_height", err.Error())
			return nil, diags
		}
		f := float32(n)
		if err := out.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight0(f); err != nil {
			diags.AddError("Invalid row_height", err.Error())
		}
	}
	return &out, diags
}

// --- Read path ---

// discoverSessionAPIConfigLooksByReference distinguishes inline vs linked Discover session configs.
// kbapi's Config union unmarshals successfully into both generated structs, so we key off JSON `ref_id`
// (present only on the by-reference branch).
func discoverSessionAPIConfigLooksByReference(apiCfg kbapi.KbnDashboardPanelTypeDiscoverSession_Config) bool {
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

func populateDiscoverSessionPanelFromAPI(ctx context.Context, pm *models.PanelModel, tfPanel *models.PanelModel, apiPanel kbapi.KbnDashboardPanelTypeDiscoverSession) {
	if tfPanel == nil {
		pm.DiscoverSessionConfig = discoverSessionPanelConfigFromAPIImport(ctx, apiPanel)
		return
	}

	existing := pm.DiscoverSessionConfig
	if existing == nil {
		return
	}

	prior := tfPanel.DiscoverSessionConfig
	apiByRef := discoverSessionAPIConfigLooksByReference(apiPanel.Config)

	if discoverSessionPriorTFBranchMismatchesAPI(apiByRef, prior) {
		// Drift import: replace typed config from API so the next plan surfaces the branch change.
		if apiByRef {
			cfg1, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig1()
			if err == nil {
				if imported := discoverSessionConfig1FromAPIImport(ctx, cfg1); imported != nil {
					*existing = *imported
				}
			}
			return
		}
		cfg0, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig0()
		if err == nil {
			if imported := discoverSessionConfig0FromAPIImport(ctx, cfg0); imported != nil {
				*existing = *imported
			}
		}
		return
	}

	if apiByRef {
		cfg1, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig1()
		if err == nil {
			discoverSessionMergeConfig1FromAPI(ctx, existing, tfPanel, cfg1)
		}
		return
	}

	cfg0, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig0()
	if err == nil {
		discoverSessionMergeConfig0FromAPI(ctx, existing, tfPanel, cfg0)
	}
}

func discoverSessionPanelConfigFromAPIImport(ctx context.Context, apiPanel kbapi.KbnDashboardPanelTypeDiscoverSession) *models.DiscoverSessionPanelConfigModel {
	if discoverSessionAPIConfigLooksByReference(apiPanel.Config) {
		cfg1, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig1()
		if err == nil {
			return discoverSessionConfig1FromAPIImport(ctx, cfg1)
		}
		return nil
	}

	cfg0, err := apiPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig0()
	if err == nil {
		return discoverSessionConfig0FromAPIImport(ctx, cfg0)
	}
	return nil
}

func discoverSessionConfig0FromAPIImport(ctx context.Context, cfg0 kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0) *models.DiscoverSessionPanelConfigModel {
	cfg := &models.DiscoverSessionPanelConfigModel{
		Title:       types.StringPointerValue(cfg0.Title),
		Description: types.StringPointerValue(cfg0.Description),
		HideTitle:   types.BoolPointerValue(cfg0.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg0.HideBorder),
		ByValue: &models.DiscoverSessionPanelByValueModel{
			TimeRange: discoverSessionTimeRangePtrFromAPI(cfg0.TimeRange),
			Tab:       discoverSessionTabFromAPIConfig0(ctx, cfg0.Tabs),
		},
	}
	cfg.Drilldowns = readDiscoverSessionDrilldownsFromConfig0(cfg0.Drilldowns, nil)
	return cfg
}

func discoverSessionConfig1FromAPIImport(ctx context.Context, cfg1 kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1) *models.DiscoverSessionPanelConfigModel {
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
	if cfg1.SelectedTabId != nil {
		cfg.ByReference.SelectedTabID = types.StringValue(*cfg1.SelectedTabId)
	} else {
		cfg.ByReference.SelectedTabID = types.StringNull()
	}
	if cfg1.Overrides != nil {
		cfg.ByReference.Overrides = discoverSessionOverridesFromAPI(ctx, *cfg1.Overrides)
	}
	cfg.Drilldowns = readDiscoverSessionDrilldownsFromConfig1(cfg1.Drilldowns, nil)
	return cfg
}

func discoverSessionTimeRangePtrFromAPI(api kbapi.KbnEsQueryServerTimeRangeSchema) *models.TimeRangeModel {
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

func discoverSessionTabFromAPIConfig0(ctx context.Context, tabs []kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item) models.DiscoverSessionTabModel {
	if len(tabs) == 0 {
		return models.DiscoverSessionTabModel{}
	}
	tab := tabs[0]
	if dsl, err := tab.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil {
		return models.DiscoverSessionTabModel{DSL: discoverSessionDSLTabFromAPI(ctx, dsl)}
	}
	if esql, err := tab.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil {
		return models.DiscoverSessionTabModel{ESQL: discoverSessionESQLTabFromAPI(ctx, esql)}
	}
	return models.DiscoverSessionTabModel{}
}

func discoverSessionDSLTabFromAPI(ctx context.Context, api kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0) *models.DiscoverSessionDSLTabModel {
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
			diags.Append(chartFilterJSONPopulateFromAPIItem(&fm, item)...)
			filters = append(filters, fm)
		}
		m.Filters = filters
	}

	return m
}

func discoverSessionESQLTabFromAPI(ctx context.Context, api kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1) *models.DiscoverSessionESQLTabModel {
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

	return m
}

func discoverSessionQueryFromKbnAsCode(q kbapi.KbnAsCodeQuery) models.FilterSimpleModel {
	return models.FilterSimpleModel{
		Expression: types.StringValue(q.Expression),
		Language:   types.StringValue(string(q.Language)),
	}
}

func discoverSessionSortSliceFromAPI0(api []struct {
	Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0SortDirection `json:"direction"`
	Name      string                                                              `json:"name"`
}) []models.DiscoverSessionSortModel {
	out := make([]models.DiscoverSessionSortModel, len(api))
	for i, s := range api {
		out[i].Name = types.StringValue(s.Name)
		out[i].Direction = types.StringValue(string(s.Direction))
	}
	return out
}

func discoverSessionSortSliceFromAPI1(api []struct {
	Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1SortDirection `json:"direction"`
	Name      string                                                              `json:"name"`
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
	Density         *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
	HeaderRowHeight *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
	RowHeight       *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
	RowsPerPage     *float32                                                                       `json:"rows_per_page,omitempty"`
	SampleSize      *float32                                                                       `json:"sample_size,omitempty"`
	Sort            *[]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                  `json:"name"`
	} `json:"sort,omitempty"`
}) *models.DiscoverSessionOverridesModel {
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

	return m
}

func discoverSessionColumnSettingsFromAPI(ctx context.Context, api *map[string]struct {
	Width *float32 `json:"width,omitempty"`
}, diags *diag.Diagnostics) types.Map {
	if api == nil || len(*api) == 0 {
		return types.MapNull(discoverSessionColumnSettingObjectType())
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
	return typeutils.MapValueFrom(ctx, elems, discoverSessionColumnSettingObjectType(), path.Empty(), diags)
}

func discoverSessionDSLHeaderRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_HeaderRowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0HeaderRowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func discoverSessionDSLRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_0_RowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0RowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func discoverSessionESQLHeaderRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_HeaderRowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1HeaderRowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func discoverSessionESQLRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_1_RowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1RowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func discoverSessionOverridesHeaderRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesHeaderRowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func discoverSessionOverridesRowHeightFromAPI(h *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight) types.String {
	if h == nil {
		return types.StringNull()
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight0(); err == nil {
		return types.StringValue(strconv.FormatInt(int64(v), 10))
	}
	if v, err := h.AsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesRowHeight1(); err == nil {
		return types.StringValue(string(v))
	}
	return types.StringNull()
}

func readDiscoverSessionDrilldownsFromConfig0(
	api *[]struct {
		EncodeUrl    *bool                                                              `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                             `json:"label"`
		OpenInNewTab *bool                                                              `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsType    `json:"type"`
		Url          string                                                             `json:"url"` //nolint:revive
	},
	prior []models.DiscoverSessionPanelDrilldown,
) []models.DiscoverSessionPanelDrilldown {
	if api == nil || len(*api) == 0 {
		return nil
	}
	out := make([]models.DiscoverSessionPanelDrilldown, len(*api))
	for i, d := range *api {
		out[i].URL = types.StringValue(d.Url)
		out[i].Label = types.StringValue(d.Label)
		var p *models.DiscoverSessionPanelDrilldown
		if i < len(prior) {
			p = &prior[i]
		}
		if p == nil {
			out[i].EncodeURL = panelDrilldownBoolImportPreserving(d.EncodeUrl, drilldownURLEncodeURLDefault)
			out[i].OpenInNewTab = panelDrilldownBoolImportPreserving(d.OpenInNewTab, drilldownURLOpenInNewTabDefault)
			continue
		}
		switch {
		case p.EncodeURL.IsNull():
			out[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			out[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			out[i].EncodeURL = types.BoolNull()
		}
		switch {
		case p.OpenInNewTab.IsNull():
			out[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			out[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			out[i].OpenInNewTab = types.BoolNull()
		}
	}
	return out
}

func readDiscoverSessionDrilldownsFromConfig1(
	api *[]struct {
		EncodeUrl    *bool                                                              `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                             `json:"label"`
		OpenInNewTab *bool                                                              `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1DrilldownsType    `json:"type"`
		Url          string                                                             `json:"url"` //nolint:revive
	},
	prior []models.DiscoverSessionPanelDrilldown,
) []models.DiscoverSessionPanelDrilldown {
	if api == nil || len(*api) == 0 {
		return nil
	}
	out := make([]models.DiscoverSessionPanelDrilldown, len(*api))
	for i, d := range *api {
		out[i].URL = types.StringValue(d.Url)
		out[i].Label = types.StringValue(d.Label)
		var p *models.DiscoverSessionPanelDrilldown
		if i < len(prior) {
			p = &prior[i]
		}
		if p == nil {
			out[i].EncodeURL = panelDrilldownBoolImportPreserving(d.EncodeUrl, drilldownURLEncodeURLDefault)
			out[i].OpenInNewTab = panelDrilldownBoolImportPreserving(d.OpenInNewTab, drilldownURLOpenInNewTabDefault)
			continue
		}
		switch {
		case p.EncodeURL.IsNull():
			out[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			out[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			out[i].EncodeURL = types.BoolNull()
		}
		switch {
		case p.OpenInNewTab.IsNull():
			out[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			out[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			out[i].OpenInNewTab = types.BoolNull()
		}
	}
	return out
}

func discoverSessionMergeConfig0FromAPI(ctx context.Context, existing *models.DiscoverSessionPanelConfigModel, tfPanel *models.PanelModel, cfg0 kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0) {
	prior := tfPanel.DiscoverSessionConfig
	if prior == nil || prior.ByValue == nil {
		return
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
		discoverSessionMergeTabFromAPI(ctx, &existing.ByValue.Tab, prior.ByValue.Tab, cfg0.Tabs[0])
	}
}

func discoverSessionMergeConfig1FromAPI(ctx context.Context, existing *models.DiscoverSessionPanelConfigModel, tfPanel *models.PanelModel, cfg1 kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1) {
	prior := tfPanel.DiscoverSessionConfig
	if prior == nil || prior.ByReference == nil {
		return
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

	existing.Drilldowns = readDiscoverSessionDrilldownsFromConfig1(cfg1.Drilldowns, existing.Drilldowns)

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
		discoverSessionMergeOverridesFromAPI(ctx, existing.ByReference.Overrides, prior.ByReference.Overrides, *cfg1.Overrides)
	}
}

func discoverSessionMergeTimeRangeModel(prior *models.TimeRangeModel, api kbapi.KbnEsQueryServerTimeRangeSchema) *models.TimeRangeModel {
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
	tab kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item,
) {
	if dsl, err := tab.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(); err == nil && existing.DSL != nil && prior.DSL != nil {
		discoverSessionMergeDSLTabFromAPI(ctx, existing.DSL, *prior.DSL, dsl)
		return
	}
	if esql, err := tab.AsKbnDashboardPanelTypeDiscoverSessionConfig0Tabs1(); err == nil && existing.ESQL != nil && prior.ESQL != nil {
		discoverSessionMergeESQLTabFromAPI(ctx, existing.ESQL, *prior.ESQL, esql)
	}
}

func discoverSessionMergeDSLTabFromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionDSLTabModel,
	prior models.DiscoverSessionDSLTabModel,
	api kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0,
) {
	var diags diag.Diagnostics

	if typeutils.IsKnown(prior.ColumnOrder) && api.ColumnOrder != nil {
		existing.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	if typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)
	}

	if api.Sort != nil && len(*api.Sort) > 0 {
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
			diags.Append(chartFilterJSONPopulateFromAPIItem(&fm, item)...)
			filters = append(filters, fm)
		}
		existing.Filters = filters
	}
}

func discoverSessionMergeESQLTabFromAPI(
	ctx context.Context,
	existing *models.DiscoverSessionESQLTabModel,
	prior models.DiscoverSessionESQLTabModel,
	api kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs1,
) {
	var diags diag.Diagnostics

	if typeutils.IsKnown(prior.ColumnOrder) && api.ColumnOrder != nil {
		existing.ColumnOrder = typeutils.SliceToListTypeString(ctx, *api.ColumnOrder, path.Empty(), &diags)
	}

	if typeutils.IsKnown(prior.ColumnSettings) {
		existing.ColumnSettings = discoverSessionColumnSettingsFromAPI(ctx, api.ColumnSettings, &diags)
	}

	if api.Sort != nil && len(*api.Sort) > 0 {
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
}

func discoverSessionMergeOverridesFromAPI(ctx context.Context, existing *models.DiscoverSessionOverridesModel, prior *models.DiscoverSessionOverridesModel, api struct {
	ColumnOrder    *[]string `json:"column_order,omitempty"`
	ColumnSettings *map[string]struct {
		Width *float32 `json:"width,omitempty"`
	} `json:"column_settings,omitempty"`
	Density         *kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity             `json:"density,omitempty"`
	HeaderRowHeight *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight `json:"header_row_height,omitempty"`
	RowHeight       *kbapi.KbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight       `json:"row_height,omitempty"`
	RowsPerPage     *float32                                                                       `json:"rows_per_page,omitempty"`
	SampleSize      *float32                                                                       `json:"sample_size,omitempty"`
	Sort            *[]struct {
		Direction kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                  `json:"name"`
	} `json:"sort,omitempty"`
}) {
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
}
