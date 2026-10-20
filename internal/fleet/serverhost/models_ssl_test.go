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

package serverhost

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sslObjectValue(t *testing.T, certificate, key, clientAuth *string, cas []string) types.Object {
	t.Helper()

	casValue := types.ListNull(types.StringType)
	if cas != nil {
		elems := make([]attr.Value, 0, len(cas))
		for _, ca := range cas {
			elems = append(elems, types.StringValue(ca))
		}
		casValue = types.ListValueMust(types.StringType, elems)
	}

	obj, diags := types.ObjectValue(getSslAttrTypes(context.Background()), map[string]attr.Value{
		"certificate":             types.StringPointerValue(certificate),
		"key":                     types.StringPointerValue(key),
		"certificate_authorities": casValue,
		"client_auth":             types.StringPointerValue(clientAuth),
	})
	require.False(t, diags.HasError(), "%v", diags)
	return obj
}

func baseSSLModel(ssl types.Object) serverHostModel {
	return serverHostModel{
		Name: types.StringValue("test-host"),
		Hosts: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("https://fleet-server:8220"),
		}),
		Ssl: ssl,
	}
}

func TestToAPICreateModel_Ssl(t *testing.T) {
	t.Parallel()

	t.Run("all fields", func(t *testing.T) {
		t.Parallel()

		ssl := sslObjectValue(t, new("cert"), new("key"), new("required"), []string{"ca1", "ca2"})
		body, diags := baseSSLModel(ssl).toAPICreateModel(t.Context())
		require.False(t, diags.HasError())
		require.NotNil(t, body.Ssl)
		assert.Equal(t, new("cert"), body.Ssl.Certificate)
		assert.Equal(t, new("key"), body.Ssl.Key)
		assert.Equal(t, &[]string{"ca1", "ca2"}, body.Ssl.CertificateAuthorities)
		require.NotNil(t, body.Ssl.ClientAuth)
		assert.Equal(t, kbapi.PostFleetFleetServerHostsJSONBodySslClientAuthRequired, *body.Ssl.ClientAuth)
	})

	t.Run("client_auth omitted is not sent", func(t *testing.T) {
		t.Parallel()

		ssl := sslObjectValue(t, new("cert"), new("key"), nil, nil)
		body, diags := baseSSLModel(ssl).toAPICreateModel(t.Context())
		require.False(t, diags.HasError())
		require.NotNil(t, body.Ssl)
		assert.Nil(t, body.Ssl.ClientAuth)

		raw, err := json.Marshal(body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"certificate":"cert","key":"key"}`, string(sslJSON(t, raw)))
	})

	t.Run("null ssl is not sent", func(t *testing.T) {
		t.Parallel()

		body, diags := baseSSLModel(types.ObjectNull(getSslAttrTypes(context.Background()))).toAPICreateModel(t.Context())
		require.False(t, diags.HasError())
		assert.Nil(t, body.Ssl)
	})
}

func sslJSON(t *testing.T, raw []byte) json.RawMessage {
	t.Helper()
	var envelope struct {
		Ssl json.RawMessage `json:"ssl"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))
	return envelope.Ssl
}

func TestToAPIUpdateModel_Ssl(t *testing.T) {
	t.Parallel()

	nullSSL := types.ObjectNull(getSslAttrTypes(context.Background()))
	populated := sslObjectValue(t, new("cert"), new("key"), new("optional"), []string{"ca1"})

	t.Run("planned ssl is sent", func(t *testing.T) {
		t.Parallel()

		body, diags := baseSSLModel(populated).toAPIUpdateModel(t.Context(), baseSSLModel(nullSSL))
		require.False(t, diags.HasError())
		require.NotNil(t, body.Ssl)
		assert.Equal(t, new("cert"), body.Ssl.Certificate)
		assert.Equal(t, new("key"), body.Ssl.Key)
		assert.Equal(t, &[]string{"ca1"}, body.Ssl.CertificateAuthorities)
		require.NotNil(t, body.Ssl.ClientAuth)
		assert.Equal(t, kbapi.PutFleetFleetServerHostsItemidJSONBodySslClientAuthOptional, *body.Ssl.ClientAuth)
	})

	t.Run("removed ssl is cleared with an empty object", func(t *testing.T) {
		t.Parallel()

		body, diags := baseSSLModel(nullSSL).toAPIUpdateModel(t.Context(), baseSSLModel(populated))
		require.False(t, diags.HasError())
		require.NotNil(t, body.Ssl)

		raw, err := json.Marshal(body)
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(sslJSON(t, raw)))
	})

	t.Run("dropping one child omits it from the replacement object", func(t *testing.T) {
		t.Parallel()

		withoutKey := sslObjectValue(t, new("cert"), nil, new("optional"), []string{"ca1"})
		body, diags := baseSSLModel(withoutKey).toAPIUpdateModel(t.Context(), baseSSLModel(populated))
		require.False(t, diags.HasError())
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"certificate":"cert","certificate_authorities":["ca1"],"client_auth":"optional"}`, string(sslJSON(t, raw)))
	})

	t.Run("ssl absent in plan and prior is omitted", func(t *testing.T) {
		t.Parallel()

		body, diags := baseSSLModel(nullSSL).toAPIUpdateModel(t.Context(), baseSSLModel(nullSSL))
		require.False(t, diags.HasError())
		assert.Nil(t, body.Ssl)
	})
}

func TestPopulateFromAPI_Ssl(t *testing.T) {
	t.Parallel()

	nullSSL := types.ObjectNull(getSslAttrTypes(context.Background()))

	tests := []struct {
		name string
		ssl  string
		want types.Object
	}{
		{
			name: "full round trip",
			ssl:  `{"certificate":"cert","key":"key","certificate_authorities":["ca1","ca2"],"client_auth":"required"}`,
			want: sslObjectValue(t, new("cert"), new("key"), new("required"), []string{"ca1", "ca2"}),
		},
		{
			name: "certificate authorities only",
			ssl:  `{"certificate_authorities":["ca1"]}`,
			want: sslObjectValue(t, nil, nil, nil, []string{"ca1"}),
		},
		{
			name: "certificate and key only",
			ssl:  `{"certificate":"cert","key":"key"}`,
			want: sslObjectValue(t, new("cert"), new("key"), nil, nil),
		},
		{
			name: "client_auth only",
			ssl:  `{"client_auth":"required"}`,
			want: sslObjectValue(t, nil, nil, new("required"), nil),
		},
		{
			name: "all nil normalizes to null",
			ssl:  `{}`,
			want: nullSSL,
		},
		{
			name: "absent normalizes to null",
			ssl:  ``,
			want: nullSSL,
		},
		{
			name: "empty CA list normalizes to null list",
			ssl:  `{"certificate":"cert","certificate_authorities":[]}`,
			want: sslObjectValue(t, new("cert"), nil, nil, nil),
		},
		{
			name: "deferred fields alone normalize to null",
			ssl:  `{"agent_certificate":"agent","es_key":"es"}`,
			want: nullSSL,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := serverHostFromSSLJSON(t, tc.ssl)
			var m serverHostModel
			diags := m.populateFromAPI(t.Context(), host)
			require.False(t, diags.HasError(), "%v", diags)
			assert.True(t, tc.want.Equal(m.Ssl), "want %s got %s", tc.want, m.Ssl)
		})
	}
}

func TestDeferredSSLFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ssl  string
		want []string
	}{
		{name: "nil ssl"},
		{name: "supported fields only", ssl: `{"certificate":"c","key":"k"}`},
		{name: "agent and es fields", ssl: `{"agent_key":"k","es_certificate_authorities":["ca"]}`, want: []string{"agent_key", "es_certificate_authorities"}},
		{name: "empty values are ignored", ssl: `{"agent_certificate":"","agent_certificate_authorities":[]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, deferredSSLFields(serverHostFromSSLJSON(t, tc.ssl)))
		})
	}
}

func serverHostFromSSLJSON(t *testing.T, ssl string) *kbapi.ServerHost {
	t.Helper()

	raw := `{"id":"id","name":"n","host_urls":["https://h:8220"]`
	if ssl != "" {
		raw += `,"ssl":` + ssl
	}
	raw += `}`

	var host kbapi.ServerHost
	require.NoError(t, json.Unmarshal([]byte(raw), &host))
	return &host
}
