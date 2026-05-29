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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileVars_UnionArms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		elem       cloudConnectorVarsElement
		assertFunc func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties)
	}{
		{
			name: "string",
			elem: cloudConnectorVarsElement{String: types.StringValue("us-east-1")},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				raw, err := json.Marshal(wire)
				require.NoError(t, err)
				assert.JSONEq(t, `"us-east-1"`, string(raw))
			},
		},
		{
			name: "number",
			elem: cloudConnectorVarsElement{Number: types.Float64Value(42.5)},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				v, err := wire.AsPostFleetCloudConnectorsJSONBodyVars1()
				require.NoError(t, err)
				assert.InDelta(t, float32(42.5), v, 0.001)
			},
		},
		{
			name: "bool",
			elem: cloudConnectorVarsElement{Bool: types.BoolValue(true)},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				v, err := wire.AsPostFleetCloudConnectorsJSONBodyVars2()
				require.NoError(t, err)
				assert.True(t, v)
			},
		},
		{
			name: "structured value",
			elem: cloudConnectorVarsElement{
				Type:   types.StringValue("text"),
				Frozen: types.BoolValue(true),
				Value:  types.StringValue("arn:aws:iam::123:role/x"),
			},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
				require.NoError(t, err)
				assert.Equal(t, "text", structured.Type)
				require.NotNil(t, structured.Frozen)
				assert.True(t, *structured.Frozen)
				value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
				require.NoError(t, err)
				assert.Equal(t, "arn:aws:iam::123:role/x", value)
			},
		},
		{
			name: "structured secret_value",
			elem: cloudConnectorVarsElement{
				Type:        types.StringValue("password"),
				SecretValue: types.StringValue("super-secret"),
			},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
				require.NoError(t, err)
				assert.Equal(t, "password", structured.Type)
				value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
				require.NoError(t, err)
				assert.Equal(t, "super-secret", value)
			},
		},
		{
			name: "structured frozen false",
			elem: cloudConnectorVarsElement{
				Type:   types.StringValue("text"),
				Frozen: types.BoolValue(false),
				Value:  types.StringValue("plain"),
			},
			assertFunc: func(t *testing.T, wire kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties) {
				t.Helper()
				structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
				require.NoError(t, err)
				require.NotNil(t, structured.Frozen)
				assert.False(t, *structured.Frozen)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
				"key": tc.elem,
			}), nil, false)
			require.False(t, diags.HasError(), diags)
			tc.assertFunc(t, vars["key"])
		})
	}
}

func TestCompileVars_UpdatePreservesSecretRef(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("secret-ref-1"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	plan := mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"external_id": {
			Type: types.StringValue("password"),
		},
	})
	prior := mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"external_id": {
			Type:      types.StringValue("password"),
			SecretRef: secretRef,
		},
	})

	vars, diags := compileVars(plan, &prior, true)
	require.False(t, diags.HasError(), diags)

	structured, err := vars["external_id"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	ref, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value1()
	require.NoError(t, err)
	assert.Equal(t, "secret-ref-1", ref.Id)
	assert.True(t, ref.IsSecretRef)
}

func TestCompileStructuredVarElement_TypeFallbackFromPrior(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("secret-ref-1"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	planElem := cloudConnectorVarsElement{
		Type: types.StringNull(),
	}
	priorElem := cloudConnectorVarsElement{
		Type:      types.StringValue("password"),
		SecretRef: secretRef,
	}

	wire, diags := compileStructuredVarElement(planElem, &priorElem, true, "external_id")
	require.False(t, diags.HasError(), diags)

	structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	assert.Equal(t, "password", structured.Type)
}

func TestCompileStructuredVarElement_NewSecretOverridesPriorRef(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("old-ref"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	planElem := cloudConnectorVarsElement{
		Type:        types.StringValue("password"),
		SecretValue: types.StringValue("new-secret"),
	}
	priorElem := cloudConnectorVarsElement{
		Type:      types.StringValue("password"),
		SecretRef: secretRef,
	}

	wire, diags := compileStructuredVarElement(planElem, &priorElem, true, "token")
	require.False(t, diags.HasError(), diags)

	structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "new-secret", value)
}

func TestCompileAWS_Create(t *testing.T) {
	vars, diags := compileAWS(awsBlockModel{
		RoleArn:    types.StringValue("arn:aws:iam::123:role/x"),
		ExternalID: types.StringValue("ext-id"),
	}, nil, false, nil)
	require.False(t, diags.HasError(), diags)
	require.Len(t, vars, 2)

	roleArn, err := vars["role_arn"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	assert.Equal(t, "text", roleArn.Type)
	roleValue, err := roleArn.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:iam::123:role/x", roleValue)

	externalID, err := vars["external_id"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	assert.Equal(t, "password", externalID.Type)
	extValue, err := externalID.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "ext-id", extValue)
}

func TestCompileAWS_CreateMissingExternalID(t *testing.T) {
	_, diags := compileAWS(awsBlockModel{
		RoleArn:    types.StringValue("arn:aws:iam::123:role/x"),
		ExternalID: types.StringNull(),
	}, nil, false, nil)
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing AWS external_id", diags.Errors()[0].Summary())
}

func TestCompileAWS_UpdatePreservesExternalIDSecretRef(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("secret-ref-aws"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	vars, diags := compileAWS(
		awsBlockModel{
			RoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
			ExternalID:          types.StringNull(),
			ExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
		},
		&awsBlockModel{
			RoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
			ExternalID:          types.StringNull(),
			ExternalIDSecretRef: secretRef,
		},
		true,
		nil,
	)
	require.False(t, diags.HasError(), diags)

	externalID, err := vars["external_id"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	ref, err := externalID.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value1()
	require.NoError(t, err)
	assert.Equal(t, "secret-ref-aws", ref.Id)
	assert.True(t, ref.IsSecretRef)
}

func TestCompileAWSExternalID_NewSecretOverridesPriorRef(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("old-ref"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	wire, diags := compileAWSExternalID(
		awsBlockModel{
			ExternalID: types.StringValue("new-secret"),
		},
		&awsBlockModel{
			ExternalIDSecretRef: secretRef,
		},
		true,
		map[string]struct{}{writeOnlyAttributeAWSExternalID: {}},
	)
	require.False(t, diags.HasError(), diags)

	structured, err := wire.AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "new-secret", value)
}

func TestCompileAzure(t *testing.T) {
	vars, diags := compileAzure(azureBlockModel{
		TenantID:         types.StringValue("tenant-1"),
		ClientID:         types.StringValue("client-1"),
		CloudConnectorID: types.StringValue("connector-1"),
	}, nil, false, nil)
	require.False(t, diags.HasError(), diags)
	require.Len(t, vars, 3)

	for key, expected := range map[string]string{
		"tenant_id":                             "tenant-1",
		"client_id":                             "client-1",
		wireKeyAzureCredentialsCloudConnectorID: "connector-1",
	} {
		structured, err := vars[key].AsPostFleetCloudConnectorsJSONBodyVars3()
		require.NoError(t, err, key)
		if key == wireKeyAzureCredentialsCloudConnectorID {
			assert.Equal(t, "text", structured.Type)
		} else {
			assert.Equal(t, "password", structured.Type)
		}
		value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
		require.NoError(t, err, key)
		assert.Equal(t, expected, value)
	}
}

func TestCompileVarsForWrite_VarsOnlyUpdateUsesVarsPath(t *testing.T) {
	awsObj, objDiags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringNull(),
		attrAWSExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	})
	require.False(t, objDiags.HasError())

	config := cloudConnectorModel{
		Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"custom_key": {String: types.StringValue("updated")},
		}),
		AWS: types.ObjectNull(awsAttrTypes()),
	}
	plan := cloudConnectorModel{
		Vars: config.Vars,
		AWS:  awsObj,
	}

	vars, diags := plan.compileVarsForWrite(config, nil, true, nil)
	require.False(t, diags.HasError(), diags)
	require.Contains(t, vars, "custom_key")
	assert.Len(t, vars, 1)
}

func TestCompileVarsForWrite_AWSConfigUsesAWSPath(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("secret-ref-1"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	awsObj, objDiags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringNull(),
		attrAWSExternalIDSecretRef: secretRef,
	})
	require.False(t, objDiags.HasError())

	config := cloudConnectorModel{AWS: awsObj}
	plan := cloudConnectorModel{
		AWS: awsObj,
		Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"custom_key": {String: types.StringValue("from-read")},
		}),
	}
	prior := cloudConnectorModel{AWS: awsObj}

	vars, diags := plan.compileVarsForWrite(config, &prior, true, nil)
	require.False(t, diags.HasError(), diags)
	require.Len(t, vars, 2)
	assert.Contains(t, vars, "role_arn")
	assert.Contains(t, vars, "external_id")
	assert.NotContains(t, vars, "custom_key")
}

func TestToAPIUpdateBody_OmitsCloudProvider(t *testing.T) {
	secretRef, refDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue("secret-ref-1"),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, refDiags.HasError())

	awsObj, objDiags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringNull(),
		attrAWSExternalIDSecretRef: secretRef,
	})
	require.False(t, objDiags.HasError())

	config := cloudConnectorModel{AWS: awsObj}
	plan := cloudConnectorModel{
		Name:          types.StringValue("updated-name"),
		CloudProvider: types.StringValue("aws"),
		AccountType:   types.StringValue(accountTypeSingleAccount),
		AWS:           awsObj,
	}
	prior := cloudConnectorModel{AWS: awsObj}

	body, diags := plan.toAPIUpdateBody(config, prior, nil)
	require.False(t, diags.HasError(), diags)

	require.NotNil(t, body.Name)
	assert.Equal(t, "updated-name", *body.Name)
	require.NotNil(t, body.Vars)
	require.NotNil(t, body.AccountType)
	assert.Equal(t, kbapi.PutFleetCloudConnectorsCloudconnectoridJSONBodyAccountType(accountTypeSingleAccount), *body.AccountType)

	externalID := (*body.Vars)["external_id"]
	structured, err := externalID.AsPutFleetCloudConnectorsCloudconnectoridJSONBodyVars3()
	require.NoError(t, err)
	ref, err := structured.Value.AsPutFleetCloudConnectorsCloudconnectoridJSONBodyVars3Value1()
	require.NoError(t, err)
	assert.Equal(t, "secret-ref-1", ref.Id)
}

func TestToAPIUpdateBody_resubmitWriteOnlySecret(t *testing.T) {
	t.Parallel()

	const newSecret = "rotated-secret"
	awsObj, diags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringNull(),
		attrAWSExternalIDSecretRef: mustSecretRefObject(t, "secret-ref-1"),
	})
	require.False(t, diags.HasError())

	configAWS, diags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringValue(newSecret),
		attrAWSExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	})
	require.False(t, diags.HasError())

	plan := cloudConnectorModel{
		Name:          types.StringValue("same-name"),
		CloudProvider: types.StringValue("aws"),
		AccountType:   types.StringValue(accountTypeSingleAccount),
		AWS:           awsObj,
	}
	prior := plan
	config := plan
	config.AWS = configAWS

	resubmit := map[string]struct{}{writeOnlyAttributeAWSExternalID: {}}
	body, bodyDiags := plan.toAPIUpdateBody(config, prior, resubmit)
	require.False(t, bodyDiags.HasError(), bodyDiags)

	externalID := (*body.Vars)["external_id"]
	structured, err := externalID.AsPutFleetCloudConnectorsCloudconnectoridJSONBodyVars3()
	require.NoError(t, err)
	plain, err := structured.Value.AsPutFleetCloudConnectorsCloudconnectoridJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, newSecret, plain)
}

func mustVarsMap(t *testing.T, elems map[string]cloudConnectorVarsElement) types.Map {
	t.Helper()

	values := make(map[string]attr.Value, len(elems))
	for key, elem := range elems {
		obj, diags := varsElementToObject(normalizeVarsElement(elem))
		require.False(t, diags.HasError(), diags)
		values[key] = obj
	}

	vars, diags := types.MapValue(types.ObjectType{AttrTypes: varsElementAttrTypes()}, values)
	require.False(t, diags.HasError(), diags)
	return vars
}

func normalizeVarsElement(elem cloudConnectorVarsElement) cloudConnectorVarsElement {
	nullElem := cloudConnectorVarsElement{
		String:      types.StringNull(),
		Number:      types.Float64Null(),
		Bool:        types.BoolNull(),
		Type:        types.StringNull(),
		Frozen:      types.BoolNull(),
		Value:       types.StringNull(),
		SecretValue: types.StringNull(),
		SecretRef:   types.ObjectNull(secretRefAttrTypes()),
	}

	if !elem.String.IsNull() && !elem.String.IsUnknown() {
		nullElem.String = elem.String
	}
	if !elem.Number.IsNull() && !elem.Number.IsUnknown() {
		nullElem.Number = elem.Number
	}
	if !elem.Bool.IsNull() && !elem.Bool.IsUnknown() {
		nullElem.Bool = elem.Bool
	}
	if !elem.Type.IsNull() && !elem.Type.IsUnknown() {
		nullElem.Type = elem.Type
	}
	if !elem.Frozen.IsNull() && !elem.Frozen.IsUnknown() {
		nullElem.Frozen = elem.Frozen
	}
	if !elem.Value.IsNull() && !elem.Value.IsUnknown() {
		nullElem.Value = elem.Value
	}
	if !elem.SecretValue.IsNull() && !elem.SecretValue.IsUnknown() {
		nullElem.SecretValue = elem.SecretValue
	}
	if !elem.SecretRef.IsNull() && !elem.SecretRef.IsUnknown() {
		nullElem.SecretRef = elem.SecretRef
	}

	return nullElem
}

func TestAugmentInUseConflictDiagnostic(t *testing.T) {
	t.Run("includes count when package_policy_count present", func(t *testing.T) {
		diags := diag.Diagnostics{
			diag.NewErrorDiagnostic("Delete failed", `{"package_policy_count":3}`),
		}
		out := augmentInUseConflictDiagnostic(diags)
		require.Len(t, out, 2)
		assert.Contains(t, out[1].Detail(), "3 package policies")
		assert.Contains(t, out[1].Detail(), "force_delete = true")
	})

	t.Run("singular policy wording", func(t *testing.T) {
		diags := diag.Diagnostics{
			diag.NewErrorDiagnostic("Delete failed", `{"package_policy_count":1}`),
		}
		out := augmentInUseConflictDiagnostic(diags)
		require.Len(t, out, 2)
		assert.Contains(t, out[1].Detail(), "1 package policy")
	})

	t.Run("unchanged when package_policy_count absent", func(t *testing.T) {
		diags := diag.Diagnostics{
			diag.NewErrorDiagnostic("Delete failed", "some other error"),
		}
		out := augmentInUseConflictDiagnostic(diags)
		require.Len(t, out, 1)
	})
}
