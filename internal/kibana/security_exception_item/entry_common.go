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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// entrySink is implemented by *EntryModel and *NestedEntryModel so the "from API"
// conversions for single- and multi-value entry kinds (match, wildcard, match_any)
// can be written once and shared between the flat and nested entry representations
// instead of duplicated per model type.
type entrySink interface {
	setValue(v types.String)
	setValues(v types.List)
	resetValueFields(active string)
}

func (e *EntryModel) setValue(v types.String)        { e.Value = v }
func (e *EntryModel) setValues(v types.List)         { e.Values = v }
func (e *EntryModel) resetValueFields(active string) { resetEntryModelFields(e, active) }

func (e *NestedEntryModel) setValue(v types.String)        { e.Value = v }
func (e *NestedEntryModel) setValues(v types.List)         { e.Values = v }
func (e *NestedEntryModel) resetValueFields(active string) { resetNestedEntryModelFields(e, active) }

// convertSingleValueEntryFromAPI sets Value from entryMap["value"] on any entrySink.
// Shared by the flat match/wildcard and nested match "from API" conversions.
func convertSingleValueEntryFromAPI(entryMap map[string]any, entry entrySink) {
	entry.resetValueFields(attrValue)
	if value, ok := entryMap["value"].(string); ok {
		entry.setValue(types.StringValue(value))
	} else {
		entry.setValue(types.StringNull())
	}
}

// convertMultiValueEntryFromAPI sets Values from entryMap["value"] on any entrySink.
// Shared by the flat and nested match_any "from API" conversions.
func convertMultiValueEntryFromAPI(ctx context.Context, entryMap map[string]any, entry entrySink) diag.Diagnostics {
	var diags diag.Diagnostics
	entry.resetValueFields(attrValues)

	if values, ok := entryMap["value"].([]any); ok {
		strValues := make([]string, 0, len(values))
		for _, v := range values {
			if str, ok := v.(string); ok {
				strValues = append(strValues, str)
			}
		}
		list, d := types.ListValueFrom(ctx, types.StringType, strValues)
		diags.Append(d...)
		entry.setValues(list)
	} else {
		entry.setValues(types.ListNull(types.StringType))
	}
	return diags
}

// matchAPIEntry is implemented by the generated "entry" union types
// (SecurityExceptionsAPIExceptionListItemEntry and its nested-entry-item
// counterpart) via their FromSecurityExceptionsAPIExceptionListItemEntryMatch setter.
type matchAPIEntry interface {
	FromSecurityExceptionsAPIExceptionListItemEntryMatch(v kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch) error
}

// fillMatchEntry builds a "match" API entry from value/field/operator into result.
// Shared by the flat and nested "match" entry "to API" conversions.
func fillMatchEntry[R matchAPIEntry](
	result R,
	value types.String,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
	missingValueMsg string,
	failureMsg string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(value) || value.ValueString() == "" {
		diags.AddError("Invalid Configuration", missingValueMsg)
		return diags
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatch{
		Type:     entryTypeMatch,
		Field:    field,
		Operator: operator,
		Value:    value.ValueString(),
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatch(apiEntry); err != nil {
		diags.AddError(failureMsg, err.Error())
	}
	return diags
}

// matchAnyAPIEntry is implemented by the generated "entry" union types via their
// FromSecurityExceptionsAPIExceptionListItemEntryMatchAny setter.
type matchAnyAPIEntry interface {
	FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(v kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny) error
}

// fillMatchAnyEntry builds a "match_any" API entry from values/field/operator into
// result. Shared by the flat and nested "match_any" entry "to API" conversions.
func fillMatchAnyEntry[R matchAnyAPIEntry](
	ctx context.Context,
	result R,
	values types.List,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
	missingValueMsg string,
	emptyValueMsg string,
	failureMsg string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(values) {
		diags.AddError("Invalid Configuration", missingValueMsg)
		return diags
	}

	strValues := typeutils.ListTypeAs[string](ctx, values, path.Empty(), &diags)
	if diags.HasError() {
		return diags
	}

	if len(strValues) == 0 {
		diags.AddError("Invalid Configuration", emptyValueMsg)
		return diags
	}

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryMatchAny{
		Type:     entryTypeMatchAny,
		Field:    field,
		Operator: operator,
		Value:    strValues,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryMatchAny(apiEntry); err != nil {
		diags.AddError(failureMsg, err.Error())
	}
	return diags
}

// existsAPIEntry is implemented by the generated "entry" union types via their
// FromSecurityExceptionsAPIExceptionListItemEntryExists setter.
type existsAPIEntry interface {
	FromSecurityExceptionsAPIExceptionListItemEntryExists(v kbapi.SecurityExceptionsAPIExceptionListItemEntryExists) error
}

// fillExistsEntry builds an "exists" API entry from field/operator into result.
// Shared by the flat and nested "exists" entry "to API" conversions.
func fillExistsEntry[R existsAPIEntry](
	result R,
	field kbapi.SecurityExceptionsAPINonEmptyString,
	operator kbapi.SecurityExceptionsAPIExceptionListItemEntryOperator,
	failureMsg string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	apiEntry := kbapi.SecurityExceptionsAPIExceptionListItemEntryExists{
		Type:     entryTypeExists,
		Field:    field,
		Operator: operator,
	}
	if err := result.FromSecurityExceptionsAPIExceptionListItemEntryExists(apiEntry); err != nil {
		diags.AddError(failureMsg, err.Error())
	}
	return diags
}
