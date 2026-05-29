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
	"encoding/json"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestConfigurationValueToWireJSON(t *testing.T) {
	t.Parallel()

	bf, _ := new(big.Float).SetString("-42.5")
	num := fwtypes.NumberValue(bf)

	tests := []struct {
		name    string
		elem    ConfigurationValueModel
		wantRaw string
	}{
		{
			name:    "string",
			elem:    ConfigurationValueModel{String: fwtypes.StringValue("hello")},
			wantRaw: `"hello"`,
		},
		{
			name:    "number",
			elem:    ConfigurationValueModel{Number: num},
			wantRaw: `-42.5`,
		},
		{
			name:    "bool",
			elem:    ConfigurationValueModel{Bool: fwtypes.BoolValue(true)},
			wantRaw: `true`,
		},
		{
			name:    "json",
			elem:    ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(`{"k":"v"}`)},
			wantRaw: `{"k":"v"}`,
		},
		{
			name:    "json array",
			elem:    ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(`[1,2,3]`)},
			wantRaw: `[1,2,3]`,
		},
		{
			name:    "secret_value",
			elem:    ConfigurationValueModel{SecretValue: fwtypes.StringValue("s3cr3t")},
			wantRaw: `"s3cr3t"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, err := configurationValueToWireJSON(tt.elem)
			require.NoError(t, err)
			require.JSONEq(t, tt.wantRaw, string(raw))
		})
	}
}

func TestDecodeConfigurationValueIntoBranch_roundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		branch  string
		raw     string
		checker func(t *testing.T, m ConfigurationValueModel)
	}{
		{
			branch: stringBranchAttrName,
			raw:    `"text"`,
			checker: func(t *testing.T, m ConfigurationValueModel) {
				require.Equal(t, "text", m.String.ValueString())
			},
		},
		{
			branch: numberBranchAttrName,
			raw:    `5432`,
			checker: func(t *testing.T, m ConfigurationValueModel) {
				f, _ := m.Number.ValueBigFloat().Float64()
				require.InEpsilon(t, float64(5432), f, 0.001)
			},
		},
		{
			branch: boolBranchAttrName,
			raw:    `false`,
			checker: func(t *testing.T, m ConfigurationValueModel) {
				require.False(t, m.Bool.ValueBool())
			},
		},
		{
			branch: jsonBranchAttrName,
			raw:    `{"a":1}`,
			checker: func(t *testing.T, m ConfigurationValueModel) {
				require.JSONEq(t, `{"a":1}`, m.JSON.ValueString())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.branch, func(t *testing.T) {
			t.Parallel()
			decoded, diags := decodeConfigurationValueIntoBranch(json.RawMessage(tc.raw), tc.branch)
			require.False(t, diags.HasError())
			tc.checker(t, decoded)

			wire, err := configurationValueToWireJSON(decoded)
			require.NoError(t, err)
			require.JSONEq(t, tc.raw, string(wire))
		})
	}
}

func TestEncodeConfigurationValuesWire_usesConfigSecrets(t *testing.T) {
	t.Parallel()

	planMap := map[string]ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringNull()},
	}
	configMap := map[string]ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("from-config")},
	}

	var diags diag.Diagnostics
	values := encodeConfigurationValuesWire(planMap, configMap, &diags)
	require.False(t, diags.HasError())
	require.JSONEq(t, `"from-config"`, string(values["password"]))
}

func TestSecretHashKey(t *testing.T) {
	t.Parallel()
	require.Equal(
		t,
		`secret_hash:configuration_values["password"].secret_value`,
		secretHashKey("password"),
	)
}
