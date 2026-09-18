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

package lensxy

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataLayerFromAPINoESQL_preservesPriorYWhenKibanaInjectsAxis(t *testing.T) {
	t.Parallel()

	prior := `{"empty_as_null":true,"operation":"count"}`
	m := &models.DataLayerModel{
		Y: []models.YMetricModel{{
			ConfigJSON: jsontypes.NewNormalizedValue(prior),
		}},
	}

	var yItem kbapi.KibanaHTTPAPIsXyLayerNoESQL_Y_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count","empty_as_null":true,"axis":"y","color":{"type":"auto"}}`), &yItem))

	diags := dataLayerFromAPINoESQL(t.Context(), m, kbapi.KibanaHTTPAPIsXyLayerNoESQL{
		Y: []kbapi.KibanaHTTPAPIsXyLayerNoESQL_Y_Item{yItem},
	})
	require.False(t, diags.HasError(), "%v", diags)
	require.Len(t, m.Y, 1)
	assert.JSONEq(t, prior, m.Y[0].ConfigJSON.ValueString())
}
