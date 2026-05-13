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
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// jsonNullString is the JSON encoding of null; json.Marshal uses it for unset union/API fields.
const jsonNullString = "null"

// dashboardModel is the top-level Terraform model
type dashboardModel struct {
	ID               types.String          `tfsdk:"id"`
	KibanaConnection types.List            `tfsdk:"kibana_connection"`
	SpaceID          types.String          `tfsdk:"space_id"`
	DashboardID      types.String          `tfsdk:"dashboard_id"`
	Title            types.String          `tfsdk:"title"`
	Description      types.String          `tfsdk:"description"`
	TimeRange        *timeRangeModel       `tfsdk:"time_range"`
	RefreshInterval  *refreshIntervalModel `tfsdk:"refresh_interval"`
	Query            *dashboardQueryModel  `tfsdk:"query"`
	Filters          types.List            `tfsdk:"filters"`
	Tags             types.List            `tfsdk:"tags"`
	Options          *optionsModel         `tfsdk:"options"`
	AccessControl    *AccessControlValue   `tfsdk:"access_control"`
	Panels           []panelModel          `tfsdk:"panels"`
	PinnedPanels     []pinnedPanelModel    `tfsdk:"pinned_panels"`
	Sections         []sectionModel        `tfsdk:"sections"`
}

// populateFromAPI populates the Terraform model from the API response
func (m *dashboardModel) populateFromAPI(ctx context.Context, resp *kbapi.GetDashboardsIdResponse, dashboardID string, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics
	priorPinnedPanels := m.PinnedPanels
	data := resp.JSON200

	// Set composite ID
	resourceID := clients.CompositeID{ClusterID: spaceID, ResourceID: dashboardID}
	m.ID = types.StringValue(resourceID.String())
	m.DashboardID = types.StringValue(dashboardID)
	m.SpaceID = types.StringValue(spaceID)

	// Map the dashboard data fields
	m.Title = types.StringValue(data.Data.Title)

	if data.Data.Description != nil {
		m.Description = types.StringValue(*data.Data.Description)
	} else {
		m.Description = types.StringNull()
	}

	// Map time range (preserve prior time_range.mode when GET omits it; see REQ-009)
	var preservedMode types.String
	if m.TimeRange != nil {
		preservedMode = m.TimeRange.Mode
	}
	m.TimeRange = &timeRangeModel{
		From: types.StringValue(data.Data.TimeRange.From),
		To:   types.StringValue(data.Data.TimeRange.To),
		Mode: preservedMode,
	}

	// Map refresh interval
	m.RefreshInterval = &refreshIntervalModel{
		Pause: types.BoolValue(data.Data.RefreshInterval.Pause),
		Value: types.Int64Value(int64(data.Data.RefreshInterval.Value)),
	}

	// Map query (KbnAsCodeQuery: language + expression string)
	q := &dashboardQueryModel{
		Language: types.StringValue(string(data.Data.Query.Language)),
	}
	expr := data.Data.Query.Expression
	trimmed := bytes.TrimSpace([]byte(expr))
	if len(trimmed) > 0 && trimmed[0] == '{' {
		var obj map[string]any
		if err := json.Unmarshal(trimmed, &obj); err == nil {
			q.Text = types.StringNull()
			q.JSON = jsontypes.NewNormalizedValue(string(trimmed))
		} else {
			q.Text = types.StringValue(expr)
			q.JSON = jsontypes.NewNormalizedNull()
		}
	} else {
		q.Text = types.StringValue(expr)
		q.JSON = jsontypes.NewNormalizedNull()
	}
	m.Query = q

	m.mapDashboardFiltersFromAPI(ctx, &data.Data, &diags)

	// Map tags
	if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
		m.Tags = typeutils.SliceToListTypeString(ctx, *data.Data.Tags, path.Root("tags"), &diags)
	} else {
		m.Tags = types.ListNull(types.StringType)
	}

	// Map options
	m.Options = m.mapOptionsFromAPI(data.Data.Options)

	// Map access control
	var accessMode *string
	if data.Data.AccessControl.AccessMode != nil {
		s := string(*data.Data.AccessControl.AccessMode)
		accessMode = &s
	}
	m.AccessControl = newAccessControlFromAPI(accessMode)

	// Map panels
	panels, sections, panelsDiags := m.mapPanelsFromAPI(ctx, data.Data.Panels)
	diags.Append(panelsDiags...)
	m.Panels = panels
	m.Sections = sections

	pinnedPanels, pinnedDiags := m.mapPinnedPanelsFromAPI(ctx, priorPinnedPanels, data.Data.PinnedPanels)
	diags.Append(pinnedDiags...)
	m.PinnedPanels = pinnedPanels

	return diags
}

// toAPICreateRequest converts the Terraform model to an API create request
func (m *dashboardModel) toAPICreateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PostDashboardsJSONRequestBody {
	req := kbapi.PostDashboardsJSONRequestBody{}
	req.Title = m.Title.ValueString()
	if m.RefreshInterval != nil {
		req.RefreshInterval.Pause = m.RefreshInterval.Pause.ValueBool()
		req.RefreshInterval.Value = float32(m.RefreshInterval.Value.ValueInt64())
	}
	if m.TimeRange != nil {
		req.TimeRange.From = m.TimeRange.From.ValueString()
		req.TimeRange.To = m.TimeRange.To.ValueString()
	}

	// Set description
	if typeutils.IsKnown(m.Description) {
		desc := m.Description.ValueString()
		req.Description = &desc
	}

	// Set time range mode
	if m.TimeRange != nil && typeutils.IsKnown(m.TimeRange.Mode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRange.Mode.ValueString())
		req.TimeRange.Mode = &mode
	}

	// Set query text - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Query = queryModel

	// Set tags
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.ListTypeToSliceString(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Tags = &tags
		}
	}

	// Set options
	options, optionsDiags := m.optionsToAPI()
	diags.Append(optionsDiags...)
	req.Options = options

	// Set access control
	req.AccessControl = m.AccessControl.toCreateAPI()

	// Set panels
	panels, panelsDiags := m.panelsToAPI(ctx)
	diags.Append(panelsDiags...)
	req.Panels = panels

	m.dashboardFiltersToCreateAPI(ctx, &req, diags)

	pinnedPanels, pinnedDiags := m.pinnedPanelsToAPICreateItems()
	diags.Append(pinnedDiags...)
	req.PinnedPanels = pinnedPanels

	return req
}

// toAPIUpdateRequest converts the Terraform model to an API update request
func (m *dashboardModel) toAPIUpdateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PutDashboardsIdJSONRequestBody {
	req := kbapi.PutDashboardsIdJSONRequestBody{}
	req.Title = m.Title.ValueString()
	if m.RefreshInterval != nil {
		req.RefreshInterval.Pause = m.RefreshInterval.Pause.ValueBool()
		req.RefreshInterval.Value = float32(m.RefreshInterval.Value.ValueInt64())
	}
	if m.TimeRange != nil {
		req.TimeRange.From = m.TimeRange.From.ValueString()
		req.TimeRange.To = m.TimeRange.To.ValueString()
	}

	// Set description
	if typeutils.IsKnown(m.Description) {
		desc := m.Description.ValueString()
		req.Description = &desc
	}

	// Set time range mode
	if m.TimeRange != nil && typeutils.IsKnown(m.TimeRange.Mode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRange.Mode.ValueString())
		req.TimeRange.Mode = &mode
	}

	// Set query text - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Query = queryModel

	// Set tags
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.ListTypeToSliceString(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Tags = &tags
		}
	}

	// Set options
	options, optionsDiags := m.optionsToAPI()
	diags.Append(optionsDiags...)
	req.Options = options

	// Set panels.
	panels, panelsDiags := m.panelsToAPI(ctx)
	diags.Append(panelsDiags...)
	if panels != nil {
		req.Panels = panels
	}

	m.dashboardFiltersToUpdateAPI(ctx, &req, diags)

	pinnedPanels, pinnedDiags := m.pinnedPanelsToAPICreateItems()
	diags.Append(pinnedDiags...)
	if pinnedPanels != nil {
		req.PinnedPanels = pinnedPanels
	}

	return req
}

func (m *dashboardModel) queryToAPI() (kbapi.KbnAsCodeQuery, diag.Diagnostics) {
	query := kbapi.KbnAsCodeQuery{}
	if m.Query == nil {
		return query, nil
	}
	query.Language = kbapi.KbnAsCodeQueryLanguage(m.Query.Language.ValueString())
	textKnown := typeutils.IsKnown(m.Query.Text)
	jsonKnown := typeutils.IsKnown(m.Query.JSON)

	if textKnown == jsonKnown {
		var diags diag.Diagnostics
		diags.AddError(
			"Invalid dashboard query",
			"Exactly one of `query.text` or `query.json` must be set.",
		)
		return query, diags
	}

	switch {
	case textKnown:
		query.Expression = m.Query.Text.ValueString()
	case jsonKnown:
		query.Expression = m.Query.JSON.ValueString()
	}

	return query, nil
}

func dashboardRootSavedFiltersElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"filter_json": jsontypes.NormalizedType{},
		},
	}
}

// mapDashboardFiltersFromAPI sets m.Filters from the API in response order.
// REQ-037 / REQ-009: when filters were unset in state and the API returns no filters (nil or empty),
// the attribute stays null rather than becoming an empty list.
func (m *dashboardModel) mapDashboardFiltersFromAPI(ctx context.Context, api *kbapi.KbnDashboardData, diags *diag.Diagnostics) {
	priorUnset := m.Filters.IsNull()
	apiFilters := api.Filters
	hasItems := apiFilters != nil && len(*apiFilters) > 0

	if !hasItems {
		if priorUnset {
			return
		}
		m.Filters = typeutils.ListValueFrom(ctx, []chartFilterJSONModel{}, dashboardRootSavedFiltersElementType(), path.Root("filters"), diags)
		return
	}

	elems := make([]chartFilterJSONModel, 0, len(*apiFilters))
	for _, item := range *apiFilters {
		fm := chartFilterJSONModel{}
		fd := fm.populateFromAPIItem(item)
		diags.Append(fd...)
		if fd.HasError() {
			return
		}
		elems = append(elems, fm)
	}
	m.Filters = typeutils.ListValueFrom(ctx, elems, dashboardRootSavedFiltersElementType(), path.Root("filters"), diags)
}

// buildDashboardFiltersForAPI converts m.Filters into the shared kbapi.DashboardFilters
// payload used by both the create (POST) and update (PUT) dashboard request bodies.
// Returns (nil, false) when the attribute is unknown/null so callers leave the request
// field untouched; returns (&empty, true) when the list is known-empty so callers send
// an explicit empty array.
func (m *dashboardModel) buildDashboardFiltersForAPI(ctx context.Context, diags *diag.Diagnostics) (*kbapi.DashboardFilters, bool) {
	if !typeutils.IsKnown(m.Filters) {
		return nil, false
	}
	elems := typeutils.ListTypeAs[chartFilterJSONModel](ctx, m.Filters, path.Root("filters"), diags)
	if diags.HasError() {
		return nil, false
	}
	items := make(kbapi.DashboardFilters, 0, len(elems))
	for _, el := range elems {
		var item kbapi.DashboardFilters_Item
		fd := decodeChartFilterJSON(el.FilterJSON, &item)
		diags.Append(fd...)
		if fd.HasError() {
			return nil, false
		}
		items = append(items, item)
	}
	return &items, true
}

func (m *dashboardModel) dashboardFiltersToCreateAPI(ctx context.Context, req *kbapi.PostDashboardsJSONRequestBody, diags *diag.Diagnostics) {
	if filters, ok := m.buildDashboardFiltersForAPI(ctx, diags); ok {
		req.Filters = filters
	}
}

func (m *dashboardModel) dashboardFiltersToUpdateAPI(ctx context.Context, req *kbapi.PutDashboardsIdJSONRequestBody, diags *diag.Diagnostics) {
	if filters, ok := m.buildDashboardFiltersForAPI(ctx, diags); ok {
		req.Filters = filters
	}
}
