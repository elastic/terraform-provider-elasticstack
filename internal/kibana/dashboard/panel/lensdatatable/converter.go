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

package lensdatatable

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func init() {
	lenscommon.Register(converter{})
}

type converter struct{}

func (converter) VizType() string {
	return string(kbapi.DatatableNoESQLTypeDataTable)
}

func (converter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	return blocks != nil && blocks.DatatableConfig != nil
}

func (converter) SchemaAttribute() schema.Attribute {
	return lenscommon.ByValueChartNestedAttribute("datatable_config", getDatatableSchema(true))
}

func (converter) PopulateFromAttributes(ctx context.Context, resolver lenscommon.Resolver, blocks *models.LensByValueChartBlocks, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var priorNo *models.DatatableNoESQLConfigModel
	var priorEsql *models.DatatableESQLConfigModel
	if blocks != nil && blocks.DatatableConfig != nil {
		if blocks.DatatableConfig.NoESQL != nil {
			cpy := *blocks.DatatableConfig.NoESQL
			priorNo = &cpy
		}
		if blocks.DatatableConfig.ESQL != nil {
			cpy := *blocks.DatatableConfig.ESQL
			priorEsql = &cpy
		}
	}
	if blocks == nil {
		var d diag.Diagnostics
		d.AddError("Lens chart blocks missing", "cannot populate datatable_config without chart blocks")
		return d
	}
	blocks.DatatableConfig = &models.DatatableConfigModel{}

	if datatableNoESQL, err := attrs.AsDatatableNoESQL(); err == nil && !isDatatableNoESQLCandidateActuallyESQL(datatableNoESQL) {
		blocks.DatatableConfig.NoESQL = &models.DatatableNoESQLConfigModel{}
		return datatableNoESQLConfigFromAPI(ctx, blocks.DatatableConfig.NoESQL, resolver, priorNo, datatableNoESQL)
	}
	datatableESQL, err := attrs.AsDatatableESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	blocks.DatatableConfig.ESQL = &models.DatatableESQLConfigModel{}
	return datatableESQLConfigFromAPI(ctx, blocks.DatatableConfig.ESQL, resolver, priorEsql, datatableESQL)
}

func (converter) BuildAttributes(blocks *models.LensByValueChartBlocks, resolver lenscommon.Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	if blocks == nil || blocks.DatatableConfig == nil {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	switch {
	case blocks.DatatableConfig.NoESQL != nil:
		noESQL, noDiags := datatableNoESQLConfigToAPI(blocks.DatatableConfig.NoESQL, resolver)
		diags.Append(noDiags...)
		if diags.HasError() {
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}

		if err := attrs.FromDatatableNoESQL(noESQL); err != nil {
			diags.AddError("Failed to convert datatable no-esql config", err.Error())
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}
	case blocks.DatatableConfig.ESQL != nil:
		esql, esqlDiags := datatableESQLConfigToAPI(blocks.DatatableConfig.ESQL, resolver)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}

		if err := attrs.FromDatatableESQL(esql); err != nil {
			diags.AddError("Failed to convert datatable esql config", err.Error())
			return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
		}
	default:
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

func (converter) AlignStateFromPlan(_ context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	if plan.DatatableConfig == nil || state.DatatableConfig == nil {
		return
	}
	alignDatatableStateFromPlan(plan.DatatableConfig, state.DatatableConfig)
}

func (converter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return populateDatatableLensAttributes(attrs)
}
