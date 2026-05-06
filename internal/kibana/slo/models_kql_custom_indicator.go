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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfKqlKqlObjectAttrTypes matches the filter_kql / good_kql / total_kql SingleNestedAttribute schema
// (kql_query + list of filter rows with JSON query). Using types.Object holds unknown/null safely.
var (
	tfKqlFilterRowObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
		"query": jsontypes.NormalizedType{},
	}}
	tfKqlKqlObjectAttrTypes = map[string]attr.Type{
		"kql_query": types.StringType,
		"filters":   types.ListType{ElemType: tfKqlFilterRowObjectType},
	}
)

type tfKqlCustomIndicator struct {
	Index          types.String `tfsdk:"index"`
	DataViewID     types.String `tfsdk:"data_view_id"`
	Filter         types.String `tfsdk:"filter"`
	FilterKql      types.Object `tfsdk:"filter_kql"`
	Good           types.String `tfsdk:"good"`
	GoodKql        types.Object `tfsdk:"good_kql"`
	Total          types.String `tfsdk:"total"`
	TotalKql       types.Object `tfsdk:"total_kql"`
	TimestampField types.String `tfsdk:"timestamp_field"`
}

func (m tfModel) kqlCustomIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.KqlCustomIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.KqlCustomIndicator[0]

	var filterObj *kbapi.SLOsKqlWithFilters
	if typeutils.IsKnown(ind.FilterKql) && !ind.FilterKql.IsNull() {
		f, fDiags := kqlTFFormToKqlWithFiltersUnion(ind.FilterKql, "kql_custom_indicator.filter_kql")
		diags.Append(fDiags...)
		if diags.HasError() {
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		filterObj = &f
	} else if typeutils.IsKnown(ind.Filter) {
		v := ind.Filter.ValueString()
		var f kbapi.SLOsKqlWithFilters
		if err := f.FromSLOsKqlWithFilters0(v); err != nil {
			diags.AddError("Invalid configuration", "kql_custom_indicator.filter: "+err.Error())
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		filterObj = &f
	}

	// Default good and total to empty string if not provided, as they are required by the API.
	var good kbapi.SLOsKqlWithFiltersGood
	if typeutils.IsKnown(ind.GoodKql) && !ind.GoodKql.IsNull() {
		g, gDiags := kqlTFFormToKqlWithFiltersGoodUnion(ind.GoodKql, "kql_custom_indicator.good_kql")
		diags.Append(gDiags...)
		if diags.HasError() {
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		good = g
	} else {
		goodStr := ""
		if typeutils.IsKnown(ind.Good) {
			goodStr = ind.Good.ValueString()
		}
		if err := good.FromSLOsKqlWithFiltersGood0(goodStr); err != nil {
			diags.AddError("Invalid configuration", "kql_custom_indicator.good: "+err.Error())
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
	}

	var total kbapi.SLOsKqlWithFiltersTotal
	if typeutils.IsKnown(ind.TotalKql) && !ind.TotalKql.IsNull() {
		t, tDiags := kqlTFFormToKqlWithFiltersTotalUnion(ind.TotalKql, "kql_custom_indicator.total_kql")
		diags.Append(tDiags...)
		if diags.HasError() {
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		total = t
	} else {
		totalStr := ""
		if typeutils.IsKnown(ind.Total) {
			totalStr = ind.Total.ValueString()
		}
		if err := total.FromSLOsKqlWithFiltersTotal0(totalStr); err != nil {
			diags.AddError("Invalid configuration", "kql_custom_indicator.total: "+err.Error())
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
	}

	kqlIndicator := kbapi.SLOsIndicatorPropertiesCustomKql{
		Type: indicatorAddressToType["kql_custom_indicator"],
		Params: struct {
			DataViewId     *string                       `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
			Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
			Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
			Index          string                        `json:"index"`
			TimestampField string                        `json:"timestampField"`
			Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
		}{
			Index:          ind.Index.ValueString(),
			DataViewId:     typeutils.ValueStringPointer(ind.DataViewID),
			Filter:         filterObj,
			Good:           good,
			Total:          total,
			TimestampField: ind.TimestampField.ValueString(),
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesCustomKql(kqlIndicator); err != nil {
		diags.AddError("Failed to build KQL indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromKqlCustomIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesCustomKql) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p := apiIndicator.Params
	ind := tfKqlCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         types.StringNull(),
		FilterKql:      types.ObjectNull(tfKqlKqlObjectAttrTypes),
		Good:           types.StringNull(),
		GoodKql:        types.ObjectNull(tfKqlKqlObjectAttrTypes),
		Total:          types.StringNull(),
		TotalKql:       types.ObjectNull(tfKqlKqlObjectAttrTypes),
		DataViewID:     types.StringNull(),
	}

	if p.Filter != nil {
		s, k, kDiags := kqlWithFiltersAPIToTFFormFilter(p.Filter)
		diags.Append(kDiags...)
		ind.Filter = s
		ind.FilterKql = k
	}
	g, gK, gDiags := kqlWithFiltersAPIToTFFormGood(p.Good)
	diags.Append(gDiags...)
	ind.Good, ind.GoodKql = g, gK
	t, tK, tDiags := kqlWithFiltersAPIToTFFormTotal(p.Total)
	diags.Append(tDiags...)
	ind.Total, ind.TotalKql = t, tK

	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	m.KqlCustomIndicator = []tfKqlCustomIndicator{ind}
	return diags
}

func kqlTFFormToKqlWithFiltersUnion(obj types.Object, errPrefix string) (kbapi.SLOsKqlWithFilters, diag.Diagnostics) {
	var diags diag.Diagnostics
	one, d := kqlTFFormToAPI1(obj, errPrefix)
	diags.Append(d...)
	if diags.HasError() {
		return kbapi.SLOsKqlWithFilters{}, diags
	}
	var out kbapi.SLOsKqlWithFilters
	if err := out.FromSLOsKqlWithFilters1(one); err != nil {
		diags.AddError("Invalid configuration", errPrefix+": "+err.Error())
	}
	return out, diags
}

func kqlTFFormToKqlWithFiltersGoodUnion(obj types.Object, errPrefix string) (kbapi.SLOsKqlWithFiltersGood, diag.Diagnostics) {
	var diags diag.Diagnostics
	one, d := kqlTFFormToAPI1(obj, errPrefix)
	diags.Append(d...)
	if diags.HasError() {
		return kbapi.SLOsKqlWithFiltersGood{}, diags
	}
	g1 := kbapi.SLOsKqlWithFiltersGood1(one)
	var out kbapi.SLOsKqlWithFiltersGood
	if err := out.FromSLOsKqlWithFiltersGood1(g1); err != nil {
		diags.AddError("Invalid configuration", errPrefix+": "+err.Error())
	}
	return out, diags
}

func kqlTFFormToKqlWithFiltersTotalUnion(obj types.Object, errPrefix string) (kbapi.SLOsKqlWithFiltersTotal, diag.Diagnostics) {
	var diags diag.Diagnostics
	one, d := kqlTFFormToAPI1(obj, errPrefix)
	diags.Append(d...)
	if diags.HasError() {
		return kbapi.SLOsKqlWithFiltersTotal{}, diags
	}
	t1 := kbapi.SLOsKqlWithFiltersTotal1(one)
	var out kbapi.SLOsKqlWithFiltersTotal
	if err := out.FromSLOsKqlWithFiltersTotal1(t1); err != nil {
		diags.AddError("Invalid configuration", errPrefix+": "+err.Error())
	}
	return out, diags
}

// kqlTFFormToAPI1 maps Terraform object-form KQL to the kbapi object union arm (KqlQuery + Filters).
func kqlTFFormToAPI1(obj types.Object, errPrefix string) (kbapi.SLOsKqlWithFilters1, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := obj.Attributes()
	var kq *string
	if k, ok := attrs["kql_query"].(types.String); ok && typeutils.IsKnown(k) {
		if s := k.ValueString(); s != "" {
			kq = &s
		}
	}
	var filters *[]kbapi.SLOsFilter
	fl, ok := attrs["filters"].(types.List)
	if ok && typeutils.IsKnown(fl) && !fl.IsNull() {
		elems := fl.Elements()
		if len(elems) > 0 {
			rows := make([]kbapi.SLOsFilter, 0, len(elems))
			for i, e := range elems {
				rowObj, ok := e.(types.Object)
				if !ok {
					diags.AddError("Invalid configuration", errPrefix+fmt.Sprintf(".filters[%d]", i)+": expected object row")
					continue
				}
				qv, has := rowObj.Attributes()["query"]
				if !has {
					diags.AddError("Invalid configuration", errPrefix+fmt.Sprintf(".filters[%d]", i)+": missing query")
					continue
				}
				n, ok := qv.(jsontypes.Normalized)
				if !ok {
					diags.AddError("Invalid configuration", errPrefix+fmt.Sprintf(".filters[%d]", i)+": query is not JSON")
					continue
				}
				// A known null filter query is not the same as unknown: IsKnown is false for
				// both, so we must inspect null/unknown on the value directly.
				if n.IsUnknown() {
					diags.AddError("Invalid configuration", fmt.Sprintf("%s.filters[%d].query: value is not yet known", errPrefix, i))
					continue
				}
				if n.IsNull() {
					diags.AddError("Invalid configuration", fmt.Sprintf("%s.filters[%d].query: a JSON object is required for a filter row in this list (null is not valid)", errPrefix, i))
					continue
				}
				qm := make(map[string]any)
				if err := json.Unmarshal([]byte(n.ValueString()), &qm); err != nil {
					diags.AddError("Invalid configuration", fmt.Sprintf("%s.filters[%d]: query JSON: %s", errPrefix, i, err))
					continue
				}
				rows = append(rows, kbapi.SLOsFilter{Query: &qm})
			}
			if diags.HasError() {
				return kbapi.SLOsKqlWithFilters1{}, diags
			}
			filters = &rows
		}
	}
	// Allow object form with only filters (KqlQuery optional) or only kql_query.
	return kbapi.SLOsKqlWithFilters1{KqlQuery: kq, Filters: filters}, diags
}

func kqlAPIFilterRowToTF(f kbapi.SLOsFilter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if f.Query == nil {
		qn := jsontypes.NewNormalizedNull()
		row, oDiags := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{
			"query": qn,
		})
		diags.Append(oDiags...)
		return row, diags
	}
	b, err := json.Marshal(*f.Query)
	if err != nil {
		diags.AddError("Unexpected API response", "SLO filter query: "+err.Error())
		return types.ObjectNull(tfKqlFilterRowObjectType.AttrTypes), diags
	}
	qn := jsontypes.NewNormalizedValue(string(b))
	row, oDiags := types.ObjectValue(tfKqlFilterRowObjectType.AttrTypes, map[string]attr.Value{
		"query": qn,
	})
	diags.Append(oDiags...)
	return row, diags
}

func kqlObjectShapeAPIToTF(kq *string, filters *[]kbapi.SLOsFilter) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	var kqlStr types.String
	if kq != nil {
		kqlStr = types.StringValue(*kq)
	} else {
		kqlStr = types.StringNull()
	}
	var listVal types.List
	if filters == nil || len(*filters) == 0 {
		empty, lDiags := types.ListValue(tfKqlFilterRowObjectType, []attr.Value{})
		diags.Append(lDiags...)
		listVal = empty
	} else {
		rows := make([]attr.Value, 0, len(*filters))
		for i := range *filters {
			row, d := kqlAPIFilterRowToTF((*filters)[i])
			diags.Append(d...)
			if diags.HasError() {
				return types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
			}
			rows = append(rows, row)
		}
		lv, lDiags := types.ListValue(tfKqlFilterRowObjectType, rows)
		diags.Append(lDiags...)
		listVal = lv
	}
	obj, oDiags := types.ObjectValue(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
		"kql_query": kqlStr,
		"filters":   listVal,
	})
	diags.Append(oDiags...)
	return obj, diags
}

func kqlObjectAPIToTFFormIfRich(o kbapi.SLOsKqlWithFilters1) (types.String, types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	hasFilters := o.Filters != nil && len(*o.Filters) > 0
	if hasFilters {
		obj, d := kqlObjectShapeAPIToTF(o.KqlQuery, o.Filters)
		diags.Append(d...)
		if diags.HasError() {
			return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
		}
		return types.StringNull(), obj, diags
	}
	if o.KqlQuery != nil {
		return types.StringValue(*o.KqlQuery), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
	}
	return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
}

// kqlWithFiltersAPIToTFFormFilter maps API *SLOsKqlWithFilters to legacy string and/or _kql object in Terraform state.
func kqlWithFiltersAPIToTFFormFilter(union *kbapi.SLOsKqlWithFilters) (types.String, types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if union == nil {
		return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
	}
	if o, err := union.AsSLOsKqlWithFilters1(); err == nil {
		return kqlObjectAPIToTFFormIfRich(o)
	}
	if s, err := union.AsSLOsKqlWithFilters0(); err == nil {
		return types.StringValue(s), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
	}
	return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diags
}

func kqlWithFiltersAPIToTFFormGood(union kbapi.SLOsKqlWithFiltersGood) (types.String, types.Object, diag.Diagnostics) {
	if o, err := union.AsSLOsKqlWithFiltersGood1(); err == nil {
		return kqlObjectAPIToTFFormIfRich(kbapi.SLOsKqlWithFilters1(o))
	}
	if s, err := union.AsSLOsKqlWithFiltersGood0(); err == nil {
		return types.StringValue(s), types.ObjectNull(tfKqlKqlObjectAttrTypes), diag.Diagnostics{}
	}
	return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diag.Diagnostics{}
}

func kqlWithFiltersAPIToTFFormTotal(union kbapi.SLOsKqlWithFiltersTotal) (types.String, types.Object, diag.Diagnostics) {
	if o, err := union.AsSLOsKqlWithFiltersTotal1(); err == nil {
		return kqlObjectAPIToTFFormIfRich(kbapi.SLOsKqlWithFilters1(o))
	}
	if s, err := union.AsSLOsKqlWithFiltersTotal0(); err == nil {
		return types.StringValue(s), types.ObjectNull(tfKqlKqlObjectAttrTypes), diag.Diagnostics{}
	}
	return types.StringNull(), types.ObjectNull(tfKqlKqlObjectAttrTypes), diag.Diagnostics{}
}
