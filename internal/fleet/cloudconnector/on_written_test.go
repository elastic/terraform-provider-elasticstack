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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnWrittenCloudConnector_WritesHashesAndIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("aws external_id", func(t *testing.T) {
		t.Parallel()
		priv := newMapPrivateState()
		hasher := cloudConnectorHasher()

		config := cloudConnectorModel{
			AWS: mustAWSBlockObject(t, types.StringValue("aws-secret")),
		}

		diags := onWrittenCloudConnector(ctx, nil, cloudConnectorModel{}, config, priv)
		require.False(t, diags.HasError())

		awsHash := priv.data[awsExternalIDPrivateStateKey()]
		require.NotEmpty(t, awsHash)
		assert.True(t, hasher.Matches("aws-secret", awsHash))
	})

	t.Run("vars secret_value", func(t *testing.T) {
		t.Parallel()
		priv := newMapPrivateState()
		hasher := cloudConnectorHasher()

		config := cloudConnectorModel{
			Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
				"token": {
					Type:        types.StringValue("password"),
					SecretValue: types.StringValue("var-secret"),
				},
			}),
		}

		diags := onWrittenCloudConnector(ctx, nil, cloudConnectorModel{}, config, priv)
		require.False(t, diags.HasError())

		varHash := priv.data[varsSecretValuePrivateStateKey("token")]
		require.NotEmpty(t, varHash)
		assert.True(t, hasher.Matches("var-secret", varHash))

		indexBytes := priv.data[varsSecretIndexPrivateStateKey]
		require.NotEmpty(t, indexBytes)
		var indexed []string
		require.NoError(t, json.Unmarshal(indexBytes, &indexed))
		assert.Equal(t, []string{"token"}, indexed)
	})
}

func TestOnWrittenCloudConnector_RemovesStaleVarHashes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	priv := newMapPrivateState()
	hasher := cloudConnectorHasher()

	staleHash, err := hasher.Compute("old-secret")
	require.NoError(t, err)
	priv.data[varsSecretValuePrivateStateKey("stale")] = staleHash
	oldIndex, err := json.Marshal([]string{"stale", "keep"})
	require.NoError(t, err)
	priv.data[varsSecretIndexPrivateStateKey] = oldIndex

	config := cloudConnectorModel{
		Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"keep": {
				Type:        types.StringValue("password"),
				SecretValue: types.StringValue("keep-secret"),
			},
		}),
	}

	diags := onWrittenCloudConnector(ctx, (*clients.KibanaScopedClient)(nil), cloudConnectorModel{}, config, priv)
	require.False(t, diags.HasError())

	_, stalePresent := priv.data[varsSecretValuePrivateStateKey("stale")]
	assert.False(t, stalePresent)
	assert.NotEmpty(t, priv.data[varsSecretValuePrivateStateKey("keep")])

	indexBytes := priv.data[varsSecretIndexPrivateStateKey]
	var indexed []string
	require.NoError(t, json.Unmarshal(indexBytes, &indexed))
	assert.Equal(t, []string{"keep"}, indexed)
}

func TestOnWrittenCloudConnector_RemovesAWSHashWhenUnset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	priv := newMapPrivateState()
	priv.data[awsExternalIDPrivateStateKey()] = []byte("placeholder")

	config := cloudConnectorModel{
		AWS: mustAWSBlockObject(t, types.StringNull()),
	}

	diags := onWrittenCloudConnector(ctx, nil, cloudConnectorModel{}, config, priv)
	require.False(t, diags.HasError())

	_, present := priv.data[awsExternalIDPrivateStateKey()]
	assert.False(t, present)
}

func TestOnWrittenCloudConnector_RemovesIndexWhenVarsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	priv := newMapPrivateState()
	hasher := cloudConnectorHasher()

	staleHash, err := hasher.Compute("old-secret")
	require.NoError(t, err)
	priv.data[varsSecretValuePrivateStateKey("stale")] = staleHash
	oldIndex, err := json.Marshal([]string{"stale"})
	require.NoError(t, err)
	priv.data[varsSecretIndexPrivateStateKey] = oldIndex

	config := cloudConnectorModel{
		Vars: types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}),
	}

	diags := onWrittenCloudConnector(ctx, nil, cloudConnectorModel{}, config, priv)
	require.False(t, diags.HasError())

	_, indexPresent := priv.data[varsSecretIndexPrivateStateKey]
	assert.False(t, indexPresent)
	_, stalePresent := priv.data[varsSecretValuePrivateStateKey("stale")]
	assert.False(t, stalePresent)
}

func TestOnWrittenCloudConnector_CorruptIndexGraceful(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	priv := newMapPrivateState()
	priv.data[varsSecretIndexPrivateStateKey] = []byte(`not-json`)

	config := cloudConnectorModel{
		Vars: mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"token": {
				Type:        types.StringValue("password"),
				SecretValue: types.StringValue("secret"),
			},
		}),
	}

	diags := onWrittenCloudConnector(ctx, nil, cloudConnectorModel{}, config, priv)
	require.False(t, diags.HasError())
	require.NotEmpty(t, diags.Warnings())
	assert.NotEmpty(t, priv.data[varsSecretValuePrivateStateKey("token")])
}

func TestOnWrittenCloudConnector_WarnsWhenPrivateStateUnsupported(t *testing.T) {
	t.Parallel()

	config := cloudConnectorModel{
		AWS: mustAWSBlockObject(t, types.StringValue("secret")),
	}
	diags := onWrittenCloudConnector(context.Background(), nil, cloudConnectorModel{}, config, "unsupported")
	require.False(t, diags.HasError())
	require.NotEmpty(t, diags.Warnings())
}

func TestOnWrittenCloudConnector_NilPrivateStateIsNoOp(t *testing.T) {
	t.Parallel()

	config := cloudConnectorModel{
		AWS: mustAWSBlockObject(t, types.StringValue("secret")),
	}
	diags := onWrittenCloudConnector(context.Background(), nil, cloudConnectorModel{}, config, nil)
	require.False(t, diags.HasError())
}
