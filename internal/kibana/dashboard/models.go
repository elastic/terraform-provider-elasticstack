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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dashboardModel is the top-level Terraform model
type dashboardModel struct {
	ID              types.String          `tfsdk:"id"`
	SpaceID         types.String          `tfsdk:"space_id"`
	DashboardID     types.String          `tfsdk:"dashboard_id"`
	Title           types.String          `tfsdk:"title"`
	Description     types.String          `tfsdk:"description"`
	TimeRange       *timeRangeModel       `tfsdk:"time_range"`
	RefreshInterval *refreshIntervalModel `tfsdk:"refresh_interval"`
	Query           *dashboardQueryModel  `tfsdk:"query"`
	Tags            types.List            `tfsdk:"tags"`
	Options         *optionsModel         `tfsdk:"options"`
	AccessControl   *AccessControlValue   `tfsdk:"access_control"`
	Panels          []panelModel          `tfsdk:"panels"`
	Sections        []sectionModel        `tfsdk:"sections"`
}

type timeRangeModel struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
	Mode types.String `tfsdk:"mode"`
}

type refreshIntervalModel struct {
	Pause types.Bool  `tfsdk:"pause"`
	Value types.Int64 `tfsdk:"value"`
}

type dashboardQueryModel struct {
	Language types.String         `tfsdk:"language"`
	Text     types.String         `tfsdk:"text"`
	JSON     jsontypes.Normalized `tfsdk:"json"`
}

// populateFromAPI populates the Terraform model from the API response
func (m *dashboardModel) populateFromAPI(ctx context.Context, resp *kbapi.GetDashboardsIdResponse, dashboardID string, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics
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

	// Map time range
	preservedMode := types.StringNull()
	if m.TimeRange != nil && typeutils.IsKnown(m.TimeRange.Mode) {
		preservedMode = m.TimeRange.Mode
	}
	m.TimeRange = &timeRangeModel{
		From: types.StringValue(data.Data.TimeRange.From),
		To:   types.StringValue(data.Data.TimeRange.To),
		// TimeRange.Mode isn't currently returned by the API on GET requests
		Mode: preservedMode,
	}

	// Map refresh interval
	m.RefreshInterval = &refreshIntervalModel{
		Pause: types.BoolValue(data.Data.RefreshInterval.Pause),
		Value: types.Int64Value(int64(data.Data.RefreshInterval.Value)),
	}

	// Map query
	if m.Query == nil {
		m.Query = &dashboardQueryModel{}
	}
	m.Query.Language = types.StringValue(data.Data.Query.Language)
	// Query.Query is a union type with json.RawMessage - can be string or JSON object
	queryBytes, err := json.Marshal(data.Data.Query.Query)
	if err != nil {
		diags.AddError("Failed to marshal query", err.Error())
		m.Query.Text = types.StringNull()
		m.Query.JSON = jsontypes.NewNormalizedNull()
	} else {
		// Try to unmarshal as string first (KQL/Lucene)
		var queryString string
		if err := json.Unmarshal(queryBytes, &queryString); err == nil {
			m.Query.Text = types.StringValue(queryString)
			m.Query.JSON = jsontypes.NewNormalizedNull()
		} else {
			// Store JSON objects as a JSON string (e.g. from `jsonencode({ ... })`)
			m.Query.Text = types.StringNull()
			m.Query.JSON = jsontypes.NewNormalizedValue(string(queryBytes))
		}
	}

	// Map tags
	if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
		m.Tags = typeutils.SliceToListTypeString(ctx, *data.Data.Tags, path.Root("tags"), &diags)
	} else {
		m.Tags = types.ListNull(types.StringType)
	}

	// Map options
	m.Options = m.mapOptionsFromAPI(data.Data.Options)

	// Map access control
	if data.Data.AccessControl != nil {
		var accessMode *string
		if data.Data.AccessControl.AccessMode != nil {
			s := string(*data.Data.AccessControl.AccessMode)
			accessMode = &s
		}
		m.AccessControl = newAccessControlFromAPI(accessMode, data.Data.AccessControl.Owner)
	}

	// Map panels
	panels, sections, panelsDiags := m.mapPanelsFromAPI(ctx, data.Data.Panels)
	diags.Append(panelsDiags...)
	m.Panels = panels
	m.Sections = sections

	return diags
}

// toAPICreateRequest converts the Terraform model to an API create request
func (m *dashboardModel) toAPICreateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PostDashboardsJSONRequestBody {
	req := kbapi.PostDashboardsJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	if m.RefreshInterval == nil {
		diags.AddError("Missing refresh_interval", "refresh_interval is required but was not provided")
		return req
	}
	if m.TimeRange == nil {
		diags.AddError("Missing time_range", "time_range is required but was not provided")
		return req
	}
	if m.Query == nil {
		diags.AddError("Missing query", "query is required but was not provided")
		return req
	}

	req.Data.RefreshInterval.Pause = m.RefreshInterval.Pause.ValueBool()
	req.Data.RefreshInterval.Value = float32(m.RefreshInterval.Value.ValueInt64())
	req.Data.TimeRange.From = m.TimeRange.From.ValueString()
	req.Data.TimeRange.To = m.TimeRange.To.ValueString()

	// Set optional dashboard ID
	if typeutils.IsKnown(m.DashboardID) {
		req.Id = schemautil.Pointer(m.DashboardID.ValueString())
	}

	// Set space
	// NOTE: Space routing is handled via the request path ("/s/{space_id}").
	// Kibana currently rejects a "spaces" field in the request body.

	// Set description
	if typeutils.IsKnown(m.Description) {
		req.Data.Description = schemautil.Pointer(m.Description.ValueString())
	}

	// Set time range mode
	if typeutils.IsKnown(m.TimeRange.Mode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRange.Mode.ValueString())
		req.Data.TimeRange.Mode = &mode
	}

	// Set query - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Data.Query = queryModel

	// Set tags
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.ListTypeToSliceString(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Data.Tags = &tags
		}
	}

	// Set options
	options, optionsDiags := m.optionsToAPI()
	diags.Append(optionsDiags...)
	req.Data.Options = options

	// Set panels
	panels, panelsDiags := m.panelsToAPI()
	diags.Append(panelsDiags...)
	req.Data.Panels = panels

	// Set access control
	req.Data.AccessControl = m.AccessControl.toCreateAPI()

	return req
}

// toAPIUpdateRequest converts the Terraform model to an API update request
func (m *dashboardModel) toAPIUpdateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PutDashboardsIdJSONRequestBody {
	req := kbapi.PutDashboardsIdJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	if m.RefreshInterval == nil {
		diags.AddError("Missing refresh_interval", "refresh_interval is required but was not provided")
		return req
	}
	if m.TimeRange == nil {
		diags.AddError("Missing time_range", "time_range is required but was not provided")
		return req
	}
	if m.Query == nil {
		diags.AddError("Missing query", "query is required but was not provided")
		return req
	}

	req.Data.RefreshInterval.Pause = m.RefreshInterval.Pause.ValueBool()
	req.Data.RefreshInterval.Value = float32(m.RefreshInterval.Value.ValueInt64())
	req.Data.TimeRange.From = m.TimeRange.From.ValueString()
	req.Data.TimeRange.To = m.TimeRange.To.ValueString()

	// Set description
	if typeutils.IsKnown(m.Description) {
		req.Data.Description = schemautil.Pointer(m.Description.ValueString())
	}

	// Set time range mode
	if typeutils.IsKnown(m.TimeRange.Mode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRange.Mode.ValueString())
		req.Data.TimeRange.Mode = &mode
	}

	// Set query - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Data.Query = queryModel

	// Set tags
	if typeutils.IsKnown(m.Tags) {
		tags := typeutils.ListTypeToSliceString(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Data.Tags = &tags
		}
	}

	// Set options
	options, optionsDiags := m.optionsToAPI()
	diags.Append(optionsDiags...)
	req.Data.Options = options

	// Set panels
	panels, panelsDiags := m.panelsToAPI()
	diags.Append(panelsDiags...)
	req.Data.Panels = panels

	// Set access control
	req.Data.AccessControl = m.AccessControl.toUpdateAPI()

	return req
}

func (m *dashboardModel) queryToAPI() (kbapi.KbnEsQueryServerQuerySchema, diag.Diagnostics) {
	if m.Query == nil {
		var diags diag.Diagnostics
		diags.AddError("Missing query", "query is required but was not provided")
		return kbapi.KbnEsQueryServerQuerySchema{}, diags
	}
	query := kbapi.KbnEsQueryServerQuerySchema{
		Language: m.Query.Language.ValueString(),
	}

	textKnown := typeutils.IsKnown(m.Query.Text)
	jsonKnown := !m.Query.JSON.IsNull() && !m.Query.JSON.IsUnknown()

	if textKnown && jsonKnown {
		var diags diag.Diagnostics
		diags.AddError("Invalid query configuration", "Exactly one of query.text or query.json must be set")
		return query, diags
	}
	if !textKnown && !jsonKnown {
		var diags diag.Diagnostics
		diags.AddError("Missing query configuration", "Exactly one of query.text or query.json must be set")
		return query, diags
	}

	// Set query - union type (string or JSON object)
	if jsonKnown {
		raw := m.Query.JSON.ValueString()
		var decoded any
		if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
			return query, diagutil.FrameworkDiagFromError(err)
		}
		obj, ok := decoded.(map[string]any)
		if !ok {
			var diags diag.Diagnostics
			diags.AddError("Invalid query.json value", "query.json must encode a JSON object")
			return query, diags
		}
		if err := query.Query.FromKbnEsQueryServerQuerySchemaQuery1(obj); err != nil {
			return query, diagutil.FrameworkDiagFromError(err)
		}
		return query, nil
	}

	if err := query.Query.FromKbnEsQueryServerQuerySchemaQuery0(m.Query.Text.ValueString()); err != nil {
		return query, diagutil.FrameworkDiagFromError(err)
	}

	return query, nil
}
