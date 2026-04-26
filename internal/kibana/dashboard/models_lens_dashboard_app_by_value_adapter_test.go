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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_visConfig0ToLensAppConfig0_jsonBridge_metric(t *testing.T) {
	apiChart := kbapi.MetricNoESQL{
		Type:  kbapi.MetricNoESQLTypeMetric,
		Title: new("M"),
		Query: kbapi.FilterSimple{
			Expression: "",
			Language:   new(kbapi.FilterSimpleLanguage("kql")),
		},
		Metrics: []kbapi.MetricNoESQL_Metrics_Item{},
	}
	var vis0 kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, vis0.FromMetricNoESQL(apiChart))

	lens0, err := visConfig0ToLensAppConfig0(vis0)
	require.NoError(t, err)
	metricBack, err := lens0.AsMetricNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MetricNoESQLTypeMetric, metricBack.Type)
	assert.Equal(t, "M", *metricBack.Title)
}

func Test_lensByValueToScratchVisPanel_roundTripFields(t *testing.T) {
	by := lensDashboardAppByValueModel{MetricChartConfig: &metricChartConfigModel{}}
	pm, ok := lensByValueToScratchVisPanel(by)
	require.True(t, ok)
	require.NotNil(t, pm.MetricChartConfig)
}
