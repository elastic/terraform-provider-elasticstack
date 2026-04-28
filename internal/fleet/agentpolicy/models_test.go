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

package agentpolicy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeAgentFeature(t *testing.T) {
	tests := []struct {
		name       string
		existing   []apiAgentFeature
		newFeature *apiAgentFeature
		want       *[]apiAgentFeature
	}{
		{
			name:       "nil new feature with empty existing returns nil",
			existing:   nil,
			newFeature: nil,
			want:       nil,
		},
		{
			name:       "nil new feature with empty slice returns nil",
			existing:   []apiAgentFeature{},
			newFeature: nil,
			want:       nil,
		},
		{
			name: "nil new feature preserves existing features",
			existing: []apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
			newFeature: nil,
			want: &[]apiAgentFeature{
				{Name: "feature1", Enabled: true},
				{Name: "feature2", Enabled: false},
			},
		},
		{
			name:       "new feature added to empty existing",
			existing:   nil,
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "new feature added when not present",
			existing: []apiAgentFeature{
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "other", Enabled: true},
				{Name: "fqdn", Enabled: true},
			},
		},
		{
			name: "existing feature replaced",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: false},
				{Name: "other", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: true},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: true},
				{Name: "other", Enabled: true},
			},
		},
		{
			name: "feature disabled replaces enabled",
			existing: []apiAgentFeature{
				{Name: "fqdn", Enabled: true},
			},
			newFeature: &apiAgentFeature{Name: "fqdn", Enabled: false},
			want: &[]apiAgentFeature{
				{Name: "fqdn", Enabled: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeAgentFeature(tt.existing, tt.newFeature)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestConvertHostNameFormatToAgentFeature(t *testing.T) {
	tests := []struct {
		name           string
		hostNameFormat types.String
		want           *apiAgentFeature
	}{
		{
			name:           "null host_name_format returns nil",
			hostNameFormat: types.StringNull(),
			want:           nil,
		},
		{
			name:           "unknown host_name_format returns nil",
			hostNameFormat: types.StringUnknown(),
			want:           nil,
		},
		{
			name:           "fqdn returns enabled feature",
			hostNameFormat: types.StringValue(HostNameFormatFQDN),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: true},
		},
		{
			name:           "hostname returns disabled feature",
			hostNameFormat: types.StringValue(HostNameFormatHostname),
			want:           &apiAgentFeature{Name: agentFeatureFQDN, Enabled: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &agentPolicyModel{
				HostNameFormat: tt.hostNameFormat,
			}

			got := model.convertHostNameFormatToAgentFeature()

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Enabled, got.Enabled)
		})
	}
}

// TestPopulateFromAPI_Description_Null_vs_EmptyString asserts the
// null-preserving behavior for the `description` attribute. Regression test
// for https://github.com/elastic/terraform-provider-elasticstack/issues/993:
// the Fleet API returns an empty string for an unset description, which
// previously triggered "Provider produced inconsistent result after apply:
// was null, but now cty.StringVal("")" whenever the user's plan omitted the
// attribute.
func TestPopulateFromAPI_Description_Null_vs_EmptyString(t *testing.T) {
	emptyStr := ""
	foo := "foo"

	tests := []struct {
		name      string
		initial   types.String // the pre-populate plan/state value
		apiValue  *string      // data.Description as returned by Fleet
		wantNull  bool
		wantValue string // only meaningful when wantNull is false
	}{
		{
			name:     "null in plan and nil from API stays null",
			initial:  types.StringNull(),
			apiValue: nil,
			wantNull: true,
		},
		{
			name:     "null in plan and empty string from API stays null",
			initial:  types.StringNull(),
			apiValue: &emptyStr,
			wantNull: true,
		},
		{
			name:      "null in plan and value from API adopts value",
			initial:   types.StringNull(),
			apiValue:  &foo,
			wantNull:  false,
			wantValue: "foo",
		},
		{
			name:      "empty string in plan and empty string from API stays empty string",
			initial:   types.StringValue(""),
			apiValue:  &emptyStr,
			wantNull:  false,
			wantValue: "",
		},
		{
			name:      "value in plan and matching value from API stays value",
			initial:   types.StringValue("foo"),
			apiValue:  &foo,
			wantNull:  false,
			wantValue: "foo",
		},
		{
			name:      "value in plan and different value from API adopts API value",
			initial:   types.StringValue("foo"),
			apiValue:  &emptyStr, // user removed out-of-band; Kibana response is ""
			wantNull:  false,
			wantValue: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &agentPolicyModel{
				Description: tc.initial,
			}
			data := &kbapi.AgentPolicy{
				Id:          "policy-id",
				Description: tc.apiValue,
			}
			diags := model.populateFromAPI(context.Background(), data)
			assert.False(t, diags.HasError(), "populateFromAPI produced unexpected error diags: %v", diags)

			if tc.wantNull {
				assert.True(t, model.Description.IsNull(), "expected Description to be null, got %q", model.Description.ValueString())
			} else {
				assert.False(t, model.Description.IsNull(), "expected Description to be set, got null")
				assert.Equal(t, tc.wantValue, model.Description.ValueString())
			}
		})
	}
}
