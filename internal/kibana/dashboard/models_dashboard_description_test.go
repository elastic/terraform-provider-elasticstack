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
)

// newDashboardAPIResponse builds a minimal GetDashboardsIdResponse with the
// given description pointer, suitable for exercising dashboardPopulateFromAPI.
func newDashboardAPIResponse(description *string) *kbapi.GetDashboardsIdResponse {
	return &kbapi.GetDashboardsIdResponse{
		JSON200: &struct {
			Data     kbapi.KibanaHTTPAPIsKbnDashboardData                   `json:"data"`
			Id       string                                                 `json:"id"`
			Meta     kbapi.KibanaHTTPAPIsKbnAsCodeMeta                      `json:"meta"`
			Warnings *[]kbapi.KibanaHTTPAPIsKbnDashboardDroppedPanelWarning `json:"warnings,omitempty"`
		}{
			Data: kbapi.KibanaHTTPAPIsKbnDashboardData{
				Title:       "test dashboard",
				Description: description,
			},
			Id: "dashboard-id",
		},
	}
}

// TestDashboardModel_populateFromAPI_descriptionNormalization covers the
// intent-preserving null/empty-string normalization for the root-level
// description attribute (REQ-008 / REQ-009). Kibana 9.5 returns `""` for an
// omitted description; the provider must map it back to null when the prior
// plan/state intent was null, while preserving an explicit `description = ""`.
func TestDashboardModel_populateFromAPI_descriptionNormalization(t *testing.T) {
	empty := ""
	nonEmpty := "My dashboard"

	tests := []struct {
		name           string
		apiDescription *string
		priorState     types.String
		want           types.String
	}{
		{
			name:           "API empty string, prior null -> null (9.5 omitted-description bug)",
			apiDescription: &empty,
			priorState:     types.StringNull(),
			want:           types.StringNull(),
		},
		{
			name:           "API empty string, prior empty -> empty (explicit description preserved)",
			apiDescription: &empty,
			priorState:     types.StringValue(""),
			want:           types.StringValue(""),
		},
		{
			name:           "API non-empty, prior null -> value (non-empty normal case)",
			apiDescription: &nonEmpty,
			priorState:     types.StringNull(),
			want:           types.StringValue("My dashboard"),
		},
		{
			name:           "API nil, prior null -> null (8.x / omitted-field case)",
			apiDescription: nil,
			priorState:     types.StringNull(),
			want:           types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &models.DashboardModel{
				Description: tc.priorState,
			}

			diags := dashboardPopulateFromAPI(context.Background(), model, newDashboardAPIResponse(tc.apiDescription), "dashboard-id", "default")
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.want, model.Description)
		})
	}
}
