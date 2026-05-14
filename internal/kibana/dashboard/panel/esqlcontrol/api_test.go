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

package esqlcontrol_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/esqlcontrol"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract_EsqlControlPanel(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, esqlcontrol.Handler{}, contracttest.Config{
		OmitRequiredLeafPresence: true,
		OmitValidateRequiredZero: true,
		FullAPIResponse: `{
			"type": "esql_control",
			"grid": {"x": 4, "y": 1, "w": 14, "h": 5},
			"id": "esql-contract",
			"config": {
				"control_type": "STATIC_VALUES",
				"selected_options": ["opt_a"],
				"variable_name": "my_var",
				"variable_type": "values"
			}
		}`,
		SkipFields: []string{
			"config.available_options",
			"config.esql_query",
			"display_settings",
			"title",
			"single_select",
			"available_options",
		},
	})
}

func TestPinned_EsqlPinnedRoundtrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ph := esqlcontrol.Handler{}.PinnedHandler()

	in := models.PinnedPanelModel{
		Type: types.StringValue("esql_control"),
		EsqlControlConfig: &models.EsqlControlConfigModel{
			SelectedOptions: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
			}),
			VariableName: types.StringValue("q"),
			VariableType: types.StringValue("values"),
			EsqlQuery:    types.StringValue(""),
			ControlType:  types.StringValue("STATIC_VALUES"),
		},
	}

	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)
	require.True(t, out.Type.Equal(types.StringValue("esql_control")))
	require.NotNil(t, out.EsqlControlConfig)
	require.Equal(t, "q", out.EsqlControlConfig.VariableName.ValueString())
}
