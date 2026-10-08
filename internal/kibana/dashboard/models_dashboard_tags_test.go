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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDashboardAPIResponseWithTags builds a minimal GetDashboardsIdResponse
// with the given tags pointer, suitable for exercising dashboardPopulateFromAPI.
func newDashboardAPIResponseWithTags(tags *[]string) *kbapi.GetDashboardsIdResponse {
	resp := newDashboardAPIResponse(nil)
	resp.JSON200.Data.Tags = tags
	return resp
}

// TestDashboardModel_populateFromAPI_tagsNormalization covers the
// intent-preserving nil/empty-list normalization for the root-level tags
// attribute (REQ-009). A nil-or-empty API tags value must preserve any known
// prior plan/state value (including a known-empty list and null), and only an
// unknown prior value normalizes to null.
func TestDashboardModel_populateFromAPI_tagsNormalization(t *testing.T) {
	tests := []struct {
		name     string
		apiTags  *[]string
		prior    types.List
		want     types.List
		wantDiag bool
	}{
		{
			name:    "API nil, prior known-empty -> known-empty (inconsistent-result bug)",
			apiTags: nil,
			prior:   types.ListValueMust(types.StringType, nil),
			want:    types.ListValueMust(types.StringType, nil),
		},
		{
			name:    "API empty list, prior known-empty -> known-empty",
			apiTags: &[]string{},
			prior:   types.ListValueMust(types.StringType, nil),
			want:    types.ListValueMust(types.StringType, nil),
		},
		{
			name:    "API nil, prior null -> null (omitted tags intent)",
			apiTags: nil,
			prior:   types.ListNull(types.StringType),
			want:    types.ListNull(types.StringType),
		},
		{
			name:    "API non-empty, prior null -> API value (overwrites prior)",
			apiTags: &[]string{"a", "b"},
			prior:   types.ListNull(types.StringType),
			want:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
		},
		{
			name:    "API non-empty, prior known-empty -> API value (overwrites prior)",
			apiTags: &[]string{"a", "b"},
			prior:   types.ListValueMust(types.StringType, nil),
			want:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
		},
		{
			name:    "API nil, prior known non-empty -> prior value retained",
			apiTags: nil,
			prior:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			want:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
		},
		{
			name:    "API empty list, prior known non-empty -> prior value retained",
			apiTags: &[]string{},
			prior:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			want:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
		},
		{
			name:    "API nil, prior unknown -> null (import initializes optional attributes as null)",
			apiTags: nil,
			prior:   types.ListUnknown(types.StringType),
			want:    types.ListNull(types.StringType),
		},
		{
			name:    "API empty list, prior unknown -> null",
			apiTags: &[]string{},
			prior:   types.ListUnknown(types.StringType),
			want:    types.ListNull(types.StringType),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &models.DashboardModel{
				Tags: tc.prior,
			}

			diags := dashboardPopulateFromAPI(context.Background(), model, newDashboardAPIResponseWithTags(tc.apiTags), "dashboard-id", "default")
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.want, model.Tags)
		})
	}
}

// TestDashboardModel_toAPICreateRequest_emptyTagsSentAsEmptyArray confirms the
// write path sends a known-empty tags list as an empty (non-nil) array rather
// than dropping the attribute, so `tags = []` clears remote tags.
func TestDashboardModel_toAPICreateRequest_emptyTagsSentAsEmptyArray(t *testing.T) {
	model := &models.DashboardModel{
		Title: types.StringValue("test dashboard"),
		Tags:  types.ListValueMust(types.StringType, nil),
	}

	req := dashboardToAPICreateRequest(context.Background(), model, &diag.Diagnostics{})

	require.NotNil(t, req.Tags)
	assert.Empty(t, *req.Tags)
}
