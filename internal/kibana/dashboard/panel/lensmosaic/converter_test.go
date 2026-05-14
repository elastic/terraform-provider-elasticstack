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

package lensmosaic

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResolver struct{}

func (stubResolver) ResolveChartTimeRange(chartLevel *models.TimeRangeModel) kbapi.KbnEsQueryServerTimeRangeSchema {
	_ = chartLevel
	return kbapi.KbnEsQueryServerTimeRangeSchema{}
}

func (stubResolver) DashboardLensComparableTimeRange() (kbapi.KbnEsQueryServerTimeRangeSchema, bool) {
	return kbapi.KbnEsQueryServerTimeRangeSchema{}, false
}

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.MosaicNoESQLTypeMosaic), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		MosaicConfig: &models.MosaicConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "Mosaic NoESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": true,
		"sampling": 0.5,
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":"status:200"},
		"legend": {"size": "medium"},
		"metrics": [{"operation":"count"}],
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromMosaicNoESQL(api))

	var c converter
	resolver := stubResolver{}
	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.MosaicConfig)

	attrs2, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	noESQL2, err := attrs2.AsMosaicNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Mosaic NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.MosaicNoESQLTypeMosaic, noESQL2.Type)
}
