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

package resource

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorfieldtype"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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
		elem    connector.ConfigurationValueModel
		wantRaw string
	}{
		{
			name:    "string",
			elem:    connector.ConfigurationValueModel{String: fwtypes.StringValue("hello")},
			wantRaw: `"hello"`,
		},
		{
			name:    "number",
			elem:    connector.ConfigurationValueModel{Number: num},
			wantRaw: `-42.5`,
		},
		{
			name:    "bool",
			elem:    connector.ConfigurationValueModel{Bool: fwtypes.BoolValue(true)},
			wantRaw: `true`,
		},
		{
			name:    "json",
			elem:    connector.ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(`{"k":"v"}`)},
			wantRaw: `{"k":"v"}`,
		},
		{
			name:    "json array",
			elem:    connector.ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(`[1,2,3]`)},
			wantRaw: `[1,2,3]`,
		},
		{
			name:    "secret_value",
			elem:    connector.ConfigurationValueModel{SecretValue: fwtypes.StringValue("s3cr3t")},
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
		checker func(t *testing.T, m connector.ConfigurationValueModel)
	}{
		{
			branch: connector.StringBranchAttr,
			raw:    `"text"`,
			checker: func(t *testing.T, m connector.ConfigurationValueModel) {
				require.Equal(t, "text", m.String.ValueString())
			},
		},
		{
			branch: connector.NumberBranchAttr,
			raw:    `5432`,
			checker: func(t *testing.T, m connector.ConfigurationValueModel) {
				f, _ := m.Number.ValueBigFloat().Float64()
				require.InEpsilon(t, float64(5432), f, 0.001)
			},
		},
		{
			branch: connector.BoolBranchAttr,
			raw:    `false`,
			checker: func(t *testing.T, m connector.ConfigurationValueModel) {
				require.False(t, m.Bool.ValueBool())
			},
		},
		{
			branch: connector.JSONBranchAttr,
			raw:    `{"a":1}`,
			checker: func(t *testing.T, m connector.ConfigurationValueModel) {
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

	planMap := map[string]connector.ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringNull()},
	}
	configMap := map[string]connector.ConfigurationValueModel{
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

func TestSchemaTypeToBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fieldTyp *connectorfieldtype.ConnectorFieldType
		want     string
	}{
		{name: "nil", fieldTyp: nil, want: connector.StringBranchAttr},
		{name: "str", fieldTyp: new(connectorfieldtype.Str), want: connector.StringBranchAttr},
		{name: "int", fieldTyp: new(connectorfieldtype.Int), want: connector.NumberBranchAttr},
		{name: "bool", fieldTyp: new(connectorfieldtype.Bool), want: connector.BoolBranchAttr},
		{name: "list", fieldTyp: new(connectorfieldtype.List), want: connector.StringBranchAttr},
		{name: "unknown", fieldTyp: &connectorfieldtype.ConnectorFieldType{Name: "other"}, want: connector.StringBranchAttr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, schemaTypeToBranch(tt.fieldTyp))
		})
	}
}

func TestActiveConfigurationBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		elem connector.ConfigurationValueModel
		want string
	}{
		{name: "string", elem: connector.ConfigurationValueModel{String: fwtypes.StringValue("a")}, want: connector.StringBranchAttr},
		{name: "number", elem: connector.ConfigurationValueModel{Number: fwtypes.NumberValue(big.NewFloat(1))}, want: connector.NumberBranchAttr},
		{name: "bool", elem: connector.ConfigurationValueModel{Bool: fwtypes.BoolValue(true)}, want: connector.BoolBranchAttr},
		{name: "json", elem: connector.ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(`{}`)}, want: connector.JSONBranchAttr},
		{name: "secret", elem: connector.ConfigurationValueModel{SecretValue: fwtypes.StringValue("s")}, want: connector.SecretValueBranchAttr},
		{name: "secret unknown", elem: connector.ConfigurationValueModel{SecretValue: fwtypes.StringUnknown()}, want: connector.SecretValueBranchAttr},
		{name: "none", elem: connector.ConfigurationValueModel{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, activeConfigurationBranch(tt.elem))
		})
	}
}

func TestConfigurationValueToWireJSON_edgeCases(t *testing.T) {
	t.Parallel()

	bf, _, _ := big.ParseFloat("2.5", 10, 64, big.ToNearestEven)
	raw, err := configurationValueToWireJSON(connector.ConfigurationValueModel{Number: fwtypes.NumberValue(bf)})
	require.NoError(t, err)
	require.JSONEq(t, `2.5`, string(raw))

	_, err = configurationValueToWireJSON(connector.ConfigurationValueModel{})
	require.Error(t, err)
}

func TestEncodeConfigurationValuesWire_emptyPlan(t *testing.T) {
	t.Parallel()
	var diags diag.Diagnostics
	values := encodeConfigurationValuesWire(map[string]connector.ConfigurationValueModel{}, nil, &diags)
	require.Nil(t, values)
	require.False(t, diags.HasError())
}

func TestEncodeConfigurationValuesWire_invalidElement(t *testing.T) {
	t.Parallel()
	var diags diag.Diagnostics
	values := encodeConfigurationValuesWire(map[string]connector.ConfigurationValueModel{
		"bad": {},
	}, nil, &diags)
	require.Nil(t, values)
	require.True(t, diags.HasError())
}

func TestDecodeConfigurationValueIntoBranch_errors(t *testing.T) {
	t.Parallel()

	_, diags := decodeConfigurationValueIntoBranch(json.RawMessage(`not-json`), connector.StringBranchAttr)
	require.True(t, diags.HasError())

	_, diags = decodeConfigurationValueIntoBranch(json.RawMessage(`"x"`), "unsupported")
	require.True(t, diags.HasError())
}

func TestPopulateConfigurationValuesFromAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("prior branch preserved string decodes API number as string", func(t *testing.T) {
		t.Parallel()
		resp := testConnectorResponse(
			map[string]json.RawMessage{"a": json.RawMessage(`1`)},
			map[string]estypes.ConnectorConfigProperties{
				"a": {Value: json.RawMessage(`1`), Type: new(connectorfieldtype.Int)},
			},
		)
		priorMap := map[string]connector.ConfigurationValueModel{
			"a": {String: fwtypes.StringValue("old")},
		}
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, priorMap, &diags)
		require.False(t, diags.HasError())
		decoded := mapFromResult(ctx, result, &diags)
		require.Equal(t, "1", decoded["a"].String.ValueString())
	})

	t.Run("secret branch preserved from prior", func(t *testing.T) {
		t.Parallel()
		priorElem := connector.ConfigurationValueModel{SecretValue: fwtypes.StringValue("ignored-in-state")}
		resp := testConnectorResponse(
			map[string]json.RawMessage{"password": json.RawMessage(`"redacted"`)},
			map[string]estypes.ConnectorConfigProperties{
				"password": {Value: json.RawMessage(`"redacted"`)},
			},
		)
		priorMap := map[string]connector.ConfigurationValueModel{"password": priorElem}
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, priorMap, &diags)
		require.False(t, diags.HasError())
		decoded := mapFromResult(ctx, result, &diags)
		require.Equal(t, priorElem, decoded["password"])
	})

	t.Run("sensitive non-secret branch warns and preserves prior", func(t *testing.T) {
		t.Parallel()
		priorElem := connector.ConfigurationValueModel{String: fwtypes.StringValue("pw")}
		resp := testConnectorResponse(
			map[string]json.RawMessage{"password": json.RawMessage(`"x"`)},
			map[string]estypes.ConnectorConfigProperties{
				"password": {Value: json.RawMessage(`"x"`), Sensitive: true},
			},
		)
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, map[string]connector.ConfigurationValueModel{"password": priorElem}, &diags)
		require.False(t, diags.HasError())
		require.NotEmpty(t, diags.Warnings())
		decoded := mapFromResult(ctx, result, &diags)
		require.Equal(t, priorElem, decoded["password"])
	})

	t.Run("import skips sensitive keys", func(t *testing.T) {
		t.Parallel()
		resp := testConnectorResponse(
			map[string]json.RawMessage{
				"x":           json.RawMessage(`"visible"`),
				"sensitive_z": json.RawMessage(`"hidden"`),
			},
			map[string]estypes.ConnectorConfigProperties{
				"x":           {Value: json.RawMessage(`"visible"`), Type: new(connectorfieldtype.Str)},
				"sensitive_z": {Value: json.RawMessage(`"hidden"`), Sensitive: true},
			},
		)
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, nil, &diags)
		require.False(t, diags.HasError())
		decoded := mapFromResult(ctx, result, &diags)
		require.Contains(t, decoded, "x")
		require.NotContains(t, decoded, "sensitive_z")
	})

	t.Run("REQ-008 removal drops API-only keys when priorMap set", func(t *testing.T) {
		t.Parallel()
		resp := testConnectorResponse(
			map[string]json.RawMessage{
				"a": json.RawMessage(`"va"`),
				"b": json.RawMessage(`"vb"`),
				"c": json.RawMessage(`"vc"`),
			},
			map[string]estypes.ConnectorConfigProperties{
				"a": {Value: json.RawMessage(`"va"`)},
				"b": {Value: json.RawMessage(`"vb"`)},
				"c": {Value: json.RawMessage(`"vc"`)},
			},
		)
		priorMap := map[string]connector.ConfigurationValueModel{
			"a": {String: fwtypes.StringValue("a")},
			"b": {String: fwtypes.StringValue("b")},
		}
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, priorMap, &diags)
		require.False(t, diags.HasError())
		decoded := mapFromResult(ctx, result, &diags)
		require.Contains(t, decoded, "a")
		require.Contains(t, decoded, "b")
		require.NotContains(t, decoded, "c")
	})

	t.Run("prior key absent from API is dropped from state", func(t *testing.T) {
		t.Parallel()
		resp := testConnectorResponse(
			map[string]json.RawMessage{"a": json.RawMessage(`"va"`)},
			map[string]estypes.ConnectorConfigProperties{
				"a": {Value: json.RawMessage(`"va"`)},
			},
		)
		priorMap := map[string]connector.ConfigurationValueModel{
			"a": {String: fwtypes.StringValue("a")},
			"b": {String: fwtypes.StringValue("removed")},
		}
		var diags diag.Diagnostics
		result := populateConfigurationValuesFromAPI(ctx, resp, priorMap, &diags)
		decoded := mapFromResult(ctx, result, &diags)
		require.Contains(t, decoded, "a")
		require.NotContains(t, decoded, "b")
	})
}

func testConnectorResponse(
	values map[string]json.RawMessage,
	schema map[string]estypes.ConnectorConfigProperties,
) *getconnector.Response {
	cfg := make(estypes.ConnectorConfiguration)
	for key, raw := range values {
		props := schema[key]
		props.Value = raw
		cfg[key] = props
	}
	return &getconnector.Response{Configuration: cfg}
}

func mapFromResult(ctx context.Context, result fwtypes.Map, diags *diag.Diagnostics) map[string]connector.ConfigurationValueModel {
	if result.IsNull() {
		return map[string]connector.ConfigurationValueModel{}
	}
	return typeutils.MapTypeAs[connector.ConfigurationValueModel](ctx, result, configurationValuesPath, diags)
}

func TestConfigurationValuePresent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  json.RawMessage
		want bool
	}{
		{name: "absent", raw: nil, want: false},
		{name: "empty", raw: json.RawMessage{}, want: false},
		{name: "null", raw: json.RawMessage("null"), want: false},
		{name: "string", raw: json.RawMessage(`"x"`), want: true},
		{name: "number", raw: json.RawMessage(`42`), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, configurationValuePresent(tt.raw))
		})
	}
}
