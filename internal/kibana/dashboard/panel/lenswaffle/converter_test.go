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

package lenswaffle

import (
	"context"
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
	require.Equal(t, string(kbapi.WaffleNoESQLTypeWaffle), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		WaffleConfig: &models.WaffleConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()
	var c converter
	resolver := stubResolver{}

	apiJSON := `{
		"type": "waffle",
		"title": "Waffle NoESQL Round-Trip",
		"description": "test",
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":""},
		"legend": {"size":"medium","visible":"auto"},
		"metrics": [{"operation":"count"}],
		"group_by": [{"operation":"terms","field":"host.name","collapse_by":"avg"}]
	}`
	var waffle kbapi.WaffleNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &waffle))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromWaffleNoESQL(waffle))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, blocks, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	require.NotNil(t, blocks.WaffleConfig)

	attrs2, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%s", diags)

	noESQL2, err := attrs2.AsWaffleNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Waffle NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.WaffleNoESQLTypeWaffle, noESQL2.Type)
	require.Len(t, noESQL2.Metrics, 1)
}
