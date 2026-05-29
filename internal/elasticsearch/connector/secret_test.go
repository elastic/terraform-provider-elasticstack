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

package connector

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type mapPrivateState map[string][]byte

func (m mapPrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if v, ok := m[key]; ok {
		return v, nil
	}
	return nil, nil
}

func (m mapPrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if value == nil {
		delete(m, key)
	} else {
		m[key] = value
	}
	return nil
}

var _ entitycore.PrivateStateStorage = mapPrivateState{}

func TestStoreSecretHashes_nilPrivateWithSecrets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	configMap := map[string]ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("pw")},
	}
	var diags diag.Diagnostics
	storeSecretHashes(ctx, nil, configMap, &diags)
	require.True(t, diags.HasError())
	require.Equal(t, privateStateUnavailableSummary, diags.Errors()[0].Summary())
}

func TestStoreSecretHashes_nilPrivateNoSecrets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics
	storeSecretHashes(ctx, nil, map[string]ConfigurationValueModel{"host": {String: fwtypes.StringValue("x")}}, &diags)
	require.False(t, diags.HasError())
}

func TestEncodeSecretHashForPrivateState_validJSON(t *testing.T) {
	t.Parallel()

	hash, err := secretHasher.Compute("pw")
	require.NoError(t, err)

	encoded, err := encodeSecretHashForPrivateState(hash)
	require.NoError(t, err)
	require.True(t, json.Valid(encoded))

	decoded, err := decodeSecretHashFromPrivateState(encoded)
	require.NoError(t, err)
	require.Equal(t, hash, decoded)
}

func TestSecretHashKey_usesSpecBracketedPath(t *testing.T) {
	t.Parallel()
	require.Equal(t, `secret_hash:configuration_values["password"].secret_value`, secretHashKey("password"))
}

func TestStoreSecretHashes_storesJSONEncodedHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ps := mapPrivateState{}
	configMap := map[string]ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("pw")},
	}
	var diags diag.Diagnostics
	storeSecretHashes(ctx, ps, configMap, &diags)
	require.False(t, diags.HasError())
	require.True(t, json.Valid(ps[secretHashKey("password")]))
}

func TestClearRemovedSecretHashes_nilPrivateWithRemovals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	priorMap := map[string]ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("x")},
	}
	var diags diag.Diagnostics
	clearRemovedSecretHashes(ctx, nil, priorMap, map[string]ConfigurationValueModel{}, &diags)
	require.True(t, diags.HasError())
}
