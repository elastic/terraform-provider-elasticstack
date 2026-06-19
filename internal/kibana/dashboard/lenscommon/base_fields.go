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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LensChartBaseFields holds pointers to the common base fields shared by multiple Lens chart models.
// Pass a pointer to each field in the target model so PopulateLensChartBaseFromAPI writes to them directly.
type LensChartBaseFields struct {
	Title               *types.String
	Description         *types.String
	IgnoreGlobalFilters *types.Bool
	Sampling            *types.Float64
	DataSourceJSON      *jsontypes.Normalized
	Filters             *[]models.ChartFilterJSONModel
}

// PopulateLensChartBaseFromAPI writes the common Lens chart base fields from API parameters into
// the fields pointed to by f. Returns false (and appends to diags) when any field fails to populate.
func PopulateLensChartBaseFromAPI(
	f LensChartBaseFields,
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	dataSourceJSONFieldName string,
	filters *kbapi.KibanaHTTPAPIsLensPanelFilters,
	diags *diag.Diagnostics,
) bool {
	*f.Title = types.StringPointerValue(title)
	*f.Description = types.StringPointerValue(description)
	*f.IgnoreGlobalFilters = types.BoolPointerValue(ignoreGlobalFilters)
	if sampling != nil {
		*f.Sampling = types.Float64Value(float64(*sampling))
	} else {
		*f.Sampling = types.Float64Null()
	}
	dv, ok := MarshalToNormalized(datasetBytes, datasetErr, dataSourceJSONFieldName, diags)
	if !ok {
		return false
	}
	*f.DataSourceJSON = dv
	*f.Filters = PopulateFiltersFromAPI(filters, diags)
	return !diags.HasError()
}
