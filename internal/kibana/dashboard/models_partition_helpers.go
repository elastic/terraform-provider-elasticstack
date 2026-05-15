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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapOptionalBoolWithSnapshotDefault maps an optional API bool to a Terraform Bool,
// preserving snapshot defaults (e.g. when the API returns false and the user hasn't set it).
//
//nolint:unparam // snapshotDefault is a parameter for API flexibility; callers use false for partition charts
func mapOptionalBoolWithSnapshotDefault(current types.Bool, apiValue *bool, snapshotDefault bool) types.Bool {
	return lenscommon.MapOptionalBoolWithSnapshotDefault(current, apiValue, snapshotDefault)
}

// mapOptionalFloatWithSnapshotDefault maps an optional API float to a Terraform Float64,
// preserving snapshot defaults.
//
//nolint:unparam // snapshotDefault is a parameter for API flexibility; callers use 1 for partition charts
func mapOptionalFloatWithSnapshotDefault(current types.Float64, apiValue *float32, snapshotDefault float64) types.Float64 {
	return lenscommon.MapOptionalFloatWithSnapshotDefault(current, apiValue, snapshotDefault)
}

// (treemap, mosaic, pie). Used by treemap, mosaic, and pie chart config models.

func partitionLegendFromTreemapLegend(m *models.PartitionLegendModel, api kbapi.TreemapLegend) {
	lenscommon.PartitionLegendFromTreemapLegend(m, api)
}

func partitionLegendFromMosaicLegend(m *models.PartitionLegendModel, api kbapi.MosaicLegend) {
	lenscommon.PartitionLegendFromMosaicLegend(m, api)
}

func partitionLegendToTreemapLegend(m *models.PartitionLegendModel) kbapi.TreemapLegend {
	return lenscommon.PartitionLegendToTreemapLegend(m)
}

func partitionLegendToMosaicLegend(m *models.PartitionLegendModel) kbapi.MosaicLegend {
	return lenscommon.PartitionLegendToMosaicLegend(m)
}

func partitionLegendFromPieLegend(m *models.PartitionLegendModel, api kbapi.PieLegend) {
	lenscommon.PartitionLegendFromPieLegend(m, api)
}

func partitionLegendToPieLegend(m *models.PartitionLegendModel) kbapi.PieLegend {
	return lenscommon.PartitionLegendToPieLegend(m)
}

func partitionValueDisplayFromValueDisplay(m *models.PartitionValueDisplay, api kbapi.ValueDisplay) {
	lenscommon.PartitionValueDisplayFromAPI(m, api)
}

func partitionValueDisplayToValueDisplay(m *models.PartitionValueDisplay) kbapi.ValueDisplay {
	return lenscommon.PartitionValueDisplayToAPI(m)
}

// newPartitionGroupByJSONFromAPI builds group_by / group_breakdown_by JSON for Terraform state from the API payload.
// Kibana may add explicit null fields on read; dropping them avoids "inconsistent result after apply" for ES|QL treemaps/mosaics.
// Terms defaults are not merged here (that would change round-trip panelsToAPI); JSONWithDefaultsValue still applies populatePartitionGroupByDefaults for semantic equality.
func newPartitionGroupByJSONFromAPI(apiPayload any) (customtypes.JSONWithDefaultsValue[[]map[string]any], diag.Diagnostics) {
	return lenscommon.NewPartitionGroupByJSONFromAPI(apiPayload)
}
