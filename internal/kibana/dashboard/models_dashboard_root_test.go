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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dashboardModel_queryToAPI_expression(t *testing.T) {
	for name, expression := range map[string]string{
		"kql":           "response.code:200",
		"empty":         "",
		"leading brace": `{"match_all":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			m := &models.DashboardModel{
				Query: &models.DashboardQueryModel{
					Language:   types.StringValue("kql"),
					Expression: types.StringValue(expression),
				},
			}
			q := dashboardQueryToAPI(m)
			require.NotNil(t, q)
			assert.Equal(t, kbapi.KibanaHTTPAPIsKbnAsCodeQueryLanguage("kql"), q.Language)
			assert.Equal(t, expression, q.Expression)
		})
	}
}

func Test_dashboardModel_queryToAPI_nil(t *testing.T) {
	assert.Nil(t, dashboardQueryToAPI(&models.DashboardModel{}))
}

func Test_dashboardPopulateFromAPI_queryExpression(t *testing.T) {
	for name, expression := range map[string]string{
		"kql":           "response.code:200",
		"empty":         "",
		"leading brace": `{"match_all":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			resp := newDashboardAPIResponse(nil)
			resp.JSON200.Data.Query = &kbapi.KibanaHTTPAPIsKbnAsCodeQuery{
				Language:   kbapi.KibanaHTTPAPIsKbnAsCodeQueryLanguage("lucene"),
				Expression: expression,
			}
			model := &models.DashboardModel{}

			diags := dashboardPopulateFromAPI(context.Background(), model, resp, "dashboard-id", "default")
			require.False(t, diags.HasError())
			require.NotNil(t, model.Query)
			assert.Equal(t, types.StringValue("lucene"), model.Query.Language)
			assert.Equal(t, types.StringValue(expression), model.Query.Expression)
		})
	}
}
