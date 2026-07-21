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

package lenscommon

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EsqlGroupByAPIFields holds the common fields present in every ES|QL group-by API element
// (treemap, mosaic, waffle, pie). It acts as a neutral adapter between the package-local
// anonymous kbapi struct slices and the shared conversion helpers below.
type EsqlGroupByAPIFields struct {
	CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy   `json:"collapse_by,omitempty"`
	Color      *kbapi.KibanaHTTPAPIsColorMapping `json:"color,omitempty"`
	Column     string                            `json:"column"`
	Format     *kbapi.KibanaHTTPAPIsFormatType   `json:"format,omitempty"`
	Label      *string                           `json:"label,omitempty"`
}

// BuildEsqlGroupBySliceForAPI converts a []EsqlGroupByAPIFields into a []T by marshaling
// the entries to JSON and unmarshaling into the target slice type. T must be a struct whose
// JSON-tagged fields match the EsqlGroupByAPIFields JSON tags (collapse_by, color, column,
// format, label). Use this together with BuildPartitionEsqlGroupByForAPI to eliminate the
// repeated anonymous-struct copy loops across mosaic, treemap, and similar panels.
func BuildEsqlGroupBySliceForAPI[T any](entries []EsqlGroupByAPIFields, diags *diag.Diagnostics) []T {
	data, err := json.Marshal(entries)
	if err != nil {
		diags.AddError("Failed to marshal esql group_by entries", err.Error())
		return nil
	}
	var out []T
	if err := json.Unmarshal(data, &out); err != nil {
		diags.AddError("Failed to unmarshal esql group_by entries", err.Error())
		return nil
	}
	return out
}

// PopulatePartitionEsqlGroupByFromAPI converts a slice of EsqlGroupByAPIFields into a
// []models.PartitionEsqlGroupByModel. It is the single authoritative implementation of
// the from-API ES|QL group-by mapping shared by treemap and mosaic.
func PopulatePartitionEsqlGroupByFromAPI(src []EsqlGroupByAPIFields, diags *diag.Diagnostics) []models.PartitionEsqlGroupByModel {
	out := make([]models.PartitionEsqlGroupByModel, len(src))
	for i, gb := range src {
		collapseBy := ""
		if gb.CollapseBy != nil {
			collapseBy = string(*gb.CollapseBy)
		}
		out[i].Column = types.StringValue(gb.Column)
		out[i].CollapseBy = types.StringValue(collapseBy)

		colorBytes, err := json.Marshal(gb.Color)
		if err != nil {
			diags.AddError("Failed to marshal esql group_by color", err.Error())
			continue
		}
		out[i].ColorJSON = jsontypes.NewNormalizedValue(string(colorBytes))

		formatBytes, err := json.Marshal(gb.Format)
		if err != nil {
			diags.AddError("Failed to marshal esql group_by format", err.Error())
			continue
		}
		out[i].FormatJSON = jsontypes.NewNormalizedValue(string(formatBytes))

		out[i].Label = typeutils.StringishPointerValue(gb.Label)
	}
	return out
}

// BuildPartitionEsqlGroupByForAPI converts a []models.PartitionEsqlGroupByModel into a
// []EsqlGroupByAPIFields. It is the single authoritative implementation of the to-API
// ES|QL group-by mapping shared by treemap and mosaic.
func BuildPartitionEsqlGroupByForAPI(src []models.PartitionEsqlGroupByModel, diags *diag.Diagnostics) []EsqlGroupByAPIFields {
	out := make([]EsqlGroupByAPIFields, len(src))
	for i, eg := range src {
		out[i].Column = eg.Column.ValueString()
		collapseBy := kbapi.KibanaHTTPAPIsCollapseBy(eg.CollapseBy.ValueString())
		out[i].CollapseBy = &collapseBy

		var color kbapi.KibanaHTTPAPIsColorMapping
		if err := json.Unmarshal([]byte(eg.ColorJSON.ValueString()), &color); err != nil {
			diags.AddError("Failed to unmarshal esql group_by color_json", err.Error())
			return out
		}
		out[i].Color = &color

		formatSrc := DefaultLensNumberFormatJSON
		if typeutils.IsKnown(eg.FormatJSON) {
			formatSrc = eg.FormatJSON.ValueString()
		}
		var format kbapi.KibanaHTTPAPIsFormatType
		if err := json.Unmarshal([]byte(formatSrc), &format); err != nil {
			diags.AddError("Failed to unmarshal esql group_by format_json", err.Error())
			return out
		}
		out[i].Format = &format

		if typeutils.IsKnown(eg.Label) {
			l := eg.Label.ValueString()
			out[i].Label = &l
		}
	}
	return out
}
