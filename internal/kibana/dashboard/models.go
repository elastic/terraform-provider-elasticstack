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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// jsonNullString is the JSON encoding of null; json.Marshal uses it for unset union/API fields.
const jsonNullString = "null"

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

	// Map query
	q := &dashboardQueryModel{
		Language: types.StringValue(data.Data.Query.Language),
	}
	// Query.Query is a union type with json.RawMessage - can be string or JSON object
	queryBytes, err := json.Marshal(data.Data.Query.Query)
	if err != nil {
		diags.AddError("Failed to marshal query", err.Error())
		q.Text = types.StringNull()
		q.JSON = jsontypes.NewNormalizedNull()
	} else {
		// Try to unmarshal as string first (KQL/Lucene)
		var queryString string
		if err := json.Unmarshal(queryBytes, &queryString); err == nil {
			q.Text = types.StringValue(queryString)
			q.JSON = jsontypes.NewNormalizedNull()
		} else {
			// It's a JSON object
			q.Text = types.StringNull()
			q.JSON = jsontypes.NewNormalizedValue(string(queryBytes))
		}
	}
	m.Query = q

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
		m.AccessControl = newAccessControlFromAPI(accessMode)
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
	panels, panelsDiags := m.panelsToAPI()
	diags.Append(panelsDiags...)
	req.Panels = panels

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

	// Set panels
	panels, panelsDiags := m.panelsToAPI()
	diags.Append(panelsDiags...)
	req.Panels = panels

	return req
}

func (m *dashboardModel) queryToAPI() (kbapi.KbnEsQueryServerQuerySchema, diag.Diagnostics) {
	query := kbapi.KbnEsQueryServerQuerySchema{}
	if m.Query == nil {
		return query, nil
	}
	query.Language = m.Query.Language.ValueString()
	// Query.Query is a union type with json.RawMessage
	if typeutils.IsKnown(m.Query.Text) {
		err := query.Query.FromKbnEsQueryServerQuerySchemaQuery0(m.Query.Text.ValueString())
		if err != nil {
			return query, diagutil.FrameworkDiagFromError(err)
		}
	} else if typeutils.IsKnown(m.Query.JSON) {
		var qj map[string]any
		diags := m.Query.JSON.Unmarshal(&qj)
		if diags.HasError() {
			return query, diags
		}

		err := query.Query.FromKbnEsQueryServerQuerySchemaQuery1(qj)
		if err != nil {
			return query, diagutil.FrameworkDiagFromError(err)
		}
	}

	return query, nil
}
