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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAPICreateModel_PolicyID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		policyID  types.String
		wantNilID bool
		wantID    string
	}{
		{
			name:      "unknown omits id",
			policyID:  types.StringUnknown(),
			wantNilID: true,
		},
		{
			name:      "null omits id",
			policyID:  types.StringNull(),
			wantNilID: true,
		},
		{
			name:     "explicit id is sent",
			policyID: types.StringValue("my-policy-id"),
			wantID:   "my-policy-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			model := &agentPolicyModel{
				Name:      types.StringValue("test-policy"),
				Namespace: types.StringValue("default"),
				PolicyID:  tc.policyID,
			}

			body, diags := model.toAPICreateModel(ctx, agentPolicyFeatures{})
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			if tc.wantNilID {
				assert.Nil(t, body.Id, "expected Id to be nil, got %v", body.Id)
				return
			}

			require.NotNil(t, body.Id, "expected Id to be set")
			assert.Equal(t, tc.wantID, *body.Id)
		})
	}
}

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

// TestConvertGlobalDataTags_MissingValueEntry asserts that convertGlobalDataTags
// returns an error diagnostic (and does not panic) when a global_data_tags
// map entry has neither string_value nor number_value set. Regression test for
// the SIGSEGV reported when a user writes `global_data_tags = { "x" = {} }`.
// Covers REQ-GDT-002: both null and unknown variants must be guarded.
func TestConvertGlobalDataTags_MissingValueEntry(t *testing.T) {
	ctx := context.Background()

	elemType := getGlobalDataTagsAttrTypes().(attr.TypeWithElementType).ElementType().(types.ObjectType)

	tests := []struct {
		name        string
		stringValue types.String
		numberValue types.Float32
	}{
		{
			name:        "both null",
			stringValue: types.StringNull(),
			numberValue: types.Float32Null(),
		},
		{
			name:        "both unknown",
			stringValue: types.StringUnknown(),
			numberValue: types.Float32Unknown(),
		},
		{
			name:        "string null and number unknown",
			stringValue: types.StringNull(),
			numberValue: types.Float32Unknown(),
		},
		{
			name:        "string unknown and number null",
			stringValue: types.StringUnknown(),
			numberValue: types.Float32Null(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, objDiags := types.ObjectValue(elemType.AttrTypes, map[string]attr.Value{
				"string_value": tc.stringValue,
				"number_value": tc.numberValue,
			})
			assert.False(t, objDiags.HasError(), "failed to build object: %v", objDiags)

			tagsMap, mapDiags := types.MapValue(elemType, map[string]attr.Value{
				"my_tag": entry,
			})
			assert.False(t, mapDiags.HasError(), "failed to build global_data_tags map: %v", mapDiags)

			model := &agentPolicyModel{
				GlobalDataTags: tagsMap,
			}

			result, diags := model.convertGlobalDataTags(ctx, agentPolicyFeatures{SupportsGlobalDataTags: true})

			assert.True(t, diags.HasError(), "expected error diagnostics, got none")
			assert.Nil(t, result, "expected nil result on error, got %v", result)

			var found bool
			for _, d := range diags.Errors() {
				if d.Summary() != "Invalid global_data_tags entry" {
					continue
				}
				found = true
				dwp, ok := d.(diag.DiagnosticWithPath)
				if assert.True(t, ok, "expected attribute diagnostic with Path()") {
					assert.Equal(t, path.Root("global_data_tags").AtMapKey("my_tag"), dwp.Path(),
						"diagnostic should be anchored at global_data_tags[\"my_tag\"]")
				}
				break
			}
			assert.True(t, found, "expected diagnostic with summary 'Invalid global_data_tags entry', got %v", diags)
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
			data := &kbapi.KibanaHTTPAPIsAgentPolicyResponse{
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

// TestComputeFeatureGatedFields verifies the shared helper that both toAPICreateModel and
// toAPIUpdateModel use for version-gated attribute validation.
func TestComputeFeatureGatedFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	boolTrue := true

	t.Run("all fields nil when model fields are null/unknown", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			IsProtected:         types.BoolNull(),
			SupportsAgentless:   types.BoolNull(),
			InactivityTimeout:   customtypes.NewDurationNull(),
			UnenrollmentTimeout: customtypes.NewDurationNull(),
			SpaceIDs:            types.SetNull(types.StringType),
		}
		gated, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{
			SupportsTamperProtection:    true,
			SupportsSupportsAgentless:   true,
			SupportsInactivityTimeout:   true,
			SupportsUnenrollmentTimeout: true,
			SupportsSpaceIDs:            true,
		})
		assert.False(t, diags.HasError())
		assert.Nil(t, gated.isProtected)
		assert.Nil(t, gated.supportsAgentless)
		assert.Nil(t, gated.inactivityTimeout)
		assert.Nil(t, gated.unenrollTimeout)
		assert.Nil(t, gated.spaceIDs)
	})

	t.Run("is_protected error when tamper protection unsupported and true", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			IsProtected: types.BoolValue(true),
		}
		_, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsTamperProtection: false})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), MinVersionTamperProtection.String())
	})

	t.Run("is_protected nil when tamper protection unsupported and false", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			IsProtected: types.BoolValue(false),
		}
		gated, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsTamperProtection: false})
		assert.False(t, diags.HasError())
		assert.Nil(t, gated.isProtected)
	})

	t.Run("is_protected set when tamper protection supported", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			IsProtected: types.BoolValue(true),
		}
		gated, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsTamperProtection: true})
		assert.False(t, diags.HasError())
		require.NotNil(t, gated.isProtected)
		assert.Equal(t, &boolTrue, gated.isProtected)
	})

	t.Run("supports_agentless error when unsupported and set", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			SupportsAgentless: types.BoolValue(true),
		}
		_, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsSupportsAgentless: false})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), MinSupportsAgentlessVersion.String())
	})

	t.Run("inactivity_timeout error when unsupported and set", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			InactivityTimeout: customtypes.NewDurationValue("30s"),
		}
		_, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsInactivityTimeout: false})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), MinVersionInactivityTimeout.String())
	})

	t.Run("inactivity_timeout set when supported", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			InactivityTimeout: customtypes.NewDurationValue("30s"),
		}
		gated, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsInactivityTimeout: true})
		assert.False(t, diags.HasError())
		require.NotNil(t, gated.inactivityTimeout)
		assert.InDelta(t, float32(30), *gated.inactivityTimeout, 0.001)
	})

	t.Run("unenrollment_timeout error when unsupported and set", func(t *testing.T) {
		t.Parallel()
		model := &agentPolicyModel{
			UnenrollmentTimeout: customtypes.NewDurationValue("60s"),
		}
		_, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsUnenrollmentTimeout: false})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), MinVersionUnenrollmentTimeout.String())
	})

	t.Run("space_ids error when unsupported and set", func(t *testing.T) {
		t.Parallel()
		spaceSet, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("default")})
		model := &agentPolicyModel{
			SpaceIDs: spaceSet,
		}
		_, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsSpaceIDs: false})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), MinVersionSpaceIDs.String())
	})

	t.Run("space_ids set when supported", func(t *testing.T) {
		t.Parallel()
		spaceSet, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("default")})
		model := &agentPolicyModel{
			SpaceIDs: spaceSet,
		}
		gated, diags := model.computeFeatureGatedFields(ctx, agentPolicyFeatures{SupportsSpaceIDs: true})
		assert.False(t, diags.HasError())
		require.NotNil(t, gated.spaceIDs)
		assert.Equal(t, []string{"default"}, *gated.spaceIDs)
	})
}

// TestPopulateFromAPI_SpaceIDs_Null_vs_EmptyList asserts the null-preserving
// behavior for the `space_ids` attribute. When the Fleet API omits space_ids
// from its response (data.SpaceIds is nil), the model value must be preserved
// if it was previously configured, preventing the "Provider produced inconsistent
// result after apply" error.
func TestPopulateFromAPI_SpaceIDs_Null_vs_EmptyList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		initial      types.Set // the pre-populate plan/state value for SpaceIDs
		apiValue     []string  // data.SpaceIds as returned by Fleet (nil = omitted)
		wantNull     bool
		wantElements []string // expected elements for non-null cases (nil = skip check)
	}{
		{
			name:     "null in plan and nil from API stays null",
			initial:  types.SetNull(types.StringType),
			apiValue: nil,
			wantNull: true,
		},
		{
			name:         "value in plan and nil from API preserves value",
			initial:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			apiValue:     nil,
			wantNull:     false,
			wantElements: []string{"default"},
		},
		{
			name:         "value in plan and matching value from API stays set",
			initial:      types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default")}),
			apiValue:     []string{"default"},
			wantNull:     false,
			wantElements: []string{"default"},
		},
		{
			name:         "empty set in plan and nil from API preserves empty set",
			initial:      types.SetValueMust(types.StringType, []attr.Value{}),
			apiValue:     nil,
			wantNull:     false,
			wantElements: []string{},
		},
		{
			name:     "null in plan and value from API adopts value",
			initial:  types.SetNull(types.StringType),
			apiValue: []string{"test-space"},
			wantNull: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &agentPolicyModel{
				SpaceIDs: tc.initial,
			}
			var apiSpaceIDs *[]string
			if tc.apiValue != nil {
				apiSpaceIDs = &tc.apiValue
			}
			data := &kbapi.KibanaHTTPAPIsAgentPolicyResponse{
				Id:       "policy-id",
				SpaceIds: apiSpaceIDs,
			}
			diags := model.populateFromAPI(ctx, data)
			assert.False(t, diags.HasError(), "populateFromAPI produced unexpected error diags: %v", diags)

			if tc.wantNull {
				assert.True(t, model.SpaceIDs.IsNull(), "expected SpaceIDs to be null, got %v", model.SpaceIDs)
			} else {
				assert.False(t, model.SpaceIDs.IsNull(), "expected SpaceIDs to be set, got null")
				if tc.wantElements != nil {
					var elements []string
					diags := model.SpaceIDs.ElementsAs(ctx, &elements, false)
					assert.Nil(t, diags, "ElementsAs should not error")
					assert.ElementsMatch(t, tc.wantElements, elements, "expected SpaceIDs elements to match")
				}
			}
		})
	}
}
