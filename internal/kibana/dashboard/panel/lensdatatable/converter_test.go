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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	require.Equal(t, string(kbapi.DatatableNoESQLTypeDataTable), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		DatatableConfig: &models.DatatableConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	noESQL := &models.DatatableNoESQLConfigModel{
		Title:               types.StringValue("Datatable RT"),
		Description:         types.StringValue("desc"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.5),
		DataSourceJSON:      jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`),
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue(`*`),
		},
		Styling: &models.DatatableStylingModel{
			Density: &models.DatatableDensityModel{
				Mode: types.StringValue(string(kbapi.DatatableDensityModeExpanded)),
			},
		},
		Metrics: []models.DatatableMetricModel{
			{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count"}`)},
		},
	}

	blocks := &models.LensByValueChartBlocks{
		DatatableConfig: &models.DatatableConfigModel{NoESQL: noESQL},
	}
	attrs, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	out := &models.LensByValueChartBlocks{DatatableConfig: &models.DatatableConfigModel{}}
	diags = c.PopulateFromAttributes(ctx, resolver, out, attrs)
	require.False(t, diags.HasError(), "%v", diags)

	require.Equal(t, noESQL.Title.ValueString(), out.DatatableConfig.NoESQL.Title.ValueString())
	require.Equal(t, noESQL.Query.Expression.ValueString(), out.DatatableConfig.NoESQL.Query.Expression.ValueString())
	require.Len(t, out.DatatableConfig.NoESQL.Metrics, 1)
}

func TestConverter_roundTrip_ESQL_datatable(t *testing.T) {
	ctx := t.Context()
	var c converter
	resolver := stubResolver{}

	metric := kbapi.DatatableESQLMetric{
		Column: "host.name",
	}
	title := "Datatable ESQL RT"
	desc := "Converter test"
	igf := false
	samp := float32(1)
	api := kbapi.DatatableESQL{
		Type:                kbapi.DatatableESQLTypeDataTable,
		Title:               &title,
		Description:         &desc,
		IgnoreGlobalFilters: &igf,
		Sampling:            &samp,
		Styling:             kbapi.DatatableStyling{Density: kbapi.DatatableDensity{Mode: new(kbapi.DatatableDensityModeExpanded)}},
		Metrics:             &[]kbapi.DatatableESQLMetric{metric},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.DataSource))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromDatatableESQL(api))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, resolver, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.DatatableConfig)
	require.Nil(t, blocks.DatatableConfig.NoESQL)
	require.NotNil(t, blocks.DatatableConfig.ESQL)
	assert.Contains(t, blocks.DatatableConfig.ESQL.DataSourceJSON.ValueString(), "FROM metrics-*")

	attrs2, diags := c.BuildAttributes(blocks, resolver)
	require.False(t, diags.HasError(), "%v", diags)

	out, err := attrs2.AsDatatableESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.DatatableESQLTypeDataTable, out.Type)
	require.NotNil(t, out.Title)
	assert.Equal(t, "Datatable ESQL RT", *out.Title)
	dsBytes, err := json.Marshal(out.DataSource)
	require.NoError(t, err)
	assert.Contains(t, string(dsBytes), "FROM metrics-*")
}
