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
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mapPrivateState struct {
	data map[string][]byte
}

func newMapPrivateState() *mapPrivateState {
	return &mapPrivateState{data: make(map[string][]byte)}
}

func (m *mapPrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return nil, nil
}

func (m *mapPrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if value == nil {
		delete(m.data, key)
	} else {
		m.data[key] = value
	}
	return nil
}

func TestCloudConnectorPrivateStateKeys(t *testing.T) {
	t.Parallel()

	require.Equal(t, "secret_hash:aws.external_id", awsExternalIDPrivateStateKey())
	require.Equal(t, "secret_hash:vars.external_id.secret_value", varsSecretValuePrivateStateKey("external_id"))
	require.Equal(t, "vars.external_id.secret_value", varsSecretValueAttributePath("external_id"))
}

func TestDetectWriteOnlyDrift(t *testing.T) {
	t.Parallel()

	hasher := cloudConnectorHasher()
	secret := "super-secret"
	hash, err := hasher.Compute(secret)
	require.NoError(t, err)

	t.Run("no drift when hash matches config", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringValue(secret), hash)
		assert.False(t, result.Changed)
	})

	t.Run("drift when hash does not match config", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringValue("other-secret"), hash)
		require.True(t, result.Changed)
		assert.Equal(t, writeOnlyAttributeAWSExternalID, result.AttributePath)
		assert.False(t, result.IsImportBaseline)
	})

	t.Run("import baseline when config set and no stored hash", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringValue(secret), nil)
		require.True(t, result.Changed)
		assert.True(t, result.IsImportBaseline)
	})

	t.Run("drift when config null and stored hash exists", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringNull(), hash)
		require.True(t, result.Changed)
		assert.False(t, result.IsImportBaseline)
	})

	t.Run("no drift when config null and no stored hash", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringNull(), nil)
		assert.False(t, result.Changed)
	})

	t.Run("config unknown is treated as unset", func(t *testing.T) {
		t.Parallel()
		result := detectWriteOnlyDrift(hasher, writeOnlyAttributeAWSExternalID, types.StringUnknown(), hash)
		require.True(t, result.Changed)
	})
}

func TestEvaluateWriteOnlyDrift_AWSExternalID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hasher := cloudConnectorHasher()
	priv := newMapPrivateState()

	secretA := "secret-a"
	secretB := "secret-b"
	hashA, err := hasher.Compute(secretA)
	require.NoError(t, err)

	awsObj := mustAWSBlockObject(t, types.StringValue(secretA))
	config := cloudConnectorModel{AWS: awsObj}
	priv.data[awsExternalIDPrivateStateKey()] = hashA

	results, diags := evaluateWriteOnlyDrift(ctx, hasher, config, priv)
	require.False(t, diags.HasError())
	require.Empty(t, results)

	config.AWS = mustAWSBlockObject(t, types.StringValue(secretB))
	results, diags = evaluateWriteOnlyDrift(ctx, hasher, config, priv)
	require.False(t, diags.HasError())
	require.Len(t, results, 1)
	assert.Equal(t, writeOnlyAttributeAWSExternalID, results[0].AttributePath)
	assert.False(t, results[0].IsImportBaseline)
}

func TestEvaluateWriteOnlyDrift_VarsKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hasher := cloudConnectorHasher()

	t.Run("multiple vars keys are independent", func(t *testing.T) {
		t.Parallel()
		localPriv := newMapPrivateState()
		hashK1, err := hasher.Compute("k1-secret")
		require.NoError(t, err)
		localPriv.data[varsSecretValuePrivateStateKey("k1")] = hashK1
		index, err := json.Marshal([]string{"k1"})
		require.NoError(t, err)
		localPriv.data[varsSecretIndexPrivateStateKey] = index

		cfg := cloudConnectorModel{
			Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
				"k1": {
					Type:        types.StringValue("password"),
					SecretValue: types.StringValue("k1-secret"),
				},
				"k2": {
					Type:        types.StringValue("password"),
					SecretValue: types.StringValue("k2-secret"),
				},
			}),
		}
		driftResults, driftDiags := evaluateWriteOnlyDrift(ctx, hasher, cfg, localPriv)
		require.False(t, driftDiags.HasError())
		require.Len(t, driftResults, 1)
		assert.Equal(t, varsSecretValueAttributePath("k2"), driftResults[0].AttributePath)
		assert.True(t, driftResults[0].IsImportBaseline)
	})

	t.Run("removed var key with stored hash triggers drift", func(t *testing.T) {
		t.Parallel()
		localPriv := newMapPrivateState()
		hashRemoved, err := hasher.Compute("removed-secret")
		require.NoError(t, err)
		localPriv.data[varsSecretValuePrivateStateKey("removed")] = hashRemoved
		index, err := json.Marshal([]string{"removed"})
		require.NoError(t, err)
		localPriv.data[varsSecretIndexPrivateStateKey] = index

		cfg := cloudConnectorModel{
			Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{}),
		}
		driftResults, driftDiags := evaluateWriteOnlyDrift(ctx, hasher, cfg, localPriv)
		require.False(t, driftDiags.HasError())
		require.Len(t, driftResults, 1)
		assert.Equal(t, varsSecretValueAttributePath("removed"), driftResults[0].AttributePath)
	})
}

func mustAWSBlockObject(t *testing.T, externalID types.String) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          externalID,
		attrAWSExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	})
	require.False(t, diags.HasError())
	return obj
}

func TestDriftWarningDiagnostic(t *testing.T) {
	t.Parallel()

	normal := driftWarningDiagnostic(driftResult{AttributePath: writeOnlyAttributeAWSExternalID})
	assert.Contains(t, normal.Summary(), writeOnlyAttributeAWSExternalID)
	assert.NotContains(t, normal.Summary(), "super-secret")

	baseline := driftWarningDiagnostic(driftResult{
		AttributePath:    writeOnlyAttributeAWSExternalID,
		IsImportBaseline: true,
	})
	assert.Contains(t, baseline.Detail(), "import")
}
