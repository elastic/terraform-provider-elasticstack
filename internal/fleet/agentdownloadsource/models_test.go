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

func TestModelToAPICreateModel(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:     types.StringValue("download-source"),
		Host:     types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default:  types.BoolValue(true),
		ProxyID:  types.StringValue("proxy-123"),
		SourceID: types.StringValue("source-123"),
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Equal(t, "download-source", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-123", *body.ProxyId)
	require.NotNil(t, body.Id)
	require.Equal(t, "source-123", *body.Id)
}

func TestModelToAPICreateModelIncludesBasicAuth(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.NotNil(t, body.Auth.Username)
	require.Equal(t, "svc-account", *body.Auth.Username)
	require.NotNil(t, body.Auth.Password)
	require.Equal(t, "secret", *body.Auth.Password)
	require.Nil(t, body.Auth.ApiKey)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"username":"svc-account"`)
	require.Contains(t, string(encoded), `"password":"secret"`)
	require.NotContains(t, string(encoded), `"api_key"`)
}

func TestModelToAPICreateModelIncludesAPIKeyAuth(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			APIKey: types.StringValue("secret-key"),
		},
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.NotNil(t, body.Auth.ApiKey)
	require.Equal(t, "secret-key", *body.Auth.ApiKey)
	require.Nil(t, body.Auth.Username)
	require.Nil(t, body.Auth.Password)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"api_key":"secret-key"`)
	require.NotContains(t, string(encoded), `"username"`)
	require.NotContains(t, string(encoded), `"password"`)
}

func TestModelToAPICreateModelSendsHeadersOnly(t *testing.T) {
	ctx := context.Background()
	headerType := authHeadersElementType(ctx).(types.ObjectType)
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Headers: types.ListValueMust(headerType, []attr.Value{
				types.ObjectValueMust(headerType.AttrTypes, map[string]attr.Value{
					"key":   types.StringValue("X-A"),
					"value": types.StringValue("1"),
				}),
			}),
		},
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.Nil(t, body.Auth.Username)
	require.Nil(t, body.Auth.Password)
	require.Nil(t, body.Auth.ApiKey)
	require.NotNil(t, body.Auth.Headers)
	require.Equal(t, []downloadSourceAuthHeader{{Key: "X-A", Value: "1"}}, *body.Auth.Headers)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"headers":[{"key":"X-A","value":"1"}]`)
	require.NotContains(t, string(encoded), `"username"`)
	require.NotContains(t, string(encoded), `"password"`)
	require.NotContains(t, string(encoded), `"api_key"`)
}

func TestModelToAPICreateModelOmitsNullHeaders(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
			Headers:  types.ListNull(authHeadersElementType(ctx)),
		},
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.Nil(t, body.Auth.Headers)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), `"headers"`)
	require.Contains(t, string(encoded), `"username":"svc-account"`)
	require.Contains(t, string(encoded), `"password":"secret"`)
}

func TestModelToAPICreateModelSendsEmptyHeadersList(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Headers: types.ListValueMust(authHeadersElementType(ctx), []attr.Value{}),
		},
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.NotNil(t, body.Auth.Headers)
	require.Empty(t, *body.Auth.Headers)
	require.Nil(t, body.Auth.Username)
	require.Nil(t, body.Auth.Password)
	require.Nil(t, body.Auth.ApiKey)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"headers":[]`)
	require.NotContains(t, string(encoded), `"username"`)
	require.NotContains(t, string(encoded), `"password"`)
	require.NotContains(t, string(encoded), `"api_key"`)
}

func TestModelToAPICreateModelOmitsUnsetAuth(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Nil(t, body.Auth)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), `"auth"`)
}

func TestModelToAPICreateModelWithoutSourceID(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:     types.StringValue("download-source"),
		Host:     types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default:  types.BoolValue(false),
		ProxyID:  types.StringNull(),
		SourceID: types.StringNull(),
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Equal(t, "download-source", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent", body.Host)
	require.NotNil(t, body.IsDefault)
	require.False(t, *body.IsDefault)
	require.Nil(t, body.ProxyId)
	require.Nil(t, body.Id)
}

func TestModelGetSpaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spaceIDs types.Set
		want     string
	}{
		{
			name:     "null defaults to default",
			spaceIDs: types.SetNull(types.StringType),
			want:     "default",
		},
		{
			name:     "unknown defaults to default",
			spaceIDs: types.SetUnknown(types.StringType),
			want:     "default",
		},
		{
			name: "first non-empty space id",
			spaceIDs: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("my-space"),
			}),
			want: "my-space",
		},
		{
			name: "skips empty and unknown elements",
			spaceIDs: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(""),
				types.StringUnknown(),
				types.StringValue("my-space"),
			}),
			want: "my-space",
		},
		{
			name: "all empty defaults to default",
			spaceIDs: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(""),
			}),
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := model{SpaceIDs: tt.spaceIDs}
			require.Equal(t, tt.want, m.GetSpaceID().ValueString())
		})
	}
}

func TestModelToAPIUpdateModel(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(true),
		ProxyID: types.StringValue("proxy-456"),
	}

	body, diags := state.toAPIUpdateModel(ctx, model{})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Equal(t, "updated-name", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent-updated", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-456", *body.ProxyId)
}

func TestModelToAPIUpdateModelIncludesBasicAuth(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
	}

	body, diags := plan.toAPIUpdateModel(ctx, model{})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.NotNil(t, body.Auth.Username)
	require.Equal(t, "svc-account", *body.Auth.Username)
	require.NotNil(t, body.Auth.Password)
	require.Equal(t, "secret", *body.Auth.Password)
	require.Nil(t, body.Auth.ApiKey)
}

func TestModelToAPIUpdateModelClearsUnsetProxyID(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(false),
		ProxyID: types.StringNull(),
	}
	prior := model{ProxyID: types.StringValue("proxy-456")}

	body, diags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.ProxyId, "clearing proxy_id must send empty string so omitempty does not drop the field")
	require.Empty(t, *body.ProxyId)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"proxy_id":""`)
}

func TestModelToAPIUpdateModelOmitsAlreadyUnsetProxyID(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(false),
		ProxyID: types.StringNull(),
	}

	body, diags := plan.toAPIUpdateModel(ctx, model{ProxyID: types.StringNull()})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Nil(t, body.ProxyId)
}
