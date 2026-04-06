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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dashboardModel_queryToAPI_neitherTextNorJSON(t *testing.T) {
	m := &dashboardModel{
		Query: &dashboardQueryModel{
			Language: types.StringValue("kql"),
			Text:     types.StringNull(),
			JSON:     jsontypes.NewNormalizedNull(),
		},
	}
	_, diags := m.queryToAPI()
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary(), "Invalid dashboard query")
}

func Test_dashboardModel_queryToAPI_jsonBranch(t *testing.T) {
	m := &dashboardModel{
		Query: &dashboardQueryModel{
			Language: types.StringValue("kuery"),
			Text:     types.StringNull(),
			JSON:     jsontypes.NewNormalizedValue(`{"match_all":{}}`),
		},
	}
	q, diags := m.queryToAPI()
	require.False(t, diags.HasError())
	assert.Equal(t, "kuery", q.Language)
	obj, err := q.Query.AsKbnEsQueryServerQuerySchemaQuery1()
	require.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_dashboardModel_queryToAPI_bothTextAndJSON(t *testing.T) {
	m := &dashboardModel{
		Query: &dashboardQueryModel{
			Language: types.StringValue("kuery"),
			Text:     types.StringValue("response.code:200"),
			JSON:     jsontypes.NewNormalizedValue(`{"match_all":{}}`),
		},
	}
	_, diags := m.queryToAPI()
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary(), "Invalid dashboard query")
}
