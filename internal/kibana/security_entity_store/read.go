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
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func readEntityStore(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	status, rawBody, diags := waitForStarted(ctx, client, spaceID)
	if diags.HasError() {
		return model, true, diags
	}

	if status.Status == entityStoreStatusNotInstalled {
		return model, false, diags
	}

	entityTypes, started, logExtraction, flattenDiags := flattenStatus(ctx, status.Engines)
	if flattenDiags.HasError() {
		return model, true, flattenDiags
	}

	model.ID = clients.CompositeIDValue(spaceID, resourceID)
	model.SpaceID = types.StringValue(spaceID)
	model.EntityTypes = entityTypes
	model.Started = types.BoolValue(started)
	model.LogExtraction = logExtraction
	model.StatusJSON = NewStatusJSONValue(string(rawBody))
	return model, true, append(diags, flattenDiags...)
}

func diffEntityTypes(ctx context.Context, prior, plan types.Set) (added, removed []string, diags diag.Diagnostics) {
	var priorVals, planVals []string
	if typeutils.IsKnown(prior) {
		diags.Append(prior.ElementsAs(ctx, &priorVals, false)...)
	}
	if typeutils.IsKnown(plan) {
		diags.Append(plan.ElementsAs(ctx, &planVals, false)...)
	}
	if diags.HasError() {
		return nil, nil, diags
	}

	added, removed = typeutils.DiffStringSlices(priorVals, planVals)
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
		if e.Status == entityStoreEngineStatusStarted {
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
		DocsLimit:                   types.Int64Null(),
		FieldHistoryLength:          types.Int64Null(),
		Frequency:                   types.StringNull(),
		LookbackPeriod:              types.StringNull(),
		MaxLogsPerPage:              types.Int64Null(),
		MaxLogsPerWindow:            types.Int64Null(),
		MaxLogsPerWindowCapBehavior: types.StringNull(),
		MaxTimeWindowSize:           types.StringNull(),

		Delay: typeutils.StringishPointerValue(first.Delay)}
	if first.FieldHistoryLength != 0 {
		leModel.FieldHistoryLength = types.Int64Value(int64(first.FieldHistoryLength))
	}
	leModel.Frequency = typeutils.StringishPointerValue(first.Frequency)
	leModel.LookbackPeriod = typeutils.StringishPointerValue(first.LookbackPeriod)

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

			Delay:          typeutils.StringishPointerValue(e.Delay),
			Frequency:      typeutils.StringishPointerValue(e.Frequency),
			LookbackPeriod: typeutils.StringishPointerValue(e.LookbackPeriod),
			Filter:         typeutils.StringishPointerValue(e.Filter),
			Timeout:        typeutils.StringishPointerValue(e.Timeout),
			TimestampField: typeutils.StringishPointerValue(e.TimestampField)}
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
					ID:        types.StringValue(c.ID),
					Installed: types.BoolValue(c.Installed),
					Resource:  types.StringValue(c.Resource),
				}
				if c.Health != nil {
					cm.Health = types.StringValue(*c.Health)
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
