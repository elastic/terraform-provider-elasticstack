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
	"net/http"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const frequencyAttr = "frequency"

var logExtractionObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"additional_index_patterns":        types.ListType{ElemType: types.StringType},
	"excluded_index_patterns":          types.ListType{ElemType: types.StringType},
	"delay":                            types.StringType,
	"docs_limit":                       types.Int64Type,
	"field_history_length":             types.Int64Type,
	"frequency":                        types.StringType,
	"lookback_period":                  types.StringType,
	"max_logs_per_page":                types.Int64Type,
	"max_logs_per_window":              types.Int64Type,
	"max_logs_per_window_cap_behavior": types.StringType,
	"max_time_window_size":             types.StringType,
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

func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func intPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func listStringPtr(ctx context.Context, list types.List) (*[]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var values []string
	return &values, list.ElementsAs(ctx, &values, false)
}

func buildInstallBody(ctx context.Context, model tfModel) (kbapi.PostSecurityEntityStoreInstallJSONRequestBody, diag.Diagnostics) {
	body := kbapi.PostSecurityEntityStoreInstallJSONRequestBody{}
	entityTypes, diags := expandEntityTypes(ctx, model.EntityTypes)
	if diags.HasError() {
		return body, diags
	}
	if len(entityTypes) > 0 {
		body.EntityTypes = installTypes(entityTypes)
	}
	if !model.HistorySnapshot.IsNull() && !model.HistorySnapshot.IsUnknown() {
		var hs historySnapshotModel
		diags.Append(model.HistorySnapshot.As(ctx, &hs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return body, diags
		}
		if p := stringPtr(hs.Frequency); p != nil {
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

func installTypes(values []string) *[]kbapi.PostSecurityEntityStoreInstallJSONBodyEntityTypes {
	if len(values) == 0 {
		return nil
	}
	out := make([]kbapi.PostSecurityEntityStoreInstallJSONBodyEntityTypes, 0, len(values))
	for _, v := range values {
		out = append(out, kbapi.PostSecurityEntityStoreInstallJSONBodyEntityTypes(v))
	}
	return &out
}

func uninstallTypes(values []string) *[]kbapi.PostSecurityEntityStoreUninstallJSONBodyEntityTypes {
	if len(values) == 0 {
		return nil
	}
	out := make([]kbapi.PostSecurityEntityStoreUninstallJSONBodyEntityTypes, 0, len(values))
	for _, v := range values {
		out = append(out, kbapi.PostSecurityEntityStoreUninstallJSONBodyEntityTypes(v))
	}
	return &out
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
	var model logExtractionModel
	var diags diag.Diagnostics
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	add, d := listStringPtr(ctx, model.AdditionalIndexPatterns)
	diags.Append(d...)
	excl, d := listStringPtr(ctx, model.ExcludedIndexPatterns)
	diags.Append(d...)
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
		AdditionalIndexPatterns: add,
		Delay:                   stringPtr(model.Delay),
		DocsLimit:               intPtr(model.DocsLimit),
		ExcludedIndexPatterns:   excl,
		FieldHistoryLength:      intPtr(model.FieldHistoryLength),
		Frequency:               stringPtr(model.Frequency),
		LookbackPeriod:          stringPtr(model.LookbackPeriod),
		MaxLogsPerPage:          intPtr(model.MaxLogsPerPage),
		MaxLogsPerWindow:        intPtr(model.MaxLogsPerWindow),
		MaxTimeWindowSize:       stringPtr(model.MaxTimeWindowSize),
	}
	if !model.MaxLogsPerWindowCapBehavior.IsNull() && !model.MaxLogsPerWindowCapBehavior.IsUnknown() {
		behavior := kbapi.PostSecurityEntityStoreInstallJSONBodyLogExtractionMaxLogsPerWindowCapBehavior(model.MaxLogsPerWindowCapBehavior.ValueString())
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
	var model logExtractionModel
	var diags diag.Diagnostics
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	add, d := listStringPtr(ctx, model.AdditionalIndexPatterns)
	diags.Append(d...)
	excl, d := listStringPtr(ctx, model.ExcludedIndexPatterns)
	diags.Append(d...)
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
		AdditionalIndexPatterns: add,
		Delay:                   stringPtr(model.Delay),
		DocsLimit:               intPtr(model.DocsLimit),
		ExcludedIndexPatterns:   excl,
		FieldHistoryLength:      intPtr(model.FieldHistoryLength),
		Frequency:               stringPtr(model.Frequency),
		LookbackPeriod:          stringPtr(model.LookbackPeriod),
		MaxLogsPerPage:          intPtr(model.MaxLogsPerPage),
		MaxLogsPerWindow:        intPtr(model.MaxLogsPerWindow),
		MaxTimeWindowSize:       stringPtr(model.MaxTimeWindowSize),
	}
	if !model.MaxLogsPerWindowCapBehavior.IsNull() && !model.MaxLogsPerWindowCapBehavior.IsUnknown() {
		behavior := kbapi.PutSecurityEntityStoreJSONBodyLogExtractionMaxLogsPerWindowCapBehavior(model.MaxLogsPerWindowCapBehavior.ValueString())
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

func getEntityStoreStatus(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, includeComponents bool) (*entityStoreStatus, []byte, diag.Diagnostics) {
	params := &kbapi.GetSecurityEntityStoreStatusParams{}
	editors := []kbapi.RequestEditorFn{kibanautil.SpaceAwarePathRequestEditor(spaceID)}
	if includeComponents {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("include_components", "true")
			req.URL.RawQuery = q.Encode()
			return nil
		})
	}

	resp, err := client.GetKibanaOapiClient().API.GetSecurityEntityStoreStatusWithResponse(ctx, params, editors...)
	if err != nil {
		return nil, nil, diagutil.FrameworkDiagFromError(err)
	}
	if d := diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK); d.HasError() {
		return nil, nil, d
	}

	var status entityStoreStatus
	if err := json.Unmarshal(resp.Body, &status); err != nil {
		return nil, nil, diagutil.FrameworkDiagFromError(err)
	}
	return &status, resp.Body, nil
}
