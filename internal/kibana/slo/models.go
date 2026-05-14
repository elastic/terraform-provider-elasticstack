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

package slo

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var tfSettingsAttrTypes = map[string]attr.Type{
	"sync_delay":               types.StringType,
	"frequency":                types.StringType,
	"sync_field":               types.StringType,
	"prevent_initial_backfill": types.BoolType,
}

// tfSloArtifactDashboardObjectType and tfArtifactsAttrTypes match the `artifacts` SingleNestedAttribute schema.
var (
	tfSloArtifactDashboardObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
		"id": types.StringType,
	}}
	tfArtifactsAttrTypes = map[string]attr.Type{
		"dashboards": types.ListType{ElemType: tfSloArtifactDashboardObjectType},
	}
)

type tfModel struct {
	ID               types.String `tfsdk:"id"`
	KibanaConnection types.List   `tfsdk:"kibana_connection"`

	SloID        types.String `tfsdk:"slo_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	SpaceID      types.String `tfsdk:"space_id"`
	BudgetMethod types.String `tfsdk:"budgeting_method"`
	Enabled      types.Bool   `tfsdk:"enabled"`

	TimeWindow []tfTimeWindow `tfsdk:"time_window"`
	Objective  []tfObjective  `tfsdk:"objective"`
	Settings   types.Object   `tfsdk:"settings"`

	GroupBy GroupByValue   `tfsdk:"group_by"`
	Tags    []types.String `tfsdk:"tags"`

	Artifacts types.Object `tfsdk:"artifacts"`

	MetricCustomIndicator    []tfMetricCustomIndicator    `tfsdk:"metric_custom_indicator"`
	HistogramCustomIndicator []tfHistogramCustomIndicator `tfsdk:"histogram_custom_indicator"`
	ApmLatencyIndicator      []tfApmLatencyIndicator      `tfsdk:"apm_latency_indicator"`
	ApmAvailabilityIndicator []tfApmAvailabilityIndicator `tfsdk:"apm_availability_indicator"`
	KqlCustomIndicator       []tfKqlCustomIndicator       `tfsdk:"kql_custom_indicator"`
	TimesliceMetricIndicator []tfTimesliceMetricIndicator `tfsdk:"timeslice_metric_indicator"`
}

type tfTimeWindow struct {
	Duration types.String `tfsdk:"duration"`
	Type     types.String `tfsdk:"type"`
}

type tfObjective struct {
	Target          types.Float64 `tfsdk:"target"`
	TimesliceTarget types.Float64 `tfsdk:"timeslice_target"`
	TimesliceWindow types.String  `tfsdk:"timeslice_window"`
}

type tfSettings struct {
	SyncDelay              types.String `tfsdk:"sync_delay"`
	Frequency              types.String `tfsdk:"frequency"`
	SyncField              types.String `tfsdk:"sync_field"`
	PreventInitialBackfill types.Bool   `tfsdk:"prevent_initial_backfill"`
}

func (m tfModel) toAPIModel() (models.Slo, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(m.TimeWindow) != 1 {
		diags.AddError("Invalid configuration", "time_window must have exactly 1 item")
		return models.Slo{}, diags
	}
	if len(m.Objective) != 1 {
		diags.AddError("Invalid configuration", "objective must have exactly 1 item")
		return models.Slo{}, diags
	}

	indicator, diags := m.indicatorToAPI()
	if diags.HasError() {
		return models.Slo{}, diags
	}

	tw := kbapi.SLOsTimeWindow{
		Type:     kbapi.SLOsTimeWindowType(m.TimeWindow[0].Type.ValueString()),
		Duration: m.TimeWindow[0].Duration.ValueString(),
	}

	obj := kbapi.SLOsObjective{
		Target: m.Objective[0].Target.ValueFloat64(),
	}
	if typeutils.IsKnown(m.Objective[0].TimesliceTarget) {
		v := m.Objective[0].TimesliceTarget.ValueFloat64()
		obj.TimesliceTarget = &v
	}
	if typeutils.IsKnown(m.Objective[0].TimesliceWindow) {
		v := m.Objective[0].TimesliceWindow.ValueString()
		obj.TimesliceWindow = &v
	}

	var settings *kbapi.SLOsSettings
	if typeutils.IsKnown(m.Settings) {
		settingsModel, settingsDiags := tfSettingsFromObject(m.Settings)
		diags.Append(settingsDiags...)
		if diags.HasError() {
			return models.Slo{}, diags
		}
		settings = settingsModel.toAPIModel()
	}

	apiModel := models.Slo{
		Name:            m.Name.ValueString(),
		Description:     m.Description.ValueString(),
		Indicator:       indicator,
		TimeWindow:      tw,
		BudgetingMethod: kbapi.SLOsBudgetingMethod(m.BudgetMethod.ValueString()),
		Objective:       obj,
		Settings:        settings,
		SpaceID:         m.SpaceID.ValueString(),
	}

	if typeutils.IsKnown(m.SloID) {
		sloID := m.SloID.ValueString()
		if sloID != "" {
			apiModel.SloID = sloID
		}
	}

	if typeutils.IsKnown(m.GroupBy) {
		// Preserve explicit empty lists. A known-but-empty group_by must round-trip
		// as an empty slice (not nil) so it is sent to the Kibana API.
		apiModel.GroupBy = []string{}
		for i, v := range m.GroupBy.Elements() {
			g, ok := v.(types.String)
			if !ok {
				diags.AddError("Invalid configuration", fmt.Sprintf("group_by[%d] is not a string", i))
				continue
			}
			if typeutils.IsKnown(g) {
				apiModel.GroupBy = append(apiModel.GroupBy, g.ValueString())
			}
		}
	}
	// Preserve an explicitly empty tags list so that clearing tags is sent to the API.
	if m.Tags != nil {
		apiModel.Tags = []string{}
	}
	for _, t := range m.Tags {
		if typeutils.IsKnown(t) {
			apiModel.Tags = append(apiModel.Tags, t.ValueString())
		}
	}

	if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
		art, artDiags := tfArtifactsToAPIModel(m.Artifacts)
		diags.Append(artDiags...)
		if diags.HasError() {
			return models.Slo{}, diags
		}
		if art != nil {
			apiModel.Artifacts = art
		}
	}

	return apiModel, diags
}

func (m tfModel) indicatorToAPI() (kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ok, ind, indDiags := m.kqlCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.apmAvailabilityIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.apmLatencyIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.histogramCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.metricCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.timesliceMetricIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	diags.AddError(
		"Invalid configuration",
		"exactly one indicator block must be set",
	)
	return kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
}

func (m *tfModel) populateFromAPI(apiModel *models.Slo) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiModel == nil {
		return diags
	}

	m.SloID = types.StringValue(apiModel.SloID)
	m.SpaceID = types.StringValue(apiModel.SpaceID)
	m.Name = types.StringValue(apiModel.Name)
	m.Description = types.StringValue(apiModel.Description)
	m.BudgetMethod = types.StringValue(string(apiModel.BudgetingMethod))
	m.Enabled = types.BoolValue(apiModel.Enabled)

	m.TimeWindow = []tfTimeWindow{{
		Duration: types.StringValue(apiModel.TimeWindow.Duration),
		Type:     types.StringValue(string(apiModel.TimeWindow.Type)),
	}}

	obj := tfObjective{
		Target: types.Float64Value(apiModel.Objective.Target),
	}
	if apiModel.Objective.TimesliceTarget != nil {
		obj.TimesliceTarget = types.Float64Value(*apiModel.Objective.TimesliceTarget)
	} else {
		obj.TimesliceTarget = types.Float64Null()
	}
	if apiModel.Objective.TimesliceWindow != nil {
		obj.TimesliceWindow = types.StringValue(*apiModel.Objective.TimesliceWindow)
	} else {
		obj.TimesliceWindow = types.StringNull()
	}
	m.Objective = []tfObjective{obj}

	if typeutils.IsKnown(m.Settings) && apiModel.Settings != nil {
		attrValues := map[string]attr.Value{
			"sync_delay":               types.StringPointerValue(apiModel.Settings.SyncDelay),
			"frequency":                types.StringPointerValue(apiModel.Settings.Frequency),
			"sync_field":               types.StringPointerValue(apiModel.Settings.SyncField),
			"prevent_initial_backfill": types.BoolPointerValue(apiModel.Settings.PreventInitialBackfill),
		}
		settingsObj, objDiags := types.ObjectValue(tfSettingsAttrTypes, attrValues)
		diags.Append(objDiags...)
		m.Settings = settingsObj
	} else {
		m.Settings = types.ObjectNull(tfSettingsAttrTypes)
	}

	// Artifacts (REQ-020 / REQ-039): when the API returns at least one dashboard
	// reference, always write `artifacts` into state (including import) even if
	// the block was not in the practitioner file. If the block exists in state
	// and the API has no references, use an empty dashboards list. If nothing
	// was stored and the API has no references, keep null to avoid
	// `{ dashboards = [] }` drift for omitted `artifacts`.
	artifactRowCount := 0
	if apiModel.Artifacts != nil && apiModel.Artifacts.Dashboards != nil {
		artifactRowCount = len(*apiModel.Artifacts.Dashboards)
	}
	priorHasArtifacts := typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull()
	switch {
	case artifactRowCount > 0:
		rows := make([]attr.Value, 0, artifactRowCount)
		for _, row := range *apiModel.Artifacts.Dashboards {
			rowObj, rowDiags := types.ObjectValue(tfSloArtifactDashboardObjectType.AttrTypes, map[string]attr.Value{
				"id": types.StringValue(row.Id),
			})
			diags.Append(rowDiags...)
			rows = append(rows, rowObj)
		}
		if diags.HasError() {
			return diags
		}
		listVal, listDiags := types.ListValue(tfSloArtifactDashboardObjectType, rows)
		diags.Append(listDiags...)
		artObj, artDiags := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": listVal,
		})
		diags.Append(artDiags...)
		m.Artifacts = artObj
	case priorHasArtifacts:
		emptyBoards, listDiags := types.ListValue(tfSloArtifactDashboardObjectType, []attr.Value{})
		diags.Append(listDiags...)
		artObj, artDiags := types.ObjectValue(tfArtifactsAttrTypes, map[string]attr.Value{
			"dashboards": emptyBoards,
		})
		diags.Append(artDiags...)
		m.Artifacts = artObj
	default:
		m.Artifacts = types.ObjectNull(tfArtifactsAttrTypes)
	}

	if apiModel.GroupBy != nil {
		groupByValues := make([]attr.Value, len(apiModel.GroupBy))
		for i, g := range apiModel.GroupBy {
			groupByValues[i] = types.StringValue(g)
		}
		groupByList, listDiags := NewGroupByValue(groupByValues)
		diags.Append(listDiags...)
		m.GroupBy = groupByList
	} else {
		m.GroupBy = NewGroupByNull()
	}
	m.Tags = nil
	for _, t := range apiModel.Tags {
		m.Tags = append(m.Tags, types.StringValue(t))
	}

	m.MetricCustomIndicator = nil
	m.HistogramCustomIndicator = nil
	m.ApmLatencyIndicator = nil
	m.ApmAvailabilityIndicator = nil
	m.KqlCustomIndicator = nil
	m.TimesliceMetricIndicator = nil

	indicatorValue, err := apiModel.Indicator.ValueByDiscriminator()
	if err != nil {
		diags.AddError("Unexpected API response", "failed to determine indicator type: "+err.Error())
		return diags
	}

	switch ind := indicatorValue.(type) {
	case kbapi.SLOsIndicatorPropertiesApmAvailability:
		diags.Append(m.populateFromApmAvailabilityIndicator(ind)...)
	case kbapi.SLOsIndicatorPropertiesApmLatency:
		diags.Append(m.populateFromApmLatencyIndicator(ind)...)
	case kbapi.SLOsIndicatorPropertiesCustomKql:
		diags.Append(m.populateFromKqlCustomIndicator(ind)...)
	case kbapi.SLOsIndicatorPropertiesHistogram:
		diags.Append(m.populateFromHistogramCustomIndicator(ind)...)
	case kbapi.SLOsIndicatorPropertiesCustomMetric:
		diags.Append(m.populateFromMetricCustomIndicator(ind)...)
	case kbapi.SLOsIndicatorPropertiesTimesliceMetric:
		diags.Append(m.populateFromTimesliceMetricIndicator(ind)...)
	default:
		diags.AddError("Unexpected API response", fmt.Sprintf("unknown indicator type: %T", indicatorValue))
		return diags
	}

	return diags
}

func tfSettingsFromObject(obj types.Object) (tfSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := obj.Attributes()

	syncDelayVal, ok := attrs["sync_delay"].(types.String)
	if !ok {
		diags.AddError("Invalid configuration", "settings.sync_delay is not a string")
		return tfSettings{}, diags
	}

	frequencyVal, ok := attrs["frequency"].(types.String)
	if !ok {
		diags.AddError("Invalid configuration", "settings.frequency is not a string")
		return tfSettings{}, diags
	}

	syncFieldVal, ok := attrs["sync_field"].(types.String)
	if !ok {
		diags.AddError("Invalid configuration", "settings.sync_field is not a string")
		return tfSettings{}, diags
	}

	preventInitialBackfillVal, ok := attrs["prevent_initial_backfill"].(types.Bool)
	if !ok {
		diags.AddError("Invalid configuration", "settings.prevent_initial_backfill is not a bool")
		return tfSettings{}, diags
	}

	return tfSettings{
		SyncDelay:              syncDelayVal,
		Frequency:              frequencyVal,
		SyncField:              syncFieldVal,
		PreventInitialBackfill: preventInitialBackfillVal,
	}, diags
}

// sloArtifactDashboardRef is the provider-side shape for a dashboard reference (`id` in JSON).
type sloArtifactDashboardRef struct {
	ID string `json:"id"`
}

// tfArtifactsToAPIModel maps Terraform `artifacts` to the create/update payload.
// It returns an error diagnostic when the block is set but a dashboard row is invalid
// (wrong type, unknown id) instead of dropping values silently.
func tfArtifactsToAPIModel(obj types.Object) (*kbapi.SLOsArtifacts, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := obj.Attributes()
	dl, ok := attrs["dashboards"].(types.List)
	if !ok {
		diags.AddError("Invalid configuration", "artifacts: dashboards is not a list")
		return nil, diags
	}
	if dl.IsNull() {
		return &kbapi.SLOsArtifacts{Dashboards: asKbapiArtifactDashboards(nil)}, diags
	}
	if dl.IsUnknown() {
		diags.AddError("Invalid configuration", "artifacts: dashboards is unknown; the value must be known to send artifacts to Kibana")
		return nil, diags
	}
	elems := dl.Elements()
	refs := make([]sloArtifactDashboardRef, 0, len(elems))
	for i, e := range elems {
		rowObj, ok := e.(types.Object)
		if !ok {
			diags.AddError("Invalid configuration", fmt.Sprintf("artifacts.dashboards[%d] must be an object", i))
			return nil, diags
		}
		idVal, ok := rowObj.Attributes()["id"].(types.String)
		if !ok {
			diags.AddError("Invalid configuration", fmt.Sprintf("artifacts.dashboards[%d] has no id attribute", i))
			return nil, diags
		}
		if idVal.IsNull() {
			diags.AddError("Invalid configuration", fmt.Sprintf("artifacts.dashboards[%d].id is null", i))
			return nil, diags
		}
		if idVal.IsUnknown() {
			diags.AddError("Invalid configuration", fmt.Sprintf("artifacts.dashboards[%d].id is unknown", i))
			return nil, diags
		}
		refs = append(refs, sloArtifactDashboardRef{ID: idVal.ValueString()})
	}
	if len(refs) == 0 {
		empty := []sloArtifactDashboardRef{}
		return &kbapi.SLOsArtifacts{Dashboards: asKbapiArtifactDashboards(&empty)}, diags
	}
	return &kbapi.SLOsArtifacts{Dashboards: asKbapiArtifactDashboards(&refs)}, diags
}

// asKbapiArtifactDashboards converts provider refs to the generated kbapi slice type (field name Id in OpenAPI).
func asKbapiArtifactDashboards(refs *[]sloArtifactDashboardRef) *[]struct {
	//nolint:revive // var-naming: must match generated kbapi / Kibana SLOsArtifacts
	Id string `json:"id"`
} {
	if refs == nil {
		empty := []struct {
			//nolint:revive // var-naming: must match generated kbapi / Kibana SLOsArtifacts
			Id string `json:"id"`
		}{}
		return &empty
	}
	if len(*refs) == 0 {
		empty := []struct {
			//nolint:revive // var-naming: must match generated kbapi / Kibana SLOsArtifacts
			Id string `json:"id"`
		}{}
		return &empty
	}
	out := make([]struct {
		//nolint:revive // var-naming: must match generated kbapi / Kibana SLOsArtifacts
		Id string `json:"id"`
	}, 0, len(*refs))
	for _, ref := range *refs {
		out = append(out, struct {
			//nolint:revive // var-naming: must match generated kbapi / Kibana SLOsArtifacts
			Id string `json:"id"`
		}{Id: ref.ID})
	}
	return &out
}

func (s tfSettings) toAPIModel() *kbapi.SLOsSettings {
	settings := kbapi.SLOsSettings{}
	hasAny := false

	if typeutils.IsKnown(s.SyncDelay) {
		v := s.SyncDelay.ValueString()
		settings.SyncDelay = &v
		hasAny = true
	}
	if typeutils.IsKnown(s.Frequency) {
		v := s.Frequency.ValueString()
		settings.Frequency = &v
		hasAny = true
	}
	if typeutils.IsKnown(s.PreventInitialBackfill) {
		v := s.PreventInitialBackfill.ValueBool()
		settings.PreventInitialBackfill = &v
		hasAny = true
	}
	if typeutils.IsKnown(s.SyncField) {
		v := s.SyncField.ValueString()
		settings.SyncField = &v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return &settings
}

func (m tfModel) hasDataViewID() bool {
	return (len(m.MetricCustomIndicator) == 1 && typeutils.IsKnown(m.MetricCustomIndicator[0].DataViewID) && m.MetricCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.HistogramCustomIndicator) == 1 && typeutils.IsKnown(m.HistogramCustomIndicator[0].DataViewID) && m.HistogramCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.KqlCustomIndicator) == 1 && typeutils.IsKnown(m.KqlCustomIndicator[0].DataViewID) && m.KqlCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.TimesliceMetricIndicator) == 1 && typeutils.IsKnown(m.TimesliceMetricIndicator[0].DataViewID) && m.TimesliceMetricIndicator[0].DataViewID.ValueString() != "")
}
