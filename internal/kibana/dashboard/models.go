package dashboard

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	Options              types.Object         `tfsdk:"options"`
}

type optionsModel struct {
	HidePanelTitles types.Bool `tfsdk:"hide_panel_titles"`
	UseMargins      types.Bool `tfsdk:"use_margins"`
	SyncColors      types.Bool `tfsdk:"sync_colors"`
	SyncTooltips    types.Bool `tfsdk:"sync_tooltips"`
	SyncCursor      types.Bool `tfsdk:"sync_cursor"`
}

// populateFromAPI populates the Terraform model from the API response
// It accepts any of the dashboard API response types (GET, POST, PUT)
func (m *dashboardModel) populateFromAPI(ctx context.Context, resp interface{}, dashboardID string, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Extract the JSON200 data based on response type
	var data struct {
		Data struct {
			ControlGroupInput *struct {
				AutoApplySelections  *bool
				ChainingSystem       interface{}
				Controls             interface{}
				Enhancements         *map[string]interface{}
				IgnoreParentSettings *struct {
					IgnoreFilters     *bool
					IgnoreQuery       *bool
					IgnoreTimerange   *bool
					IgnoreValidations *bool
				}
				LabelPosition interface{}
			}
			Description *string
			Filters     *[]kbapi.KbnEsQueryServerStoredFilterSchema
			Options     *struct {
				HidePanelTitles *bool
				SyncColors      *bool
				SyncCursor      *bool
				SyncTooltips    *bool
				UseMargins      *bool
			}
			Panels          interface{}
			Query           kbapi.KbnEsQueryServerQuerySchema
			RefreshInterval kbapi.KbnDataServiceServerRefreshIntervalSchema
			Tags            *[]string
			TimeRange       kbapi.KbnEsQueryServerTimeRangeSchema
			Title           string
			Version         *float32
		}
		References []kbapi.KbnContentManagementUtilsReferenceSchema
	}

	// Use type switch to handle different response types
	switch v := resp.(type) {
	case *kbapi.PostDashboardsDashboardResponse:
		if v == nil || v.JSON200 == nil {
			diags.AddError("Invalid API response", "Response or JSON200 is nil")
			return diags
		}
		// Marshal and unmarshal to convert to our generic struct
		respBytes, _ := json.Marshal(v.JSON200)
		if err := json.Unmarshal(respBytes, &data); err != nil {
			diags.AddError("Failed to unmarshal response", err.Error())
			return diags
		}
	case *kbapi.GetDashboardsDashboardIdResponse:
		if v == nil || v.JSON200 == nil {
			diags.AddError("Invalid API response", "Response or JSON200 is nil")
			return diags
		}
		respBytes, _ := json.Marshal(v.JSON200)
		if err := json.Unmarshal(respBytes, &data); err != nil {
			diags.AddError("Failed to unmarshal response", err.Error())
			return diags
		}
	case *kbapi.PutDashboardsDashboardIdResponse:
		if v == nil || v.JSON200 == nil {
			diags.AddError("Invalid API response", "Response or JSON200 is nil")
			return diags
		}
		respBytes, _ := json.Marshal(v.JSON200)
		if err := json.Unmarshal(respBytes, &data); err != nil {
			diags.AddError("Failed to unmarshal response", err.Error())
			return diags
		}
	default:
		diags.AddError("Invalid response type", fmt.Sprintf("Unexpected type %T", resp))
		return diags
	}

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
	if data.Data.Options != nil {
		// Convert via JSON to get proper types
		optBytes, _ := json.Marshal(data.Data.Options)
		var options struct {
			HidePanelTitles *bool `json:"hidePanelTitles,omitempty"`
			SyncColors      *bool `json:"syncColors,omitempty"`
			SyncCursor      *bool `json:"syncCursor,omitempty"`
			SyncTooltips    *bool `json:"syncTooltips,omitempty"`
			UseMargins      *bool `json:"useMargins,omitempty"`
		}
		if err := json.Unmarshal(optBytes, &options); err != nil {
			diags.AddError("Failed to unmarshal options", err.Error())
			m.Options = types.ObjectNull(getOptionsAttrTypes())
		} else {
			m.Options = m.mapOptionsFromAPI(ctx, &options, &diags)
		}
	} else {
		m.Options = types.ObjectNull(getOptionsAttrTypes())
	}

	return diags
}

func (m *dashboardModel) mapOptionsFromAPI(ctx context.Context, options *struct {
	HidePanelTitles *bool `json:"hidePanelTitles,omitempty"`
	SyncColors      *bool `json:"syncColors,omitempty"`
	SyncCursor      *bool `json:"syncCursor,omitempty"`
	SyncTooltips    *bool `json:"syncTooltips,omitempty"`
	UseMargins      *bool `json:"useMargins,omitempty"`
}, diags *diag.Diagnostics) types.Object {
	if options == nil {
		return types.ObjectNull(getOptionsAttrTypes())
	}

	model := optionsModel{}

	if options.HidePanelTitles != nil {
		model.HidePanelTitles = types.BoolValue(*options.HidePanelTitles)
	} else {
		model.HidePanelTitles = types.BoolNull()
	}

	if options.UseMargins != nil {
		model.UseMargins = types.BoolValue(*options.UseMargins)
	} else {
		model.UseMargins = types.BoolNull()
	}

	if options.SyncColors != nil {
		model.SyncColors = types.BoolValue(*options.SyncColors)
	} else {
		model.SyncColors = types.BoolNull()
	}

	if options.SyncTooltips != nil {
		model.SyncTooltips = types.BoolValue(*options.SyncTooltips)
	} else {
		model.SyncTooltips = types.BoolNull()
	}

	if options.SyncCursor != nil {
		model.SyncCursor = types.BoolValue(*options.SyncCursor)
	} else {
		model.SyncCursor = types.BoolNull()
	}

	obj, d := types.ObjectValueFrom(ctx, getOptionsAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// toAPICreateRequest converts the Terraform model to an API create request
func (m *dashboardModel) toAPICreateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PostDashboardsDashboardJSONRequestBody {
	req := kbapi.PostDashboardsDashboardJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	req.Data.Query.Language = m.QueryLanguage.ValueString()
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
	if utils.IsKnown(m.QueryText) {
		err := req.Data.Query.Query.FromKbnEsQueryServerQuerySchemaQuery0(m.QueryText.ValueString())
		if err != nil {
			diags.AddError("Failed to set query text", err.Error())
		}
	} else if utils.IsKnown(m.QueryJSON) {
		// For JSON queries, use the raw JSON directly
		var qj map[string]interface{}
		diags.Append(m.QueryJSON.Unmarshal(&qj)...)
		err := req.Data.Query.Query.FromKbnEsQueryServerQuerySchemaQuery1(qj)
		if err != nil {
			diags.AddError("Failed to set query JSON", err.Error())
		}
	}

	// Set tags
	if utils.IsKnown(m.Tags) {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Data.Tags = &tags
		}
	}

	// Set options
	if utils.IsKnown(m.Options) {
		var optModel optionsModel
		diags.Append(m.Options.As(ctx, &optModel, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			req.Data.Options = &struct {
				HidePanelTitles *bool `json:"hidePanelTitles,omitempty"`
				SyncColors      *bool `json:"syncColors,omitempty"`
				SyncCursor      *bool `json:"syncCursor,omitempty"`
				SyncTooltips    *bool `json:"syncTooltips,omitempty"`
				UseMargins      *bool `json:"useMargins,omitempty"`
			}{}
			if utils.IsKnown(optModel.HidePanelTitles) {
				req.Data.Options.HidePanelTitles = optModel.HidePanelTitles.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.UseMargins) {
				req.Data.Options.UseMargins = optModel.UseMargins.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncColors) {
				req.Data.Options.SyncColors = optModel.SyncColors.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncTooltips) {
				req.Data.Options.SyncTooltips = optModel.SyncTooltips.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncCursor) {
				req.Data.Options.SyncCursor = optModel.SyncCursor.ValueBoolPointer()
			}
		}
	}

	return req
}

// toAPIUpdateRequest converts the Terraform model to an API update request
func (m *dashboardModel) toAPIUpdateRequest(ctx context.Context, diags *diag.Diagnostics) kbapi.PutDashboardsDashboardIdJSONRequestBody {
	req := kbapi.PutDashboardsDashboardIdJSONRequestBody{}
	req.Data.Title = m.Title.ValueString()
	req.Data.Query.Language = m.QueryLanguage.ValueString()
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
	if utils.IsKnown(m.QueryText) {
		err := req.Data.Query.Query.FromKbnEsQueryServerQuerySchemaQuery0(m.QueryText.ValueString())
		if err != nil {
			diags.AddError("Failed to set query text", err.Error())
		}
	} else if utils.IsKnown(m.QueryJSON) {
		// For JSON queries, use the raw JSON directly
		var qj map[string]interface{}
		diags.Append(m.QueryJSON.Unmarshal(&qj)...)
		err := req.Data.Query.Query.FromKbnEsQueryServerQuerySchemaQuery1(qj)
		if err != nil {
			diags.AddError("Failed to set query JSON", err.Error())
		}
	}

	// Set tags
	if utils.IsKnown(m.Tags) {
		tags := utils.ListTypeToSlice_String(ctx, m.Tags, path.Root("tags"), diags)
		if tags != nil {
			req.Data.Tags = &tags
		}
	}

	// Set options
	if utils.IsKnown(m.Options) {
		var optModel optionsModel
		diags.Append(m.Options.As(ctx, &optModel, basetypes.ObjectAsOptions{})...)
		if !diags.HasError() {
			req.Data.Options = &struct {
				HidePanelTitles *bool `json:"hidePanelTitles,omitempty"`
				SyncColors      *bool `json:"syncColors,omitempty"`
				SyncCursor      *bool `json:"syncCursor,omitempty"`
				SyncTooltips    *bool `json:"syncTooltips,omitempty"`
				UseMargins      *bool `json:"useMargins,omitempty"`
			}{}
			if utils.IsKnown(optModel.HidePanelTitles) {
				req.Data.Options.HidePanelTitles = optModel.HidePanelTitles.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.UseMargins) {
				req.Data.Options.UseMargins = optModel.UseMargins.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncColors) {
				req.Data.Options.SyncColors = optModel.SyncColors.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncTooltips) {
				req.Data.Options.SyncTooltips = optModel.SyncTooltips.ValueBoolPointer()
			}
			if utils.IsKnown(optModel.SyncCursor) {
				req.Data.Options.SyncCursor = optModel.SyncCursor.ValueBoolPointer()
			}
		}
	}

	return req
}
