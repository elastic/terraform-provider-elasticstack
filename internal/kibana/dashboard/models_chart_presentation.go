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

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// models.LensChartPresentationTFModel mirrors optional chart-root presentation fields on typed Lens configs.

func newNullLensChartPresentationTFModel() models.LensChartPresentationTFModel {
	return models.LensChartPresentationTFModel{
		TimeRange:      nil,
		HideTitle:      types.BoolNull(),
		HideBorder:     types.BoolNull(),
		ReferencesJSON: jsontypes.NewNormalizedNull(),
		Drilldowns:     nil,
	}
}

type lensChartPresentationWrites = lenscommon.LensChartPresentationWrites

func lensChartPresentationWritesFor(dashboard *models.DashboardModel, in models.LensChartPresentationTFModel) (lensChartPresentationWrites, diag.Diagnostics) {
	return lenscommon.LensChartPresentationWritesFor(lensChartResolver(dashboard), in)
}

func lensChartPresentationReadsFor(
	ctx context.Context,
	dashboard *models.DashboardModel,
	prior *models.LensChartPresentationTFModel,
	apiTimeRange kbapi.KbnEsQueryServerTimeRangeSchema,
	hideTitle *bool,
	hideBorder *bool,
	refs *[]kbapi.KbnContentManagementUtilsReferenceSchema,
	drilldownWire [][]byte,
	drilldownsOmitted bool,
) (models.LensChartPresentationTFModel, diag.Diagnostics) {
	return lenscommon.LensChartPresentationReadsFor(ctx, lensChartResolver(dashboard), prior, apiTimeRange, hideTitle, hideBorder, refs, drilldownWire, drilldownsOmitted)
}

func decodeLensDrilldownSlice[Item any](raw [][]byte) ([]Item, diag.Diagnostics) {
	return lenscommon.DecodeLensDrilldownSlice[Item](raw)
}

func lensDrilldownsAPIToWire[Item any](items *[]Item) (wire [][]byte, omitted bool, diags diag.Diagnostics) {
	return lenscommon.LensDrilldownsAPIToWire(items)
}

// Backwards-compatible names for tests and legacy call sites (implementation lives in lenscommon).
const lensDrilldownTriggerOnApplyFilter = lenscommon.LensDrilldownTriggerOnApplyFilter

func lensDrilldownItemToRawJSON(item models.LensDrilldownItemTFModel, index int) ([]byte, diag.Diagnostics) {
	return lenscommon.LensDrilldownItemToRawJSON(item, index)
}

func lensDrilldownItemFromAPIJSON(raw []byte) (models.LensDrilldownItemTFModel, diag.Diagnostics) {
	return lenscommon.LensDrilldownItemFromAPIJSON(raw, "drilldowns[0]")
}

func lensDrilldownsToRawJSON(items []models.LensDrilldownItemTFModel) ([][]byte, diag.Diagnostics) {
	return lenscommon.LensDrilldownsToRawJSON(items)
}

func lensTimeRangesAPILiteralEqual(a, b kbapi.KbnEsQueryServerTimeRangeSchema) bool {
	return lenscommon.LensTimeRangesAPILiteralEqual(a, b)
}
