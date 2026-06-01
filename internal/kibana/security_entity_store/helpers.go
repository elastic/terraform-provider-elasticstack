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

package security_entity_store

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	delayAttr              = "delay"
	fieldHistoryLengthAttr = "field_history_length"
	frequencyAttr          = "frequency"
	installedAttr          = "installed"
	lookbackPeriodAttr     = "lookback_period"
)

var logExtractionObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"additional_index_patterns":        types.ListType{ElemType: types.StringType},
	"excluded_index_patterns":          types.ListType{ElemType: types.StringType},
	delayAttr:                          types.StringType,
	"docs_limit":                       types.Int64Type,
	fieldHistoryLengthAttr:             types.Int64Type,
	frequencyAttr:                      types.StringType,
	lookbackPeriodAttr:                 types.StringType,
	"max_logs_per_page":                types.Int64Type,
	"max_logs_per_window":              types.Int64Type,
	"max_logs_per_window_cap_behavior": types.StringType,
	"max_time_window_size":             types.StringType,
}}

var engineObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"type":                 types.StringType,
	"status":               types.StringType,
	"index_pattern":        types.StringType,
	fieldHistoryLengthAttr: types.Int64Type,
	delayAttr:              types.StringType,
	frequencyAttr:          types.StringType,
	lookbackPeriodAttr:     types.StringType,
	"filter":               types.StringType,
	"timeout":              types.StringType,
	"timestamp_field":      types.StringType,
	"error_action":         types.StringType,
	"error_message":        types.StringType,
	"components":           types.ListType{ElemType: engineComponentObjectType},
}}

var engineComponentObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id":          types.StringType,
	installedAttr: types.BoolType,
	"resource":    types.StringType,
	"health":      types.StringType,
}}

// entityStoreStatus mirrors the JSON shape returned by GET /api/security/entity_store/status.
type entityStoreStatus struct {
	Status  kbapi.SecurityEntityAnalyticsAPIStoreStatus `json:"status"`
	Engines []entityStoreEngine                         `json:"engines"`
}

type entityStoreEngine struct {
	Components         *[]kbapi.SecurityEntityAnalyticsAPIEngineComponentStatus `json:"components,omitempty"`
	Delay              *string                                                  `json:"delay,omitempty"`
	DocsPerSecond      *int                                                     `json:"docsPerSecond,omitempty"`
	Error              *entityStoreEngineError                                  `json:"error,omitempty"`
	FieldHistoryLength int                                                      `json:"fieldHistoryLength"`
	Filter             *string                                                  `json:"filter,omitempty"`
	Frequency          *string                                                  `json:"frequency,omitempty"`
	IndexPattern       string                                                   `json:"indexPattern"`
	LookbackPeriod     *string                                                  `json:"lookbackPeriod,omitempty"`
	Status             kbapi.SecurityEntityAnalyticsAPIEngineStatus             `json:"status"`
	Timeout            *string                                                  `json:"timeout,omitempty"`
	TimestampField     *string                                                  `json:"timestampField,omitempty"`
	Type               kbapi.SecurityEntityAnalyticsAPIEntityType               `json:"type"`
}

type entityStoreEngineError struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func intPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func stringListPtr(ctx context.Context, list types.List) (*[]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var diags diag.Diagnostics
	result := typeutils.ListTypeToSliceString(ctx, list, path.Empty(), &diags)
	if diags.HasError() {
		return nil, diags
	}
	return &result, diags
}

func buildInstallBody(ctx context.Context, model tfModel) (kbapi.PostSecurityEntityStoreInstallJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PostSecurityEntityStoreInstallJSONRequestBody{}
	entityTypes, diags := expandEntityTypes(ctx, model.EntityTypes)
	if diags.HasError() {
		return body, diags
	}
	if len(entityTypes) > 0 {
		body.EntityTypes = stringSliceToAPITypes[kbapi.PostSecurityEntityStoreInstallJSONBodyEntityTypes](entityTypes)
	}
	if !model.HistorySnapshot.IsNull() && !model.HistorySnapshot.IsUnknown() {
		var hs historySnapshotModel
		diags.Append(model.HistorySnapshot.As(ctx, &hs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return body, diags
		}
		if p := typeutils.OptionalString(hs.Frequency); p != nil {
			body.HistorySnapshot = &struct {
				Frequency *string `json:"frequency,omitempty"`
			}{Frequency: p}
		}
	}
	if !model.LogExtraction.IsNull() && !model.LogExtraction.IsUnknown() {
		le, d := expandInstallLogExtraction(ctx, model.LogExtraction)
		diags.Append(d...)
		if diags.HasError() {
			return body, diags
		}
		body.LogExtraction = le
	}
	return body, diags
}

func buildUpdateBody(ctx context.Context, model tfModel) (kbapi.PutSecurityEntityStoreJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PutSecurityEntityStoreJSONRequestBody{}
	if model.LogExtraction.IsNull() || model.LogExtraction.IsUnknown() {
		return body, nil
	}
	le, diags := expandUpdateLogExtraction(ctx, model.LogExtraction)
	if diags.HasError() {
		return body, diags
	}
	body.LogExtraction = *le
	return body, diags
}

func expandEntityTypes(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var values []string
	return values, set.ElementsAs(ctx, &values, false)
}

// stringSliceToAPITypes converts a []string to a pointer to a slice of a ~string
// type alias, covering the pattern used by generated kbapi enum slices.
func stringSliceToAPITypes[T ~string](values []string) *[]T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, 0, len(values))
	for _, v := range values {
		out = append(out, T(v))
	}
	return &out
}

// logExtractionCommon holds type-neutral parsed values from a logExtractionModel,
// used to eliminate duplicated parsing logic between the install and update paths.
type logExtractionCommon struct {
	AdditionalIndexPatterns     *[]string
	Delay                       *string
	DocsLimit                   *int
	ExcludedIndexPatterns       *[]string
	FieldHistoryLength          *int
	Frequency                   *string
	LookbackPeriod              *string
	MaxLogsPerPage              *int
	MaxLogsPerWindow            *int
	MaxLogsPerWindowCapBehavior *string
	MaxTimeWindowSize           *string
}

func expandLogExtractionCommon(ctx context.Context, obj types.Object) (*logExtractionCommon, diag.Diagnostics) {
	var model logExtractionModel
	var diags diag.Diagnostics
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	add, d := stringListPtr(ctx, model.AdditionalIndexPatterns)
	diags.Append(d...)
	excl, d := stringListPtr(ctx, model.ExcludedIndexPatterns)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	c := &logExtractionCommon{
		AdditionalIndexPatterns: add,
		Delay:                   typeutils.OptionalString(model.Delay),
		DocsLimit:               intPtr(model.DocsLimit),
		ExcludedIndexPatterns:   excl,
		FieldHistoryLength:      intPtr(model.FieldHistoryLength),
		Frequency:               typeutils.OptionalString(model.Frequency),
		LookbackPeriod:          typeutils.OptionalString(model.LookbackPeriod),
		MaxLogsPerPage:          intPtr(model.MaxLogsPerPage),
		MaxLogsPerWindow:        intPtr(model.MaxLogsPerWindow),
		MaxTimeWindowSize:       typeutils.OptionalString(model.MaxTimeWindowSize),
	}
	if !model.MaxLogsPerWindowCapBehavior.IsNull() && !model.MaxLogsPerWindowCapBehavior.IsUnknown() {
		s := model.MaxLogsPerWindowCapBehavior.ValueString()
		c.MaxLogsPerWindowCapBehavior = &s
	}
	return c, diags
}

func expandInstallLogExtraction(ctx context.Context, obj types.Object) (*struct {
	AdditionalIndexPatterns     *[]string                                                                             `json:"additionalIndexPatterns,omitempty"`
	Delay                       *string                                                                               `json:"delay,omitempty"`
	DocsLimit                   *int                                                                                  `json:"docsLimit,omitempty"`
	ExcludedIndexPatterns       *[]string                                                                             `json:"excludedIndexPatterns,omitempty"`
	FieldHistoryLength          *int                                                                                  `json:"fieldHistoryLength,omitempty"`
	Frequency                   *string                                                                               `json:"frequency,omitempty"`
	LookbackPeriod              *string                                                                               `json:"lookbackPeriod,omitempty"`
	MaxLogsPerPage              *int                                                                                  `json:"maxLogsPerPage,omitempty"`
	MaxLogsPerWindow            *int                                                                                  `json:"maxLogsPerWindow,omitempty"`
	MaxLogsPerWindowCapBehavior *kbapi.PostSecurityEntityStoreInstallJSONBodyLogExtractionMaxLogsPerWindowCapBehavior `json:"maxLogsPerWindowCapBehavior,omitempty"`
	MaxTimeWindowSize           *string                                                                               `json:"maxTimeWindowSize,omitempty"`
}, diag.Diagnostics) {
	c, diags := expandLogExtractionCommon(ctx, obj)
	if diags.HasError() {
		return nil, diags
	}
	result := &struct {
		AdditionalIndexPatterns     *[]string                                                                             `json:"additionalIndexPatterns,omitempty"`
		Delay                       *string                                                                               `json:"delay,omitempty"`
		DocsLimit                   *int                                                                                  `json:"docsLimit,omitempty"`
		ExcludedIndexPatterns       *[]string                                                                             `json:"excludedIndexPatterns,omitempty"`
		FieldHistoryLength          *int                                                                                  `json:"fieldHistoryLength,omitempty"`
		Frequency                   *string                                                                               `json:"frequency,omitempty"`
		LookbackPeriod              *string                                                                               `json:"lookbackPeriod,omitempty"`
		MaxLogsPerPage              *int                                                                                  `json:"maxLogsPerPage,omitempty"`
		MaxLogsPerWindow            *int                                                                                  `json:"maxLogsPerWindow,omitempty"`
		MaxLogsPerWindowCapBehavior *kbapi.PostSecurityEntityStoreInstallJSONBodyLogExtractionMaxLogsPerWindowCapBehavior `json:"maxLogsPerWindowCapBehavior,omitempty"`
		MaxTimeWindowSize           *string                                                                               `json:"maxTimeWindowSize,omitempty"`
	}{
		AdditionalIndexPatterns: c.AdditionalIndexPatterns,
		Delay:                   c.Delay,
		DocsLimit:               c.DocsLimit,
		ExcludedIndexPatterns:   c.ExcludedIndexPatterns,
		FieldHistoryLength:      c.FieldHistoryLength,
		Frequency:               c.Frequency,
		LookbackPeriod:          c.LookbackPeriod,
		MaxLogsPerPage:          c.MaxLogsPerPage,
		MaxLogsPerWindow:        c.MaxLogsPerWindow,
		MaxTimeWindowSize:       c.MaxTimeWindowSize,
	}
	if c.MaxLogsPerWindowCapBehavior != nil {
		behavior := kbapi.PostSecurityEntityStoreInstallJSONBodyLogExtractionMaxLogsPerWindowCapBehavior(*c.MaxLogsPerWindowCapBehavior)
		result.MaxLogsPerWindowCapBehavior = &behavior
	}
	return result, diags
}

func expandUpdateLogExtraction(ctx context.Context, obj types.Object) (*struct {
	AdditionalIndexPatterns     *[]string                                                                     `json:"additionalIndexPatterns,omitempty"`
	Delay                       *string                                                                       `json:"delay,omitempty"`
	DocsLimit                   *int                                                                          `json:"docsLimit,omitempty"`
	ExcludedIndexPatterns       *[]string                                                                     `json:"excludedIndexPatterns,omitempty"`
	FieldHistoryLength          *int                                                                          `json:"fieldHistoryLength,omitempty"`
	Frequency                   *string                                                                       `json:"frequency,omitempty"`
	LookbackPeriod              *string                                                                       `json:"lookbackPeriod,omitempty"`
	MaxLogsPerPage              *int                                                                          `json:"maxLogsPerPage,omitempty"`
	MaxLogsPerWindow            *int                                                                          `json:"maxLogsPerWindow,omitempty"`
	MaxLogsPerWindowCapBehavior *kbapi.PutSecurityEntityStoreJSONBodyLogExtractionMaxLogsPerWindowCapBehavior `json:"maxLogsPerWindowCapBehavior,omitempty"`
	MaxTimeWindowSize           *string                                                                       `json:"maxTimeWindowSize,omitempty"`
}, diag.Diagnostics) {
	c, diags := expandLogExtractionCommon(ctx, obj)
	if diags.HasError() {
		return nil, diags
	}
	result := &struct {
		AdditionalIndexPatterns     *[]string                                                                     `json:"additionalIndexPatterns,omitempty"`
		Delay                       *string                                                                       `json:"delay,omitempty"`
		DocsLimit                   *int                                                                          `json:"docsLimit,omitempty"`
		ExcludedIndexPatterns       *[]string                                                                     `json:"excludedIndexPatterns,omitempty"`
		FieldHistoryLength          *int                                                                          `json:"fieldHistoryLength,omitempty"`
		Frequency                   *string                                                                       `json:"frequency,omitempty"`
		LookbackPeriod              *string                                                                       `json:"lookbackPeriod,omitempty"`
		MaxLogsPerPage              *int                                                                          `json:"maxLogsPerPage,omitempty"`
		MaxLogsPerWindow            *int                                                                          `json:"maxLogsPerWindow,omitempty"`
		MaxLogsPerWindowCapBehavior *kbapi.PutSecurityEntityStoreJSONBodyLogExtractionMaxLogsPerWindowCapBehavior `json:"maxLogsPerWindowCapBehavior,omitempty"`
		MaxTimeWindowSize           *string                                                                       `json:"maxTimeWindowSize,omitempty"`
	}{
		AdditionalIndexPatterns: c.AdditionalIndexPatterns,
		Delay:                   c.Delay,
		DocsLimit:               c.DocsLimit,
		ExcludedIndexPatterns:   c.ExcludedIndexPatterns,
		FieldHistoryLength:      c.FieldHistoryLength,
		Frequency:               c.Frequency,
		LookbackPeriod:          c.LookbackPeriod,
		MaxLogsPerPage:          c.MaxLogsPerPage,
		MaxLogsPerWindow:        c.MaxLogsPerWindow,
		MaxTimeWindowSize:       c.MaxTimeWindowSize,
	}
	if c.MaxLogsPerWindowCapBehavior != nil {
		behavior := kbapi.PutSecurityEntityStoreJSONBodyLogExtractionMaxLogsPerWindowCapBehavior(*c.MaxLogsPerWindowCapBehavior)
		result.MaxLogsPerWindowCapBehavior = &behavior
	}
	return result, diags
}

func diffEntityTypes(ctx context.Context, prior, plan types.Set) (added, removed []string, diags diag.Diagnostics) {
	var priorVals, planVals []string
	if !prior.IsNull() && !prior.IsUnknown() {
		diags.Append(prior.ElementsAs(ctx, &priorVals, false)...)
	}
	if !plan.IsNull() && !plan.IsUnknown() {
		diags.Append(plan.ElementsAs(ctx, &planVals, false)...)
	}
	if diags.HasError() {
		return nil, nil, diags
	}

	priorSet := make(map[string]bool, len(priorVals))
	for _, v := range priorVals {
		priorSet[v] = true
	}
	planSet := make(map[string]bool, len(planVals))
	for _, v := range planVals {
		planSet[v] = true
	}

	for v := range planSet {
		if !priorSet[v] {
			added = append(added, v)
		}
	}
	for v := range priorSet {
		if !planSet[v] {
			removed = append(removed, v)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	return added, removed, diags
}

func flattenStatus(ctx context.Context, engines []entityStoreEngine) (entityTypes types.Set, started bool, logExtraction types.Object, diags diag.Diagnostics) {
	if len(engines) == 0 {
		entityTypes, diags = types.SetValue(types.StringType, nil)
		return entityTypes, false, types.ObjectNull(logExtractionObjectType.AttrTypes), diags
	}

	typesList := make([]string, 0, len(engines))
	for _, e := range engines {
		typesList = append(typesList, string(e.Type))
		if e.Status == kbapi.SecurityEntityAnalyticsAPIEngineStatusStarted {
			started = true
		}
	}
	sort.Strings(typesList)
	entityTypes, diags = types.SetValueFrom(ctx, types.StringType, typesList)
	if diags.HasError() {
		return types.SetNull(types.StringType), false, types.ObjectNull(logExtractionObjectType.AttrTypes), diags
	}

	first := engines[0]
	leModel := logExtractionModel{
		AdditionalIndexPatterns:     types.ListNull(types.StringType),
		ExcludedIndexPatterns:       types.ListNull(types.StringType),
		Delay:                       types.StringNull(),
		DocsLimit:                   types.Int64Null(),
		FieldHistoryLength:          types.Int64Null(),
		Frequency:                   types.StringNull(),
		LookbackPeriod:              types.StringNull(),
		MaxLogsPerPage:              types.Int64Null(),
		MaxLogsPerWindow:            types.Int64Null(),
		MaxLogsPerWindowCapBehavior: types.StringNull(),
		MaxTimeWindowSize:           types.StringNull(),
	}
	if first.Delay != nil {
		leModel.Delay = types.StringValue(*first.Delay)
	}
	if first.FieldHistoryLength != 0 {
		leModel.FieldHistoryLength = types.Int64Value(int64(first.FieldHistoryLength))
	}
	if first.Frequency != nil {
		leModel.Frequency = types.StringValue(*first.Frequency)
	}
	if first.LookbackPeriod != nil {
		leModel.LookbackPeriod = types.StringValue(*first.LookbackPeriod)
	}

	logExtraction, diags = types.ObjectValueFrom(ctx, logExtractionObjectType.AttrTypes, leModel)
	if diags.HasError() {
		return entityTypes, started, types.ObjectNull(logExtractionObjectType.AttrTypes), diags
	}
	return entityTypes, started, logExtraction, diags
}

func flattenEngines(ctx context.Context, engines []entityStoreEngine) (types.List, diag.Diagnostics) {
	if len(engines) == 0 {
		return types.ListValueFrom(ctx, engineObjectType, []engineModel{})
	}

	elems := make([]engineModel, 0, len(engines))
	for _, e := range engines {
		em := engineModel{
			Type:               types.StringValue(string(e.Type)),
			Status:             types.StringValue(string(e.Status)),
			IndexPattern:       types.StringValue(e.IndexPattern),
			FieldHistoryLength: types.Int64Value(int64(e.FieldHistoryLength)),
		}
		if e.Delay != nil {
			em.Delay = types.StringValue(*e.Delay)
		} else {
			em.Delay = types.StringNull()
		}
		if e.Frequency != nil {
			em.Frequency = types.StringValue(*e.Frequency)
		} else {
			em.Frequency = types.StringNull()
		}
		if e.LookbackPeriod != nil {
			em.LookbackPeriod = types.StringValue(*e.LookbackPeriod)
		} else {
			em.LookbackPeriod = types.StringNull()
		}
		if e.Filter != nil {
			em.Filter = types.StringValue(*e.Filter)
		} else {
			em.Filter = types.StringNull()
		}
		if e.Timeout != nil {
			em.Timeout = types.StringValue(*e.Timeout)
		} else {
			em.Timeout = types.StringNull()
		}
		if e.TimestampField != nil {
			em.TimestampField = types.StringValue(*e.TimestampField)
		} else {
			em.TimestampField = types.StringNull()
		}
		if e.Error != nil {
			em.ErrorAction = types.StringValue(e.Error.Action)
			em.ErrorMessage = types.StringValue(e.Error.Message)
		} else {
			em.ErrorAction = types.StringNull()
			em.ErrorMessage = types.StringNull()
		}
		if e.Components != nil && len(*e.Components) > 0 {
			components := make([]engineComponentModel, 0, len(*e.Components))
			for _, c := range *e.Components {
				cm := engineComponentModel{
					ID:        types.StringValue(c.Id),
					Installed: types.BoolValue(c.Installed),
					Resource:  types.StringValue(string(c.Resource)),
				}
				if c.Health != nil {
					cm.Health = types.StringValue(string(*c.Health))
				} else {
					cm.Health = types.StringNull()
				}
				components = append(components, cm)
			}
			cmpList, diags := types.ListValueFrom(ctx, engineComponentObjectType, components)
			if diags.HasError() {
				return types.ListNull(engineObjectType), diags
			}
			em.Components = cmpList
		} else {
			em.Components = types.ListNull(engineComponentObjectType)
		}
		elems = append(elems, em)
	}

	list, diags := types.ListValueFrom(ctx, engineObjectType, elems)
	if diags.HasError() {
		return types.ListNull(engineObjectType), diags
	}
	return list, nil
}

func getEntityStoreStatus(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, includeComponents bool) (*entityStoreStatus, []byte, diag.Diagnostics) {
	resp, diags := kibanaoapi.GetSecurityEntityStoreStatus(ctx, client.GetKibanaOapiClient(), spaceID, includeComponents)
	if diags.HasError() {
		return nil, nil, diags
	}

	var status entityStoreStatus
	if err := json.Unmarshal(resp.Body, &status); err != nil {
		return nil, nil, diagutil.FrameworkDiagFromError(err)
	}
	return &status, resp.Body, nil
}
