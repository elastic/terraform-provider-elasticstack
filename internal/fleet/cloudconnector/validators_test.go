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
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVarsElementValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	nullSecretRef := types.ObjectNull(secretRefAttrTypes())

	validCases := []struct {
		name  string
		attrs map[string]attr.Value
	}{
		{
			name: "arm 1 string",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringValue("hello"),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
		},
		{
			name: "arm 2 number",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Value(3.14),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
		},
		{
			name: "arm 3 bool",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolValue(true),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
		},
		{
			name: "arm 4 structured value",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("text"),
				attrVarsFrozen:      types.BoolValue(false),
				attrVarsValue:       types.StringValue("configured"),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
		},
		{
			name: "arm 4 write-only secret value",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("password"),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringValue("shhh"),
				attrVarsSecretRef:   nullSecretRef,
			},
		},
	}

	for _, tc := range validCases {
		t.Run("valid/"+tc.name, func(t *testing.T) {
			t.Parallel()
			obj := types.ObjectValueMust(varsElementAttrTypes(), tc.attrs)
			resp := &validator.ObjectResponse{}
			varsElementValidator{}.ValidateObject(ctx, validator.ObjectRequest{
				Path:        path.Root(attrVarsMap).AtMapKey("k"),
				Config:      tfsdk.Config{},
				ConfigValue: obj,
			}, resp)
			assert.False(t, resp.Diagnostics.HasError(), "expected valid config, got: %v", resp.Diagnostics)
		})
	}

	invalidCases := []struct {
		name       string
		attrs      map[string]attr.Value
		wantSubstr string
	}{
		{
			name: "multiple bare arms",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringValue("a"),
				attrVarsNumber:      types.Float64Value(1),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "At most one bare var arm",
		},
		{
			name: "bare and structured arms",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringValue("a"),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("text"),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "cannot be combined",
		},
		{
			name: "structured without value arm",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("text"),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "exactly one of `value` or `secret_value`",
		},
		{
			name: "structured with multiple value arms",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("password"),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringValue("plain"),
				attrVarsSecretValue: types.StringValue("secret"),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "exactly one of `value` or `secret_value`",
		},
		{
			name: "computed secret_ref in config",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringValue("password"),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef: types.ObjectValueMust(secretRefAttrTypes(), map[string]attr.Value{
					attrSecretRefID:          types.StringValue("ref-id"),
					attrSecretRefIsSecretRef: types.BoolValue(true),
				}),
			},
			wantSubstr: "computed-only",
		},
		{
			name: "empty element",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "At least one vars union arm",
		},
		{
			name: "frozen without type",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolValue(true),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "`type` is required when any of `value`, `secret_value`, or `frozen`",
		},
		{
			name: "value without type",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringValue("abc"),
				attrVarsSecretValue: types.StringNull(),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "`type` is required when any of `value`, `secret_value`, or `frozen`",
		},
		{
			name: "secret_value without type",
			attrs: map[string]attr.Value{
				attrVarsString:      types.StringNull(),
				attrVarsNumber:      types.Float64Null(),
				attrVarsBool:        types.BoolNull(),
				attrVarsType:        types.StringNull(),
				attrVarsFrozen:      types.BoolNull(),
				attrVarsValue:       types.StringNull(),
				attrVarsSecretValue: types.StringValue("shhh"),
				attrVarsSecretRef:   nullSecretRef,
			},
			wantSubstr: "`type` is required when any of `value`, `secret_value`, or `frozen`",
		},
	}

	for _, tc := range invalidCases {
		t.Run("invalid/"+tc.name, func(t *testing.T) {
			t.Parallel()
			obj := types.ObjectValueMust(varsElementAttrTypes(), tc.attrs)
			resp := &validator.ObjectResponse{}
			varsElementValidator{}.ValidateObject(ctx, validator.ObjectRequest{
				Path:        path.Root(attrVarsMap).AtMapKey("k"),
				Config:      tfsdk.Config{},
				ConfigValue: obj,
			}, resp)
			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Summary()+resp.Diagnostics.Errors()[0].Detail(), tc.wantSubstr)
		})
	}
}

func TestProviderBlockMatchesCloudProviderValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewResource()

	t.Run("aws block with azure provider", func(t *testing.T) {
		t.Parallel()
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrCloudProvider: tftypes.NewValue(tftypes.String, cloudProviderAzure),
			attrAWSBlock: tftypes.NewValue(awsTerraformObjectType(ctx, t, r), map[string]tftypes.Value{
				attrAWSRoleArn:             tftypes.NewValue(tftypes.String, "arn:aws:iam::123456789012:role/Elastic"),
				attrAWSExternalID:          tftypes.NewValue(tftypes.String, nil),
				attrAWSExternalIDSecretRef: tftypes.NewValue(secretRefTerraformObjectType(ctx, t, r), nil),
			}),
		})

		resp := &resource.ValidateConfigResponse{}
		providerBlockMatchesCloudProvider{}.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		require.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "requires `cloud_provider = \"aws\"`")
	})

	t.Run("azure block with aws provider", func(t *testing.T) {
		t.Parallel()
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrCloudProvider: tftypes.NewValue(tftypes.String, cloudProviderAWS),
			attrAzureBlock:    tftypes.NewValue(azureTerraformObjectType(ctx, t, r), azureConfigBlockValues(ctx, t, r, "tenant-uuid", "client-uuid", "azure-connector-id")),
		})

		resp := &resource.ValidateConfigResponse{}
		providerBlockMatchesCloudProvider{}.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		require.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "requires `cloud_provider = \"azure\"`")
	})

	t.Run("aws block with matching provider", func(t *testing.T) {
		t.Parallel()
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrCloudProvider: tftypes.NewValue(tftypes.String, cloudProviderAWS),
			attrAWSBlock: tftypes.NewValue(awsTerraformObjectType(ctx, t, r), map[string]tftypes.Value{
				attrAWSRoleArn:             tftypes.NewValue(tftypes.String, "arn:aws:iam::123456789012:role/Elastic"),
				attrAWSExternalID:          tftypes.NewValue(tftypes.String, nil),
				attrAWSExternalIDSecretRef: tftypes.NewValue(secretRefTerraformObjectType(ctx, t, r), nil),
			}),
		})

		resp := &resource.ValidateConfigResponse{}
		providerBlockMatchesCloudProvider{}.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		assert.False(t, resp.Diagnostics.HasError())
	})

	t.Run("azure block with matching provider", func(t *testing.T) {
		t.Parallel()
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrCloudProvider: tftypes.NewValue(tftypes.String, cloudProviderAzure),
			attrAzureBlock:    tftypes.NewValue(azureTerraformObjectType(ctx, t, r), azureConfigBlockValues(ctx, t, r, "tenant-uuid", "client-uuid", "azure-connector-id")),
		})

		resp := &resource.ValidateConfigResponse{}
		providerBlockMatchesCloudProvider{}.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		assert.False(t, resp.Diagnostics.HasError())
	})

	t.Run("tolerates unknown cloud_provider", func(t *testing.T) {
		t.Parallel()
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrCloudProvider: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			attrAWSBlock: tftypes.NewValue(awsTerraformObjectType(ctx, t, r), map[string]tftypes.Value{
				attrAWSRoleArn:             tftypes.NewValue(tftypes.String, "arn:aws:iam::123456789012:role/Elastic"),
				attrAWSExternalID:          tftypes.NewValue(tftypes.String, nil),
				attrAWSExternalIDSecretRef: tftypes.NewValue(secretRefTerraformObjectType(ctx, t, r), nil),
			}),
		})

		resp := &resource.ValidateConfigResponse{}
		providerBlockMatchesCloudProvider{}.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		assert.False(t, resp.Diagnostics.HasError())
	})
}

func buildCloudConnectorTestConfig(ctx context.Context, t *testing.T, r resource.Resource, overrides map[string]tftypes.Value) tfsdk.Config {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	attrTypes := schemaResp.Schema.Type().TerraformType(ctx)
	values := make(map[string]tftypes.Value, len(attrTypes.(tftypes.Object).AttributeTypes)+1)
	for name, attrType := range attrTypes.(tftypes.Object).AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	kbConnType := schemaResp.Schema.Blocks["kibana_connection"].Type().TerraformType(ctx)
	values["kibana_connection"] = tftypes.NewValue(kbConnType, nil)

	maps.Copy(values, overrides)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(attrTypes, values),
	}
}

func awsTerraformObjectType(ctx context.Context, t *testing.T, r resource.Resource) tftypes.Type {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())
	return schemaResp.Schema.Attributes[attrAWSBlock].GetType().TerraformType(ctx)
}

func secretRefTerraformObjectType(ctx context.Context, t *testing.T, r resource.Resource) tftypes.Type {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())
	awsAttr := schemaResp.Schema.Attributes[attrAWSBlock].(schema.SingleNestedAttribute)
	return awsAttr.Attributes[attrAWSExternalIDSecretRef].GetType().TerraformType(ctx)
}

func azureConfigBlockValues(ctx context.Context, t *testing.T, r resource.Resource, tenantID, clientID, connectorID string) map[string]tftypes.Value {
	t.Helper()
	secretRefType := secretRefTerraformObjectType(ctx, t, r)
	return map[string]tftypes.Value{
		attrAzureTenantID:          tftypes.NewValue(tftypes.String, tenantID),
		attrAzureClientID:          tftypes.NewValue(tftypes.String, clientID),
		attrAzureTenantIDSecretRef: tftypes.NewValue(secretRefType, nil),
		attrAzureClientIDSecretRef: tftypes.NewValue(secretRefType, nil),
		attrAzureCloudConnectorID:  tftypes.NewValue(tftypes.String, connectorID),
	}
}

func azureTerraformObjectType(ctx context.Context, t *testing.T, r resource.Resource) tftypes.Type {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())
	return schemaResp.Schema.Attributes[attrAzureBlock].GetType().TerraformType(ctx)
}

func varsTerraformMapType(ctx context.Context, t *testing.T, r resource.Resource) tftypes.Type {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())
	return schemaResp.Schema.Attributes[attrVarsMap].GetType().TerraformType(ctx)
}

func TestResourceConfigValidators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := NewResource().(resource.ResourceWithConfigValidators)
	validators := r.ConfigValidators(ctx)
	require.Len(t, validators, 2)

	t.Run("ExactlyOneOf rejects aws and vars together", func(t *testing.T) {
		t.Parallel()
		varsMapType := varsTerraformMapType(ctx, t, r)
		mapType, ok := varsMapType.(tftypes.Map)
		require.True(t, ok)
		varsElemType := mapType.ElementType
		cfg := buildCloudConnectorTestConfig(ctx, t, r, map[string]tftypes.Value{
			attrAWSBlock: tftypes.NewValue(awsTerraformObjectType(ctx, t, r), map[string]tftypes.Value{
				attrAWSRoleArn:             tftypes.NewValue(tftypes.String, "arn:aws:iam::123456789012:role/Elastic"),
				attrAWSExternalID:          tftypes.NewValue(tftypes.String, nil),
				attrAWSExternalIDSecretRef: tftypes.NewValue(secretRefTerraformObjectType(ctx, t, r), nil),
			}),
			attrVarsMap: tftypes.NewValue(varsMapType, map[string]tftypes.Value{
				"role_arn": tftypes.NewValue(varsElemType, map[string]tftypes.Value{
					attrVarsString:      tftypes.NewValue(tftypes.String, "arn:aws:iam::123456789012:role/Elastic"),
					attrVarsNumber:      tftypes.NewValue(tftypes.Number, nil),
					attrVarsBool:        tftypes.NewValue(tftypes.Bool, nil),
					attrVarsType:        tftypes.NewValue(tftypes.String, nil),
					attrVarsFrozen:      tftypes.NewValue(tftypes.Bool, nil),
					attrVarsValue:       tftypes.NewValue(tftypes.String, nil),
					attrVarsSecretValue: tftypes.NewValue(tftypes.String, nil),
					attrVarsSecretRef:   tftypes.NewValue(secretRefTerraformObjectType(ctx, t, r), nil),
				}),
			}),
		})

		exactlyOne := resourcevalidator.ExactlyOneOf(
			path.MatchRoot(attrAWSBlock),
			path.MatchRoot(attrAzureBlock),
			path.MatchRoot(attrVarsMap),
		)
		resp := &resource.ValidateConfigResponse{}
		exactlyOne.ValidateResource(ctx, resource.ValidateConfigRequest{Config: cfg}, resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}
