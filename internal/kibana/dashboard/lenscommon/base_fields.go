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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LensChartBaseFieldsForAPI extracts the four common base fields from a LensChartBaseTFModel and
// returns them as pointer values suitable for assignment to any Lens API struct that carries
// Title, Description, IgnoreGlobalFilters, and Sampling. This is the write-direction counterpart
// of PopulateLensChartBaseFromAPI.
func LensChartBaseFieldsForAPI(m models.LensChartBaseTFModel) (title, description *string, ignoreGlobalFilters *bool, sampling *float32) {
	if typeutils.IsKnown(m.Title) {
		title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		ignoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.Sampling) {
		s := float32(m.Sampling.ValueFloat64())
		sampling = &s
	}
	return
}

// PopulateLensChartBaseFromAPI returns a LensChartBaseTFModel populated from the common API parameters.
// Returns false (and appends to diags) when any field fails to populate.
func PopulateLensChartBaseFromAPI(
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	dataSourceJSONFieldName string,
	filters *kbapi.KibanaHTTPAPIsLensPanelFilters,
	diags *diag.Diagnostics,
) (models.LensChartBaseTFModel, bool) {
	base := models.LensChartBaseTFModel{
		Title:               types.StringPointerValue(title),
		Description:         types.StringPointerValue(description),
		IgnoreGlobalFilters: types.BoolPointerValue(ignoreGlobalFilters),
	}
	if sampling != nil {
		base.Sampling = types.Float64Value(float64(*sampling))
	} else {
		base.Sampling = types.Float64Null()
	}
	dv, ok := WrapNormalizedJSON(datasetBytes, datasetErr, dataSourceJSONFieldName, diags)
	if !ok {
		return base, false
	}
	base.DataSourceJSON = dv
	base.Filters = PopulateFiltersFromAPI(filters, diags)
	return base, !diags.HasError()
}

// PopulateLensChartBaseAndQueryFromAPI wraps PopulateLensChartBaseFromAPI and additionally
// toggles the shared optional FilterSimple "query" block, consolidating the ESQL/NoESQL split
// repeated at the top of every Lens chart FromAPI pair: marshal the dataset, populate the base
// fields, then either clear the query (ESQL, which carries no query field on the wire) or
// populate it from the API's query (NoESQL). dataset is marshaled with json.Marshal, which
// dispatches to the concrete type's MarshalJSON when present - as it is for every NoESQL union
// DataSource type - so a single call covers both variants despite some hand-written call sites
// historically calling dataset.MarshalJSON() directly. Returns false (having appended diags) when
// base population fails; callers should return immediately.
func PopulateLensChartBaseAndQueryFromAPI(
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	dataset any,
	dataSourceJSONFieldName string,
	filters *kbapi.KibanaHTTPAPIsLensPanelFilters,
	query **models.FilterSimpleModel,
	isESQL bool,
	apiQuery *kbapi.KibanaHTTPAPIsFilterSimple,
	diags *diag.Diagnostics,
) (models.LensChartBaseTFModel, bool) {
	datasetBytes, datasetErr := json.Marshal(dataset)
	base, ok := PopulateLensChartBaseFromAPI(
		title, description, ignoreGlobalFilters, sampling,
		datasetBytes, datasetErr, dataSourceJSONFieldName, filters, diags,
	)
	if !ok {
		return base, false
	}

	if isESQL {
		*query = nil
	} else {
		*query = &models.FilterSimpleModel{}
		FilterSimpleFromAPI(*query, apiQuery)
	}

	return base, true
}
