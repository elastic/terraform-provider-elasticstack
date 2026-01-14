package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dashboardModel is the top-level Terraform model
type dashboardModel struct {
	ID                   types.String         `tfsdk:"id"`
	SpaceID              types.String         `tfsdk:"space_id"`
	DashboardID          types.String         `tfsdk:"dashboard_id"`
	Title                types.String         `tfsdk:"title"`
	Description          types.String         `tfsdk:"description"`
	TimeFrom             types.String         `tfsdk:"time_from"`
	TimeTo               types.String         `tfsdk:"time_to"`
	TimeRangeMode        types.String         `tfsdk:"time_range_mode"`
	RefreshIntervalPause types.Bool           `tfsdk:"refresh_interval_pause"`
	RefreshIntervalValue types.Int64          `tfsdk:"refresh_interval_value"`
	QueryLanguage        types.String         `tfsdk:"query_language"`
	QueryText            types.String         `tfsdk:"query_text"`
	QueryJSON            jsontypes.Normalized `tfsdk:"query_json"`
	Tags                 types.List           `tfsdk:"tags"`
	Options              *optionsModel        `tfsdk:"options"`
	AccessControl        *AccessControlValue  `tfsdk:"access_control"`
	Panels               []panelModel         `tfsdk:"panels"`
	Sections             []sectionModel       `tfsdk:"sections"`
}

// populateFromAPI populates the Terraform model from the API response
func (m *dashboardModel) populateFromAPI(ctx context.Context, resp *kbapi.GetDashboardsIdResponse, dashboardID string, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics
	data := resp.JSON200

	// Set composite ID
	resourceID := clients.CompositeId{ClusterId: spaceID, ResourceId: dashboardID}
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
	m.TimeFrom = types.StringValue(data.Data.TimeRange.From)
	m.TimeTo = types.StringValue(data.Data.TimeRange.To)
	// TODO: Dashboards
	// TimeRange.Mode isn't currently returned by the API on GET requests
	// if data.Data.TimeRange.Mode != nil {
	// 	m.TimeRangeMode = types.StringValue(string(*data.Data.TimeRange.Mode))
	// } else {
	// 	m.TimeRangeMode = types.StringNull()
	// }

	// Map refresh interval
	m.RefreshIntervalPause = types.BoolValue(data.Data.RefreshInterval.Pause)
	m.RefreshIntervalValue = types.Int64Value(int64(data.Data.RefreshInterval.Value))

	// Map query
	m.QueryLanguage = types.StringValue(data.Data.Query.Language)
	// Query.Query is a union type with json.RawMessage - can be string or JSON object
	queryBytes, err := json.Marshal(data.Data.Query.Query)
	if err != nil {
		diags.AddError("Failed to marshal query", err.Error())
		m.QueryText = types.StringNull()
		m.QueryJSON = jsontypes.NewNormalizedNull()
	} else {
		// Try to unmarshal as string first (KQL/Lucene)
		var queryString string
		if err := json.Unmarshal(queryBytes, &queryString); err == nil {
			m.QueryText = types.StringValue(queryString)
			m.QueryJSON = jsontypes.NewNormalizedNull()
		} else {
			// It's a JSON object
			m.QueryText = types.StringNull()
			m.QueryJSON = jsontypes.NewNormalizedValue(string(queryBytes))
		}
	}

	// Map tags
	if data.Data.Tags != nil && len(*data.Data.Tags) > 0 {
		m.Tags = utils.SliceToListType_String(ctx, *data.Data.Tags, path.Root("tags"), &diags)
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
	panels, sections, panelsDiags := m.mapPanelsFromAPI(data.Data.Panels)
	diags.Append(panelsDiags...)
	m.Panels = panels
	m.Sections = sections

	return diags
}

// toAPICreateRequest converts the Terraform model to an API create request
func (m *dashboardModel) toAPICreateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PostDashboardsJSONRequestBody {
	req := kbapi.PostDashboardsJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	req.Data.RefreshInterval.Pause = m.RefreshIntervalPause.ValueBool()
	req.Data.RefreshInterval.Value = float32(m.RefreshIntervalValue.ValueInt64())
	req.Data.TimeRange.From = m.TimeFrom.ValueString()
	req.Data.TimeRange.To = m.TimeTo.ValueString()

	// Set optional dashboard ID
	if utils.IsKnown(m.DashboardID) {
		req.Id = utils.Pointer(m.DashboardID.ValueString())
	}

	// Set space
	if utils.IsKnown(m.SpaceID) && m.SpaceID.ValueString() != "default" {
		req.Spaces = &[]string{m.SpaceID.ValueString()}
	}

	// Set description
	if utils.IsKnown(m.Description) {
		req.Data.Description = utils.Pointer(m.Description.ValueString())
	}

	// Set time range mode
	if utils.IsKnown(m.TimeRangeMode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRangeMode.ValueString())
		req.Data.TimeRange.Mode = &mode
	}

	// Set query text - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Data.Query = queryModel

	// Set tags
	if utils.IsKnown(m.Tags) {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), diags)
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
	req.Data.AccessControl = m.AccessControl.ToCreateAPI()

	return req
}

// toAPIUpdateRequest converts the Terraform model to an API update request
func (m *dashboardModel) toAPIUpdateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PutDashboardsIdJSONRequestBody {
	req := kbapi.PutDashboardsIdJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	req.Data.RefreshInterval.Pause = m.RefreshIntervalPause.ValueBool()
	req.Data.RefreshInterval.Value = float32(m.RefreshIntervalValue.ValueInt64())
	req.Data.TimeRange.From = m.TimeFrom.ValueString()
	req.Data.TimeRange.To = m.TimeTo.ValueString()

	// Set description
	if utils.IsKnown(m.Description) {
		req.Data.Description = utils.Pointer(m.Description.ValueString())
	}

	// Set time range mode
	if utils.IsKnown(m.TimeRangeMode) {
		mode := kbapi.KbnEsQueryServerTimeRangeSchemaMode(m.TimeRangeMode.ValueString())
		req.Data.TimeRange.Mode = &mode
	}

	// Set query text - Query is a union type with json.RawMessage
	queryModel, queryDiags := m.queryToAPI()
	diags.Append(queryDiags...)
	req.Data.Query = queryModel

	// Set tags
	if utils.IsKnown(m.Tags) {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), diags)
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
	req.Data.AccessControl = m.AccessControl.ToUpdateAPI()

	return req
}

func (m *dashboardModel) queryToAPI() (kbapi.KbnEsQueryServerQuerySchema, diag.Diagnostics) {
	query := kbapi.KbnEsQueryServerQuerySchema{
		Language: m.QueryLanguage.ValueString(),
	}
	// Set query text - Query is a union type with json.RawMessage
	if utils.IsKnown(m.QueryText) {
		err := query.Query.FromKbnEsQueryServerQuerySchemaQuery0(m.QueryText.ValueString())
		if err != nil {
			return query, diagutil.FrameworkDiagFromError(err)
		}
	} else if utils.IsKnown(m.QueryJSON) {
		// For JSON queries, use the raw JSON directly
		var qj map[string]interface{}
		diags := m.QueryJSON.Unmarshal(&qj)
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
