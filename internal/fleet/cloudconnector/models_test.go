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

package cloudconnector

import (
	"fmt"
	"testing"

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateFromAPI_AWSDualPopulation(t *testing.T) {
	item := fleetclient.CloudConnectorItem{
		ID:                 "conn-aws-1",
		Name:               "aws-connector",
		CloudProvider:      "aws",
		PackagePolicyCount: 0,
		CreatedAt:          "2026-01-01T00:00:00.000Z",
		UpdatedAt:          "2026-01-02T00:00:00.000Z",
		Vars: map[string]any{
			"role_arn": map[string]any{
				"type":  "text",
				"value": "arn:aws:iam::123456789012:role/Elastic",
			},
			"external_id": map[string]any{
				"type": "password",
				"value": map[string]any{
					"id":          "secret-ref-abc",
					"isSecretRef": true,
				},
			},
		},
	}

	var model cloudConnectorModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, "default/conn-aws-1", model.ID.ValueString())
	assert.Equal(t, "conn-aws-1", model.CloudConnectorID.ValueString())
	assert.Equal(t, "aws", model.CloudProvider.ValueString())

	require.False(t, model.Vars.IsNull())
	require.Len(t, model.Vars.Elements(), 2)

	require.False(t, model.AWS.IsNull())
	awsAttrs := model.AWS.Attributes()
	assert.Equal(t, "arn:aws:iam::123456789012:role/Elastic", awsAttrs["role_arn"].(types.String).ValueString())
	assert.True(t, awsAttrs["external_id"].(types.String).IsNull())

	secretRefObj := awsAttrs["external_id_secret_ref"].(types.Object)
	require.False(t, secretRefObj.IsNull())
	secretRefAttrs := secretRefObj.Attributes()
	assert.Equal(t, "secret-ref-abc", secretRefAttrs["id"].(types.String).ValueString())
	assert.True(t, secretRefAttrs["is_secret_ref"].(types.Bool).ValueBool())

	assert.True(t, model.Azure.IsNull())
}

func TestPopulateFromAPI_AzureDualPopulation(t *testing.T) {
	item := fleetclient.CloudConnectorItem{
		ID:                 "conn-azure-1",
		Name:               "azure-connector",
		CloudProvider:      "azure",
		PackagePolicyCount: 2,
		CreatedAt:          "2026-01-01T00:00:00.000Z",
		UpdatedAt:          "2026-01-02T00:00:00.000Z",
		Vars: map[string]any{
			"tenant_id": map[string]any{
				"type": "password",
				"value": map[string]any{
					"id":          "tenant-secret-ref",
					"isSecretRef": true,
				},
			},
			"client_id": map[string]any{
				"type": "password",
				"value": map[string]any{
					"id":          "client-secret-ref",
					"isSecretRef": true,
				},
			},
			wireKeyAzureCredentialsCloudConnectorID: map[string]any{
				"type":  "text",
				"value": "azure-connector-id",
			},
		},
	}

	var model cloudConnectorModel
	diags := model.populateFromAPI("security", item)
	require.False(t, diags.HasError())

	require.False(t, model.Vars.IsNull())
	require.Len(t, model.Vars.Elements(), 3)

	require.False(t, model.Azure.IsNull())
	azureAttrs := model.Azure.Attributes()
	assert.True(t, azureAttrs["tenant_id"].(types.String).IsNull())
	assert.True(t, azureAttrs["client_id"].(types.String).IsNull())
	assert.Equal(t, "azure-connector-id", azureAttrs["cloud_connector_id"].(types.String).ValueString())
	assert.Equal(t, "tenant-secret-ref", azureAttrs["tenant_id_secret_ref"].(types.Object).Attributes()["id"].(types.String).ValueString())
	assert.Equal(t, "client-secret-ref", azureAttrs["client_id_secret_ref"].(types.Object).Attributes()["id"].(types.String).ValueString())

	assert.True(t, model.AWS.IsNull())
	assert.Equal(t, "security", model.SpaceID.ValueString())
}

func TestPopulateFromAPI_AWSExtraVarKey(t *testing.T) {
	t.Run("extra unmodelled key nulls AWS typed block", func(t *testing.T) {
		item := fleetclient.CloudConnectorItem{
			ID:            "conn-aws-extra",
			Name:          "aws-extra",
			CloudProvider: "aws",
			CreatedAt:     "2026-01-01T00:00:00.000Z",
			UpdatedAt:     "2026-01-02T00:00:00.000Z",
			Vars: map[string]any{
				"role_arn": "arn:aws:iam::123456789012:role/Elastic",
				"external_id": map[string]any{
					"type": "password",
					"value": map[string]any{
						"id":          "secret-ref-xyz",
						"isSecretRef": true,
					},
				},
				"region": "us-east-1",
			},
		}

		var model cloudConnectorModel
		diags := model.populateFromAPI("default", item)
		require.False(t, diags.HasError())

		require.Len(t, model.Vars.Elements(), 3)
		assert.True(t, model.AWS.IsNull())
	})

	t.Run("standard keys missing leaves AWS block null but vars populated", func(t *testing.T) {
		item := fleetclient.CloudConnectorItem{
			ID:            "conn-aws-partial",
			Name:          "aws-partial",
			CloudProvider: "aws",
			CreatedAt:     "2026-01-01T00:00:00.000Z",
			UpdatedAt:     "2026-01-02T00:00:00.000Z",
			Vars: map[string]any{
				"role_arn": "arn:aws:iam::123456789012:role/Elastic",
				"region":   "us-east-1",
			},
		}

		var model cloudConnectorModel
		diags := model.populateFromAPI("default", item)
		require.False(t, diags.HasError())

		require.Len(t, model.Vars.Elements(), 2)
		assert.True(t, model.AWS.IsNull())
	})
}

func TestPopulateFromAPI_CloudProviderMismatchLeavesTypedBlocksNull(t *testing.T) {
	awsShapedVars := map[string]any{
		"role_arn": "arn:aws:iam::123456789012:role/Elastic",
		"external_id": map[string]any{
			"type": "password",
			"value": map[string]any{
				"id":          "secret-ref-xyz",
				"isSecretRef": true,
			},
		},
	}

	tests := []struct {
		name          string
		cloudProvider string
	}{
		{name: "gcp provider", cloudProvider: "gcp"},
		{name: "azure provider", cloudProvider: "azure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := fleetclient.CloudConnectorItem{
				ID:            "conn-mismatch",
				Name:          "mismatch",
				CloudProvider: tt.cloudProvider,
				CreatedAt:     "2026-01-01T00:00:00.000Z",
				UpdatedAt:     "2026-01-02T00:00:00.000Z",
				Vars:          awsShapedVars,
			}

			var model cloudConnectorModel
			diags := model.populateFromAPI("default", item)
			require.False(t, diags.HasError())

			require.Len(t, model.Vars.Elements(), 2)
			assert.True(t, model.AWS.IsNull())
			assert.True(t, model.Azure.IsNull())
		})
	}
}

func TestPopulateFromAPI_GCPVarsOnly(t *testing.T) {
	item := fleetclient.CloudConnectorItem{
		ID:            "conn-gcp-1",
		Name:          "gcp-connector",
		CloudProvider: "gcp",
		CreatedAt:     "2026-01-01T00:00:00.000Z",
		UpdatedAt:     "2026-01-02T00:00:00.000Z",
		Vars: map[string]any{
			"project_id": "my-gcp-project",
			"enabled":    true,
		},
	}

	var model cloudConnectorModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError())

	require.False(t, model.Vars.IsNull())
	require.Len(t, model.Vars.Elements(), 2)
	assert.True(t, model.AWS.IsNull())
	assert.True(t, model.Azure.IsNull())
}

func TestPopulateFromAPI_NilVars(t *testing.T) {
	item := fleetclient.CloudConnectorItem{
		ID:            "conn-nil-vars",
		Name:          "nil-vars",
		CloudProvider: "aws",
		CreatedAt:     "2026-01-01T00:00:00.000Z",
		UpdatedAt:     "2026-01-02T00:00:00.000Z",
		Vars:          nil,
	}

	var model cloudConnectorModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError())

	assert.False(t, model.Vars.IsNull())
	assert.Empty(t, model.Vars.Elements())
	assert.True(t, model.AWS.IsNull())
	assert.True(t, model.Azure.IsNull())
}

func TestPopulateFromAPI_ComputedFieldsWhenPresent(t *testing.T) {
	accountType := "single-account"
	namespace := "default"
	verificationStatus := "verified"
	verificationStartedAt := "2026-01-01T01:00:00.000Z"
	verificationFailedAt := "2026-01-01T02:00:00.000Z"

	item := fleetclient.CloudConnectorItem{
		ID:                    "conn-full",
		Name:                  "full",
		CloudProvider:         "gcp",
		AccountType:           &accountType,
		Namespace:             &namespace,
		VerificationStatus:    &verificationStatus,
		VerificationStartedAt: &verificationStartedAt,
		VerificationFailedAt:  &verificationFailedAt,
		CreatedAt:             "2026-01-01T00:00:00.000Z",
		UpdatedAt:             "2026-01-02T00:00:00.000Z",
		Vars:                  map[string]any{},
	}

	var model cloudConnectorModel
	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError())

	assert.Equal(t, accountType, model.AccountType.ValueString())
	assert.Equal(t, namespace, model.Namespace.ValueString())
	assert.Equal(t, verificationStatus, model.VerificationStatus.ValueString())
	assert.Equal(t, verificationStartedAt, model.VerificationStartedAt.ValueString())
	assert.Equal(t, verificationFailedAt, model.VerificationFailedAt.ValueString())
}

func TestPopulateFromAPI_VarsUnionArms(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		assertFn func(t *testing.T, attrs map[string]attr.Value)
	}{
		{
			name:  "string arm",
			key:   "plain_string",
			value: "hello",
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.Equal(t, "hello", attrs["string"].(types.String).ValueString())
				assert.True(t, attrs["number"].(types.Float64).IsNull())
				assert.True(t, attrs["bool"].(types.Bool).IsNull())
				assert.True(t, attrs["type"].(types.String).IsNull())
			},
		},
		{
			name:  "number arm float64",
			key:   "plain_number",
			value: float64(42.5),
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.InDelta(t, 42.5, attrs["number"].(types.Float64).ValueFloat64(), 0.0001)
				assert.True(t, attrs["string"].(types.String).IsNull())
			},
		},
		{
			name:  "number arm float32",
			key:   "plain_number_f32",
			value: float32(3.5),
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.InDelta(t, 3.5, attrs["number"].(types.Float64).ValueFloat64(), 0.0001)
			},
		},
		{
			name:  "number arm int",
			key:   "plain_number_int",
			value: 7,
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.InDelta(t, 7, attrs["number"].(types.Float64).ValueFloat64(), 0.0001)
			},
		},
		{
			name:  "number arm int64",
			key:   "plain_number_int64",
			value: int64(99),
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.InDelta(t, 99, attrs["number"].(types.Float64).ValueFloat64(), 0.0001)
			},
		},
		{
			name:  "bool arm",
			key:   "plain_bool",
			value: true,
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.True(t, attrs["bool"].(types.Bool).ValueBool())
			},
		},
		{
			name: "structured with value",
			key:  "structured_value",
			value: map[string]any{
				"type":   "text",
				"frozen": true,
				"value":  "configured",
			},
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.Equal(t, "text", attrs["type"].(types.String).ValueString())
				assert.True(t, attrs["frozen"].(types.Bool).ValueBool())
				assert.Equal(t, "configured", attrs["value"].(types.String).ValueString())
				assert.True(t, attrs["secret_ref"].(types.Object).IsNull())
				assert.True(t, attrs["secret_value"].(types.String).IsNull())
			},
		},
		{
			name: "structured with secret ref",
			key:  "structured_secret",
			value: map[string]any{
				"type": "password",
				"value": map[string]any{
					"id":          "ref-id",
					"isSecretRef": true,
				},
			},
			assertFn: func(t *testing.T, attrs map[string]attr.Value) {
				assert.Equal(t, "password", attrs["type"].(types.String).ValueString())
				assert.True(t, attrs["value"].(types.String).IsNull())
				assert.True(t, attrs["secret_value"].(types.String).IsNull())
				secretRefObj := attrs["secret_ref"].(types.Object)
				require.False(t, secretRefObj.IsNull())
				secretRefAttrs := secretRefObj.Attributes()
				assert.Equal(t, "ref-id", secretRefAttrs["id"].(types.String).ValueString())
				assert.True(t, secretRefAttrs["is_secret_ref"].(types.Bool).ValueBool())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := fleetclient.CloudConnectorItem{
				ID:            "conn-union",
				Name:          "union-test",
				CloudProvider: "gcp",
				CreatedAt:     "2026-01-01T00:00:00.000Z",
				UpdatedAt:     "2026-01-02T00:00:00.000Z",
				Vars: map[string]any{
					tt.key: tt.value,
				},
			}

			var model cloudConnectorModel
			diags := model.populateFromAPI("default", item)
			require.False(t, diags.HasError())

			obj := model.Vars.Elements()[tt.key].(types.Object)
			tt.assertFn(t, obj.Attributes())
		})
	}
}

func TestPopulateFromAPI_MissingOptionalFieldsNull(t *testing.T) {
	item := fleetclient.CloudConnectorItem{
		ID:                 "conn-minimal",
		Name:               "minimal",
		CloudProvider:      "gcp",
		PackagePolicyCount: 0,
		CreatedAt:          "2026-01-01T00:00:00.000Z",
		UpdatedAt:          "2026-01-02T00:00:00.000Z",
		Vars:               map[string]any{},
	}

	var model cloudConnectorModel
	model.ForceDelete = types.BoolValue(true)

	diags := model.populateFromAPI("default", item)
	require.False(t, diags.HasError())

	assert.True(t, model.AccountType.IsNull())
	assert.True(t, model.Namespace.IsNull())
	assert.True(t, model.VerificationStatus.IsNull())
	assert.True(t, model.VerificationStartedAt.IsNull())
	assert.True(t, model.VerificationFailedAt.IsNull())
	assert.False(t, model.Vars.IsNull())
	assert.Empty(t, model.Vars.Elements())
	assert.True(t, model.ForceDelete.ValueBool(), "populateFromAPI must not overwrite ForceDelete")
}

func TestGetVersionRequirements(t *testing.T) {
	var model cloudConnectorModel
	reqs, diags := model.GetVersionRequirements()
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	assert.Equal(t, "9.2.0", reqs[0].MinVersion.String())
	assert.Equal(t, fmt.Sprintf("Fleet cloud connectors require Kibana v%s or later.", cloudConnectorMinVersion), reqs[0].ErrorMessage)
}

func TestVarValueToElement_UnsupportedShape(t *testing.T) {
	_, diags := varValueToElement("bad", map[string]any{"value": "no-type"})
	require.True(t, diags.HasError())
}

func TestStructuredVarToElement_MissingType(t *testing.T) {
	_, diags := structuredVarToElement("k", map[string]any{
		"value": "configured",
	})
	require.True(t, diags.HasError())
}

func TestSecretRefFromMap_MissingIsSecretRef(t *testing.T) {
	_, diags := secretRefFromMap("k", map[string]any{
		"id": "ref-id",
	})
	require.True(t, diags.HasError())
}

func TestAzureBlockFromVars_MissingKey(t *testing.T) {
	obj, diags := azureBlockFromVars(map[string]any{
		"tenant_id": "tenant-uuid",
		"client_id": "client-uuid",
	})
	require.False(t, diags.HasError())
	assert.True(t, obj.IsNull())
}

func TestHasExactVarKeys(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]any
		keys     []string
		expected bool
	}{
		{
			name:     "nil vars with required keys",
			vars:     nil,
			keys:     []string{"role_arn", "external_id"},
			expected: false,
		},
		{
			name: "exact match",
			vars: map[string]any{
				"role_arn":    "arn",
				"external_id": "secret",
			},
			keys:     []string{"role_arn", "external_id"},
			expected: true,
		},
		{
			name: "extra key",
			vars: map[string]any{
				"role_arn":    "arn",
				"external_id": "secret",
				"region":      "us-east-1",
			},
			keys:     []string{"role_arn", "external_id"},
			expected: false,
		},
		{
			name: "missing key",
			vars: map[string]any{
				"role_arn": "arn",
			},
			keys:     []string{"role_arn", "external_id"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, hasExactVarKeys(tt.vars, tt.keys...))
		})
	}
}
