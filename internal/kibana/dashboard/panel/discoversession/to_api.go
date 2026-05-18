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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
			fd := lenscommon.DecodeChartFilterJSON(f.FilterJSON, &item)
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
	case valueAuto:
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
	case valueAuto:
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
	case valueAuto:
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
	case valueAuto:
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
	case valueAuto:
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
	case valueAuto:
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
