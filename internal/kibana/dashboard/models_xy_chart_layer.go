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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensxy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func xyLayerFromAPILayersNoESQL(ctx context.Context, m *models.XYLayerModel, apiLayer kbapi.XyLayersNoESQL) diag.Diagnostics {
	return lensxy.LayerFromAPILayersNoESQL(ctx, m, apiLayer)
}

func xyLayerToAPILayersNoESQL(m *models.XYLayerModel) (kbapi.XyLayersNoESQL, diag.Diagnostics) {
	return lensxy.LayerToAPILayersNoESQL(m)
}

func xyLayerToAPILayerESQL(m *models.XYLayerModel) (kbapi.XyLayerESQL, diag.Diagnostics) {
	return lensxy.LayerToAPILayerESQL(m)
}

func thresholdFromAPIJSON(m *models.ThresholdModel, jsonData []byte) diag.Diagnostics {
	return lensxy.ThresholdFromAPIJSON(m, jsonData)
}

func thresholdToAPI(m *models.ThresholdModel) (map[string]any, diag.Diagnostics) {
	return lensxy.ThresholdToAPI(m)
}

func dataLayerFromAPINoESQL(ctx context.Context, m *models.DataLayerModel, apiLayer kbapi.XyLayerNoESQL) diag.Diagnostics {
	return lensxy.DataLayerFromAPINoESQL(ctx, m, apiLayer)
}

func dataLayerToAPIXyLayerNoESQL(m *models.DataLayerModel, layerType string) (kbapi.XyLayerNoESQL, diag.Diagnostics) {
	return lensxy.DataLayerToAPIXyLayerNoESQL(m, layerType)
}

func dataLayerFromAPIESql(ctx context.Context, m *models.DataLayerModel, apiLayer kbapi.XyLayerESQL) diag.Diagnostics {
	return lensxy.DataLayerFromAPIESql(ctx, m, apiLayer)
}

func dataLayerToAPIXyLayerESQL(m *models.DataLayerModel, layerType string) (kbapi.XyLayerESQL, diag.Diagnostics) {
	return lensxy.DataLayerToAPIXyLayerESQL(m, layerType)
}

func referenceLineLayerFromAPINoESQL(m *models.ReferenceLineLayerModel, apiLayer kbapi.XyReferenceLineLayerNoESQL) diag.Diagnostics {
	return lensxy.ReferenceLineLayerFromAPINoESQL(m, apiLayer)
}

func referenceLineLayerToAPIXyReferenceLineLayerNoESQL(m *models.ReferenceLineLayerModel, layerType string) (kbapi.XyReferenceLineLayerNoESQL, diag.Diagnostics) {
	return lensxy.ReferenceLineLayerToAPIXyReferenceLineLayerNoESQL(m, layerType)
}
