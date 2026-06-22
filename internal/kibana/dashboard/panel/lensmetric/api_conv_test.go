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

package lensmetric

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricChartConfigUsesESQL(t *testing.T) {
	t.Run("detects esql data source", func(t *testing.T) {
		m := &models.MetricChartConfigModel{
			MetricChartCoreTFModel: models.MetricChartCoreTFModel{
				LensChartBaseTFModel: models.LensChartBaseTFModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT(*)"}`),
				},
			},
		}
		assert.True(t, metricChartConfigUsesESQL(m))
	})

	t.Run("data view spec is not esql", func(t *testing.T) {
		m := &models.MetricChartConfigModel{
			MetricChartCoreTFModel: models.MetricChartCoreTFModel{
				LensChartBaseTFModel: models.LensChartBaseTFModel{
					DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"logs-*"}`),
				},
				Query: &models.FilterSimpleModel{
					Language:   types.StringValue("kql"),
					Expression: types.StringValue(""),
				},
			},
		}
		assert.False(t, metricChartConfigUsesESQL(m))
	})
}

func TestMetricChartConfigToAPI_ESQLDataSource(t *testing.T) {
	m := &models.MetricChartConfigModel{
		MetricChartCoreTFModel: models.MetricChartCoreTFModel{
			LensChartBaseTFModel: models.LensChartBaseTFModel{
				DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM kibana_sample_data_logs | STATS requests = COUNT(*)"}`),
			},
			Query: &models.FilterSimpleModel{
				Language:   types.StringValue("kql"),
				Expression: types.StringValue(""),
			},
		},
	}
	require.True(t, metricChartConfigUsesESQL(m), "ES|QL routing must not depend on query being unset")
}
