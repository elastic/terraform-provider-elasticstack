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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/require"
)

func Test_chartFilterJSONModel_roundTrip_xyChart(t *testing.T) {
	raw := `{"type":"condition","condition":{"field":"host.name","operator":"is","value":"staging"}}`
	var item kbapi.XyChart_Filters_Item
	require.NoError(t, json.Unmarshal([]byte(raw), &item))

	m := chartFilterJSONModel{}
	diags := m.populateFromAPIItem(item)
	require.False(t, diags.HasError())

	var out kbapi.XyChart_Filters_Item
	diags = decodeChartFilterJSON(m.FilterJSON, &out)
	require.False(t, diags.HasError())

	cond, err := out.AsKbnAsCodeFiltersSchemaAsCodeConditionFilterSchema()
	require.NoError(t, err)
	require.Equal(t, kbapi.Condition, cond.Type)
	isCond, err := cond.Condition.AsKbnAsCodeFiltersSchemaConditionIs()
	require.NoError(t, err)
	require.Equal(t, "host.name", isCond.Field)
	val, err := isCond.Value.AsKbnAsCodeFiltersSchemaConditionIsValue0()
	require.NoError(t, err)
	require.Equal(t, "staging", val)
}

func Test_decodeChartFilterJSON_rejects_empty(t *testing.T) {
	var item kbapi.XyChart_Filters_Item
	diags := decodeChartFilterJSON(jsontypes.NewNormalizedNull(), &item)
	require.True(t, diags.HasError())
}
