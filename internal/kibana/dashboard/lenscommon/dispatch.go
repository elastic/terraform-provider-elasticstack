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

import "github.com/hashicorp/terraform-plugin-framework/diag"

// DispatchByQueryMode handles the ESQL/NoESQL dispatcher boilerplate shared across lens panel
// ConfigToAPI functions. It calls the appropriate builder, appends its diagnostics, and applies
// the result to a VisByValueConfig0 via the provided From* setter. Nil-guard and any
// panel-specific pre-validation must be done by the caller before invoking this function.
func DispatchByQueryMode[ESQL, NoESQL any](
	usesESQL bool,
	buildESQL func() (ESQL, diag.Diagnostics),
	applyESQL func(*VisByValueConfig0, ESQL) error,
	esqlErrSummary string,
	buildNoESQL func() (NoESQL, diag.Diagnostics),
	applyNoESQL func(*VisByValueConfig0, NoESQL) error,
	noESQLErrSummary string,
) (VisByValueConfig0, diag.Diagnostics) {
	var attrs VisByValueConfig0
	var diags diag.Diagnostics

	if usesESQL {
		esql, esqlDiags := buildESQL()
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := applyESQL(&attrs, esql); err != nil {
			diags.AddError(esqlErrSummary, err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := buildNoESQL()
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := applyNoESQL(&attrs, noESQL); err != nil {
		diags.AddError(noESQLErrSummary, err.Error())
	}
	return attrs, diags
}
