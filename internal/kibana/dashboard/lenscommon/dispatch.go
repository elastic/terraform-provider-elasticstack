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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DispatchByQueryMode handles the ESQL/NoESQL dispatcher boilerplate shared across lens panel
// ConfigToAPI functions. It calls the appropriate builder, appends its diagnostics, and applies
// the result to a LensApiConfig via the provided From* setter. Nil-guard and any
// panel-specific pre-validation must be done by the caller before invoking this function.
func DispatchByQueryMode[ESQL, NoESQL any](
	usesESQL bool,
	buildESQL func() (ESQL, diag.Diagnostics),
	applyESQL func(*kbapi.KibanaHTTPAPIsLensApiConfig, ESQL) error,
	esqlErrSummary string,
	buildNoESQL func() (NoESQL, diag.Diagnostics),
	applyNoESQL func(*kbapi.KibanaHTTPAPIsLensApiConfig, NoESQL) error,
	noESQLErrSummary string,
) (kbapi.KibanaHTTPAPIsLensApiConfig, diag.Diagnostics) {
	var attrs kbapi.KibanaHTTPAPIsLensApiConfig
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

// SnapshotAndResetBlock captures a copy of the block currently pointed to by field (or
// nil if it is unset) as "prior" state, then resets *field to a fresh zero-value block
// so PopulateFromAttributes implementations can populate it from scratch while their
// FromAPI helpers still have access to the previous state for value preservation.
func SnapshotAndResetBlock[TBlock any](field **TBlock) *TBlock {
	var prior *TBlock
	if *field != nil {
		cpy := **field
		prior = &cpy
	}
	*field = new(TBlock)
	return prior
}

// PopulateFromNoESQLOrESQL handles the "try the NoESQL variant, optionally guarded by a
// NoESQL-candidate-is-actually-ESQL sniff, otherwise fall back to the ESQL variant"
// dispatch shared by every by-value lens chart converter's PopulateFromAttributes.
//
// guard may be nil when the chart's NoESQL API shape has no single top-level data
// source to sniff (e.g. xy charts, where each layer carries its own data source) - in
// that case only the NoESQL decode error determines the branch.
func PopulateFromNoESQLOrESQL[TConfig, TNoESQL, TESQL any](
	ctx context.Context,
	config *TConfig,
	prior *TConfig,
	asNoESQL func() (TNoESQL, error),
	asESQL func() (TESQL, error),
	guard func(TNoESQL) bool,
	fromAPINoESQL func(context.Context, *TConfig, *TConfig, TNoESQL) diag.Diagnostics,
	fromAPIESQL func(context.Context, *TConfig, *TConfig, TESQL) diag.Diagnostics,
) diag.Diagnostics {
	if noESQL, err := asNoESQL(); err == nil && (guard == nil || guard(noESQL)) {
		return fromAPINoESQL(ctx, config, prior, noESQL)
	}
	esql, err := asESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return fromAPIESQL(ctx, config, prior, esql)
}
