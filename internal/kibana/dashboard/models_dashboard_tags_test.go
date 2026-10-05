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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestDashboardModel_populateFromAPI_tagsNormalization covers the
// intent-preserving null/empty-list normalization for the root-level tags
// attribute. Kibana omits tags (or returns an empty array) for a dashboard
// without tags; the provider must keep an explicit `tags = []` from the prior
// plan/state instead of flipping it to null, which Terraform reports as
// "produced an unexpected new value: .tags: was cty.ListValEmpty(cty.String),
// but now null".
func TestDashboardModel_populateFromAPI_tagsNormalization(t *testing.T) {
	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	tagList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("tag-1")})

	tests := []struct {
		name       string
		apiTags    *[]string
		priorState types.List
		want       types.List
	}{
		{
			name:       "API nil, prior null -> null",
			apiTags:    nil,
			priorState: types.ListNull(types.StringType),
			want:       types.ListNull(types.StringType),
		},
		{
			name:       "API nil, prior empty -> empty (explicit tags = [] preserved)",
			apiTags:    nil,
			priorState: emptyList,
			want:       emptyList,
		},
		{
			name:       "API empty, prior null -> null",
			apiTags:    &[]string{},
			priorState: types.ListNull(types.StringType),
			want:       types.ListNull(types.StringType),
		},
		{
			name:       "API empty, prior empty -> empty (explicit tags = [] preserved)",
			apiTags:    &[]string{},
			priorState: emptyList,
			want:       emptyList,
		},
		{
			name:       "API nil, prior non-empty -> null (tags removed out of band)",
			apiTags:    nil,
			priorState: tagList,
			want:       types.ListNull(types.StringType),
		},
		{
			name:       "API non-empty, prior empty -> API value",
			apiTags:    &[]string{"tag-1"},
			priorState: emptyList,
			want:       tagList,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &models.DashboardModel{
				Tags: tc.priorState,
			}

			resp := newDashboardAPIResponse(nil)
			resp.JSON200.Data.Tags = tc.apiTags

			diags := dashboardPopulateFromAPI(context.Background(), model, resp, "dashboard-id", "default")
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.want, model.Tags)
		})
	}
}
