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

func TestCompileVars_StringArm(t *testing.T) {
	vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"region": {String: types.StringValue("us-east-1")},
	}), nil, false)
	require.False(t, diags.HasError(), diags)

	raw, err := json.Marshal(vars["region"])
	require.NoError(t, err)
	assert.JSONEq(t, `"us-east-1"`, string(raw))
}

func TestCompileVars_NumberArm(t *testing.T) {
	vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"count": {Number: types.Float64Value(42.5)},
	}), nil, false)
	require.False(t, diags.HasError(), diags)

	v, err := vars["count"].AsPostFleetCloudConnectorsJSONBodyVars1()
	require.NoError(t, err)
	assert.InDelta(t, float32(42.5), v, 0.001)
}

func TestCompileVars_BoolArm(t *testing.T) {
	vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"enabled": {Bool: types.BoolValue(true)},
	}), nil, false)
	require.False(t, diags.HasError(), diags)

	v, err := vars["enabled"].AsPostFleetCloudConnectorsJSONBodyVars2()
	require.NoError(t, err)
	assert.True(t, v)
}

func TestCompileVars_StructuredStringValueArm(t *testing.T) {
	frozen := types.BoolValue(true)
	vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"role_arn": {
			Type:   types.StringValue("text"),
			Frozen: frozen,
			Value:  types.StringValue("arn:aws:iam::123:role/x"),
		},
	}), nil, false)
	require.False(t, diags.HasError(), diags)

	structured, err := vars["role_arn"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	assert.Equal(t, "text", structured.Type)
	require.NotNil(t, structured.Frozen)
	assert.True(t, *structured.Frozen)
	value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:iam::123:role/x", value)
}

func TestCompileVars_StructuredSecretValueArm(t *testing.T) {
	vars, diags := compileVars(mustVarsMap(t, map[string]cloudConnectorVarsElement{
		"token": {
			Type:        types.StringValue("password"),
			SecretValue: types.StringValue("super-secret"),
		},
	}), nil, false)
	require.False(t, diags.HasError(), diags)

	structured, err := vars["token"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	assert.Equal(t, "password", structured.Type)
	value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	require.NoError(t, err)
	assert.Equal(t, "super-secret", value)
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

func TestCompileAWS_Create(t *testing.T) {
	vars, diags := compileAWS(awsBlockModel{
		RoleArn:    types.StringValue("arn:aws:iam::123:role/x"),
		ExternalID: types.StringValue("ext-id"),
	}, nil, false)
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
	)
	require.False(t, diags.HasError(), diags)

	externalID, err := vars["external_id"].AsPostFleetCloudConnectorsJSONBodyVars3()
	require.NoError(t, err)
	ref, err := externalID.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value1()
	require.NoError(t, err)
	assert.Equal(t, "secret-ref-aws", ref.Id)
	assert.True(t, ref.IsSecretRef)
}

func TestCompileAzure(t *testing.T) {
	vars, diags := compileAzure(azureBlockModel{
		TenantID:         types.StringValue("tenant-1"),
		ClientID:         types.StringValue("client-1"),
		CloudConnectorID: types.StringValue("connector-1"),
	})
	require.False(t, diags.HasError(), diags)
	require.Len(t, vars, 3)

	for key, expected := range map[string]string{
		"tenant_id":          "tenant-1",
		"client_id":          "client-1",
		"cloud_connector_id": "connector-1",
	} {
		structured, err := vars[key].AsPostFleetCloudConnectorsJSONBodyVars3()
		require.NoError(t, err, key)
		assert.Equal(t, "text", structured.Type)
		value, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
		require.NoError(t, err, key)
		assert.Equal(t, expected, value)
	}
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

	plan := cloudConnectorModel{
		Name:          types.StringValue("updated-name"),
		CloudProvider: types.StringValue("aws"),
		AccountType:   types.StringValue(accountTypeSingleAccount),
		AWS:           awsObj,
	}

	prior := cloudConnectorModel{
		AWS: awsObj,
	}

	body, diags := plan.toAPIUpdateBody(prior)
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
	t.Run("adds hint when package_policy_count present", func(t *testing.T) {
		diags := diag.Diagnostics{
			diag.NewErrorDiagnostic("Delete failed", `{"package_policy_count":3}`),
		}
		out := augmentInUseConflictDiagnostic(diags)
		require.Len(t, out, 2)
		assert.Contains(t, out[1].Detail(), "force_delete = true")
	})

	t.Run("unchanged when package_policy_count absent", func(t *testing.T) {
		diags := diag.Diagnostics{
			diag.NewErrorDiagnostic("Delete failed", "some other error"),
		}
		out := augmentInUseConflictDiagnostic(diags)
		require.Len(t, out, 1)
	})
}
