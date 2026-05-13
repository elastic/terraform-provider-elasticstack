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

package dashboard

// visByReferenceModel duplicates lensDashboardAppByReferenceModel — both branches use getLensByReferenceAttributes()
// in schema and identical API saved-object linkage fields on read/write (design D3).
type visByReferenceModel = lensDashboardAppByReferenceModel

// visByValueModel is Terraform model for vis_config.by_value (12 Lens chart kinds, no nested config_json; design D4).
type visByValueModel struct {
	lensByValueChartBlocks
}

// visConfigModel is nested `vis_config` on panels with type vis (design D10; mapPanelFromAPI / toAPI classification in task 6).
type visConfigModel struct {
	ByValue     *visByValueModel     `tfsdk:"by_value"`
	ByReference *visByReferenceModel `tfsdk:"by_reference"`
}
