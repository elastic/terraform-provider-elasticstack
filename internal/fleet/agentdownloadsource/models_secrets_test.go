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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func testSslObject(ctx context.Context, key, keyWO types.String) types.Object {
	return types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
		"certificate":             types.StringValue("cert"),
		"certificate_authorities": types.ListNull(types.StringType),
		"key":                     key,
		"key_wo":                  keyWO,
	})
}

func marshalBody(t *testing.T, body any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func idResolver(ids map[string]string) secretResolver {
	return func(attrPath, value string) secretValue {
		if id, ok := ids[attrPath]; ok {
			return secretValue{id: &id}
		}
		return secretValue{literal: &value}
	}
}

func TestSecretWriteBranch(t *testing.T) {
	t.Parallel()

	hash, err := secretHasher.Compute("the-secret")
	require.NoError(t, err)
	id := "existing-id"
	emptyID := ""

	tests := []struct {
		name       string
		value      string
		storedHash []byte
		existingID *string
		wantID     bool
	}{
		{name: "first apply sends literal", value: "the-secret"},
		{name: "first apply with GET id still sends literal", value: "the-secret", existingID: &id},
		{name: "unchanged hash and GET id references id", value: "the-secret", storedHash: hash, existingID: &id, wantID: true},
		{name: "unchanged hash without GET id sends literal", value: "the-secret", storedHash: hash},
		{name: "unchanged hash with empty GET id sends literal", value: "the-secret", storedHash: hash, existingID: &emptyID},
		{name: "changed value sends literal", value: "rotated", storedHash: hash, existingID: &id},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := secretWriteBranch(tt.value, tt.storedHash, tt.existingID)
			if tt.wantID {
				require.NotNil(t, got.id)
				require.Equal(t, id, *got.id)
				require.Nil(t, got.literal)
				return
			}
			require.Nil(t, got.id)
			require.NotNil(t, got.literal)
			require.Equal(t, tt.value, *got.literal)
		})
	}
}

func TestSecretValueMarshalJSON(t *testing.T) {
	t.Parallel()
	lit := "lit"
	id := "abc"
	raw, err := json.Marshal(secretValue{literal: &lit})
	require.NoError(t, err)
	require.JSONEq(t, `"lit"`, string(raw))
	raw, err = json.Marshal(secretValue{id: &id})
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"abc"}`, string(raw))
}

func TestCreateBodyRoutesPlainSslKeyThroughSecrets(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringValue("plain-key"), types.StringNull()),
	}
	body, diags := m.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	require.Nil(t, body.Ssl.Key, "plain ssl.key must not use the plain wire field")
	require.Equal(t, "cert", *body.Ssl.Certificate)

	got := marshalBody(t, body)
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "plain-key"}}, got["secrets"])
}

func TestCreateBodyRoutesKeyWOThroughSecretsAsLiteral(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringNull(), types.StringValue("wo-key")),
	}
	body, diags := m.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	require.Nil(t, body.Ssl.Key)
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "wo-key"}}, marshalBody(t, body)["secrets"])
}

func TestCreateBodyKeyWOOnlySslKeepsSslObject(t *testing.T) {
	ctx := context.Background()
	obj := types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
		"certificate":             types.StringNull(),
		"certificate_authorities": types.ListNull(types.StringType),
		"key":                     types.StringNull(),
		"key_wo":                  types.StringValue("wo-key"),
	})
	body, diags := model{Name: types.StringValue("n"), Host: types.StringValue("https://h"), Ssl: obj}.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	got := marshalBody(t, body)
	require.Equal(t, map[string]any{}, got["ssl"], "ssl object must still be sent so Fleet keeps the secret")
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "wo-key"}}, got["secrets"])
}

func TestCreateBodyAuthWriteOnlyFields(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Auth: &authModel{
			Username:   types.StringValue("svc"),
			Password:   types.StringNull(),
			PasswordWO: types.StringValue("pw"),
			APIKey:     types.StringNull(),
			APIKeyWO:   types.StringNull(),
			Headers:    types.ListNull(authHeadersElementType(ctx)),
		},
	}
	body, diags := m.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	require.Nil(t, body.Auth.Password)
	got := marshalBody(t, body)
	require.Equal(t, map[string]any{"username": "svc"}, got["auth"])
	require.Equal(t, map[string]any{"auth": map[string]any{"password": "pw"}}, got["secrets"])

	m.Auth = &authModel{
		Password:   types.StringNull(),
		PasswordWO: types.StringNull(),
		APIKey:     types.StringNull(),
		APIKeyWO:   types.StringValue("ak"),
		Username:   types.StringNull(),
		Headers:    types.ListNull(authHeadersElementType(ctx)),
	}
	body, diags = m.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	require.Equal(t, map[string]any{"auth": map[string]any{"api_key": "ak"}}, marshalBody(t, body)["secrets"])
}

func TestCreateBodyPlainAuthCredentialsStayOnPlainFields(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Auth: &authModel{
			Username:   types.StringValue("svc"),
			Password:   types.StringValue("pw"),
			PasswordWO: types.StringNull(),
			APIKey:     types.StringNull(),
			APIKeyWO:   types.StringNull(),
			Headers:    types.ListNull(authHeadersElementType(ctx)),
		},
	}
	body, diags := m.toAPICreateModel(ctx)
	require.False(t, diags.HasError())
	got := marshalBody(t, body)
	require.Equal(t, map[string]any{"username": "svc", "password": "pw"}, got["auth"])
	require.NotContains(t, got, "secrets")
}

func TestCreateBodyWithoutSecretsOmitsSecrets(t *testing.T) {
	body, diags := model{Name: types.StringValue("n"), Host: types.StringValue("https://h")}.toAPICreateModel(context.Background())
	require.False(t, diags.HasError())
	require.NotContains(t, marshalBody(t, body), "secrets")
}

func TestUpdateBodyReferencesExistingSecretsByID(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringNull(), types.StringValue("wo-key")),
		Auth: &authModel{
			Username:   types.StringValue("svc"),
			Password:   types.StringNull(),
			PasswordWO: types.StringValue("pw"),
			APIKey:     types.StringNull(),
			APIKeyWO:   types.StringNull(),
			Headers:    types.ListNull(authHeadersElementType(ctx)),
		},
	}
	body, diags := m.toAPIUpdateModel(ctx, model{}, idResolver(map[string]string{
		sslKeyWOPath: "ssl-id", authPasswordWOPath: "pw-id",
	}))
	require.False(t, diags.HasError())
	require.Equal(t, map[string]any{
		"ssl":  map[string]any{"key": map[string]any{"id": "ssl-id"}},
		"auth": map[string]any{"password": map[string]any{"id": "pw-id"}},
	}, marshalBody(t, body)["secrets"])
}

func TestUpdateBodyPlainSslKeyAlwaysLiteral(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringValue("plain-key"), types.StringNull()),
	}
	body, diags := m.toAPIUpdateModel(ctx, model{}, idResolver(map[string]string{sslKeyWOPath: "ssl-id"}))
	require.False(t, diags.HasError())
	require.Equal(t, map[string]any{"ssl": map[string]any{"key": "plain-key"}}, marshalBody(t, body)["secrets"])
}

func TestUpdateBodyRemovedSecretSendsEmptySecrets(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringNull(), types.StringNull()),
	}
	prior := model{Ssl: testSslObject(ctx, types.StringNull(), types.StringValue("old"))}
	body, diags := m.toAPIUpdateModel(ctx, prior, literalSecrets)
	require.False(t, diags.HasError())
	got := marshalBody(t, body)
	// An explicit empty secrets object makes Fleet drop its reference to the
	// removed secret; without it older Fleet versions delete the secret but keep
	// the dangling reference, which then fails the resource's delete with a 404.
	require.Equal(t, map[string]any{}, got["secrets"])
	require.Contains(t, got, "ssl", "ssl is still sent so Fleet clears the removed secret")
}

func TestUpdateBodyOmitsSecretsWhenNeverConfigured(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("n"),
		Host: types.StringValue("https://h"),
		Ssl:  testSslObject(ctx, types.StringNull(), types.StringNull()),
	}
	body, diags := m.toAPIUpdateModel(ctx, model{Ssl: types.ObjectNull(getSslAttrTypes(ctx))}, literalSecrets)
	require.False(t, diags.HasError())
	require.NotContains(t, marshalBody(t, body), "secrets")
}

func TestWithWriteOnlyFromConfig(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Ssl: testSslObject(ctx, types.StringNull(), types.StringNull()),
		Auth: &authModel{
			Username: types.StringValue("svc"), Password: types.StringNull(),
			PasswordWO: types.StringNull(), APIKey: types.StringNull(), APIKeyWO: types.StringNull(),
			Headers: types.ListNull(authHeadersElementType(ctx)),
		},
	}
	config := model{
		Ssl: testSslObject(ctx, types.StringNull(), types.StringValue("cfg-key")),
		Auth: &authModel{
			Username: types.StringValue("svc"), Password: types.StringNull(),
			PasswordWO: types.StringValue("cfg-pw"), APIKey: types.StringNull(), APIKeyWO: types.StringValue("cfg-ak"),
			Headers: types.ListNull(authHeadersElementType(ctx)),
		},
	}
	merged, diags := plan.withWriteOnlyFrom(ctx, config)
	require.False(t, diags.HasError())

	api, sslDiags := objectValueToSSL(ctx, merged.Ssl)
	require.False(t, sslDiags.HasError())
	require.Equal(t, "cfg-key", *api.KeyWO)
	require.Equal(t, "cfg-pw", merged.Auth.PasswordWO.ValueString())
	require.Equal(t, "cfg-ak", merged.Auth.APIKeyWO.ValueString())
	// the original plan is untouched
	require.True(t, plan.Auth.PasswordWO.IsNull())

	// null ssl/auth in config leave the plan untouched
	merged, diags = plan.withWriteOnlyFrom(ctx, model{Ssl: types.ObjectNull(getSslAttrTypes(ctx))})
	require.False(t, diags.HasError())
	require.Equal(t, plan.Ssl, merged.Ssl)
	require.True(t, merged.Auth.PasswordWO.IsNull())
}
