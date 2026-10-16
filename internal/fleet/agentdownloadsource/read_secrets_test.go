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

package agentdownloadsource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nullSslObject() types.Object {
	return types.ObjectNull(getSslAttrTypes(context.Background()))
}

// sslPriorObject builds a prior ssl object as it appears in plan or state.
func sslPriorObject(certificate, key types.String) types.Object {
	return types.ObjectValueMust(getSslAttrTypes(context.Background()), map[string]attr.Value{
		"certificate":             certificate,
		"certificate_authorities": types.ListNull(types.StringType),
		"key":                     key,
		"key_wo":                  types.StringNull(),
	})
}

func itemServer(t *testing.T, item map[string]any) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"item": item}))
	})
}

func hydrateWithPrior(t *testing.T, item map[string]any, fallbackAuth *authModel, fallbackSsl types.Object) model {
	t.Helper()
	item["id"] = "src"
	item["name"] = "n"
	item["host"] = "https://h"
	item["is_default"] = false
	client := newTestFleetClient(t, itemServer(t, item))
	state, found, diags := readAndHydrateState(context.Background(), client, "src", "default",
		types.SetNull(types.StringType), providerschema.KibanaConnectionNullList(), fallbackAuth, fallbackSsl)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.True(t, found)
	return state
}

func decodeSsl(t *testing.T, obj types.Object) sslModel {
	t.Helper()
	var diags diag.Diagnostics
	m := typeutils.ObjectTypeAs[sslModel](context.Background(), obj, fwpath.Root("ssl"), &diags)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	return *m
}

func TestFetchSecretIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secrets any
		want    secretIDs
	}{
		{name: "no secrets object"},
		{
			name: "all three ids",
			secrets: map[string]any{
				"ssl":  map[string]any{"key": map[string]any{"id": "ssl-id"}},
				"auth": map[string]any{"password": map[string]any{"id": "pw-id"}, "api_key": map[string]any{"id": "ak-id"}},
			},
			want: secretIDs{sslKey: new("ssl-id"), authPassword: new("pw-id"), authAPIKey: new("ak-id")},
		},
		{
			name:    "string arm yields no id",
			secrets: map[string]any{"ssl": map[string]any{"key": "literal-echo"}},
		},
		{
			name:    "empty id yields no id",
			secrets: map[string]any{"ssl": map[string]any{"key": map[string]any{"id": ""}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			item := map[string]any{"id": "src", "name": "n", "host": "https://h"}
			if tt.secrets != nil {
				item["secrets"] = tt.secrets
			}
			client := newTestFleetClient(t, itemServer(t, item))
			got, diags := fetchSecretIDs(context.Background(), client, "src", "default")
			require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFetchSecretIDsNotFoundReturnsNoIDs(t *testing.T) {
	t.Parallel()
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	got, diags := fetchSecretIDs(context.Background(), client, "src", "default")
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Equal(t, secretIDs{}, got)
}

func TestHydrationNeverCopiesEchoedKeyWOIntoPlainKey(t *testing.T) {
	t.Parallel()

	// Secret storage is unavailable, so Fleet merges the key_wo literal into
	// the plain ssl.key field of the response.
	state := hydrateWithPrior(t, map[string]any{
		"ssl": map[string]any{"certificate": "cert", "key": "wo-literal"},
	}, nil, sslPriorObject(types.StringValue("cert"), types.StringNull()))

	ssl := decodeSsl(t, state.Ssl)
	require.Equal(t, "cert", ssl.Certificate.ValueString())
	require.True(t, ssl.Key.IsNull(), "echoed write-only literal must not populate ssl.key")
	require.True(t, ssl.KeyWO.IsNull())
}

func TestHydrationPreservesRedactedPlainKey(t *testing.T) {
	t.Parallel()

	state := hydrateWithPrior(t, map[string]any{
		"ssl":     map[string]any{"certificate": "cert"},
		"secrets": map[string]any{"ssl": map[string]any{"key": map[string]any{"id": "ssl-id"}}},
	}, nil, sslPriorObject(types.StringValue("cert"), types.StringValue("configured-key")))

	require.Equal(t, "configured-key", decodeSsl(t, state.Ssl).Key.ValueString())
}

func TestHydrationPlainKeyTakesAPIValueWhenReturned(t *testing.T) {
	t.Parallel()

	state := hydrateWithPrior(t, map[string]any{
		"ssl": map[string]any{"certificate": "cert", "key": "server-key"},
	}, nil, sslPriorObject(types.StringValue("cert"), types.StringValue("configured-key")))

	require.Equal(t, "server-key", decodeSsl(t, state.Ssl).Key.ValueString())
}

func TestHydrationWithoutPriorSslHydratesKeyFromAPI(t *testing.T) {
	t.Parallel()

	// Import: no prior ssl object, so the API value is the only source.
	state := hydrateWithPrior(t, map[string]any{
		"ssl": map[string]any{"certificate": "cert", "key": "server-key"},
	}, nil, nullSslObject())

	ssl := decodeSsl(t, state.Ssl)
	require.Equal(t, "server-key", ssl.Key.ValueString())
	require.True(t, ssl.KeyWO.IsNull())
}

func TestHydrationRetainsSslObjectForKeyWOOnly(t *testing.T) {
	t.Parallel()

	prior := sslPriorObject(types.StringNull(), types.StringNull())
	for name, ssl := range map[string]any{
		"ssl absent from response":              nil,
		"ssl empty in response":                 map[string]any{},
		"ssl echoes the key_wo literal as key":  map[string]any{"key": "wo-literal"},
		"ssl present with empty string members": map[string]any{"certificate": "", "key": ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			item := map[string]any{}
			if ssl != nil {
				item["ssl"] = ssl
			}
			state := hydrateWithPrior(t, item, nil, prior)

			require.False(t, state.Ssl.IsNull(), "ssl object must be retained")
			m := decodeSsl(t, state.Ssl)
			require.True(t, m.Certificate.IsNull())
			require.True(t, m.CertificateAuthorities.IsNull())
			require.True(t, m.Key.IsNull())
			require.True(t, m.KeyWO.IsNull())
		})
	}
}

func TestHydrationNullsEmptySslWhenPriorHadChildren(t *testing.T) {
	t.Parallel()

	state := hydrateWithPrior(t, map[string]any{},
		nil, sslPriorObject(types.StringValue("cert"), types.StringNull()))
	require.True(t, state.Ssl.IsNull(), "an ssl object removed on the server must still read as null")
}

func TestHydrationNeverCopiesEchoedWriteOnlyAuthIntoPlainFields(t *testing.T) {
	t.Parallel()

	fallback := &authModel{
		Username:   types.StringValue("svc"),
		Password:   types.StringNull(),
		PasswordWO: types.StringNull(),
		APIKey:     types.StringNull(),
		APIKeyWO:   types.StringNull(),
		Headers:    types.ListNull(authHeadersElementType(context.Background())),
	}
	state := hydrateWithPrior(t, map[string]any{
		"auth": map[string]any{"username": "svc", "password": "wo-literal", "api_key": "wo-ak-literal"},
	}, fallback, nullSslObject())

	require.NotNil(t, state.Auth)
	require.True(t, state.Auth.Password.IsNull(), "echoed password_wo literal must not populate password")
	require.True(t, state.Auth.APIKey.IsNull(), "echoed api_key_wo literal must not populate api_key")
	require.True(t, state.Auth.PasswordWO.IsNull())
	require.True(t, state.Auth.APIKeyWO.IsNull())
}

func TestHydrationAuthWithoutPriorHydratesCredentialsFromAPI(t *testing.T) {
	t.Parallel()

	state := hydrateWithPrior(t, map[string]any{
		"auth": map[string]any{"username": "svc", "password": "server-pw"},
	}, nil, nullSslObject())

	require.NotNil(t, state.Auth)
	require.Equal(t, "server-pw", state.Auth.Password.ValueString())
	require.True(t, state.Auth.PasswordWO.IsNull())
	require.True(t, state.Auth.APIKeyWO.IsNull())
}

func TestHydrationRetainsAuthObjectForAPIKeyWOOnly(t *testing.T) {
	t.Parallel()

	fallback := &authModel{
		Username:   types.StringNull(),
		Password:   types.StringNull(),
		PasswordWO: types.StringNull(),
		APIKey:     types.StringNull(),
		APIKeyWO:   types.StringNull(),
		Headers:    types.ListNull(authHeadersElementType(context.Background())),
	}
	state := hydrateWithPrior(t, map[string]any{}, fallback, nullSslObject())
	require.NotNil(t, state.Auth)
	require.True(t, state.Auth.APIKey.IsNull())
}
