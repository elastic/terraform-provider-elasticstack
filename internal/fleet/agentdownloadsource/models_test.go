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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type versionGateClient struct {
	server *version.Version
}

func (c versionGateClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	return !c.server.LessThan(minVersion), nil
}

func TestGetVersionRequirementsRejectsAuthBelowMinimumVersion(t *testing.T) {
	m := model{
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
	}

	diags := entitycore.EnforceVersionRequirements(context.Background(), versionGateClient{
		server: version.Must(version.NewVersion("9.3.0")),
	}, m)

	require.True(t, diags.HasError())
	require.Equal(t, "Unsupported server version", diags[0].Summary())
	require.Equal(t, "`auth` is not supported on this version; it is only supported in Elastic Stack 9.4.0 and above", diags[0].Detail())
	withPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	require.Equal(t, path.Root("auth"), withPath.Path())

	reqs, reqDiags := m.GetVersionRequirements(context.Background())
	require.False(t, reqDiags.HasError())
	require.Len(t, reqs, 2)
	require.Nil(t, reqs[0].AttributePath)
	require.Equal(t, "8.13.0", reqs[0].MinVersion.String())
	require.Equal(t, "This resource requires stack version 8.13.0 or higher.", reqs[0].ErrorMessage)
	require.NotNil(t, reqs[1].AttributePath)
	require.Equal(t, "auth", reqs[1].AttributePath.String())
	require.Equal(t, "9.4.0", reqs[1].MinVersion.String())
}

func TestGetVersionRequirementsAllowsOmittedAuthBelowAuthMinimumVersion(t *testing.T) {
	m := model{}

	diags := entitycore.EnforceVersionRequirements(context.Background(), versionGateClient{
		server: version.Must(version.NewVersion("8.13.0")),
	}, m)

	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	for _, d := range diags {
		require.NotContains(t, d.Detail(), "auth")
		require.NotContains(t, d.Summary(), "auth")
	}

	reqs, reqDiags := m.GetVersionRequirements(context.Background())
	require.False(t, reqDiags.HasError())
	require.Len(t, reqs, 1)
	require.Nil(t, reqs[0].AttributePath)
	require.Equal(t, "8.13.0", reqs[0].MinVersion.String())
	require.Equal(t, "This resource requires stack version 8.13.0 or higher.", reqs[0].ErrorMessage)
}

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
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			Headers: mustAuthHeaders(t, authHeaderModel{Key: types.StringValue("X-A"), Value: types.StringValue("1")}),
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

func TestModelToAPICreateModelSendsSSL(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		}),
	}

	body, bodyDiags := state.toAPICreateModel(ctx)
	require.False(t, bodyDiags.HasError())
	require.NotNil(t, body.Ssl)
	require.Equal(t, "cert", *body.Ssl.Certificate)
	require.Equal(t, []string{"ca"}, *body.Ssl.CertificateAuthorities)
	require.Equal(t, "key", *body.Ssl.Key)
}

func TestModelToAPICreateModelOmitsNullCertificateAuthorities(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListNull(types.StringType),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		}),
	}

	body, bodyDiags := state.toAPICreateModel(ctx)
	require.False(t, bodyDiags.HasError())
	require.NotNil(t, body.Ssl)
	require.Nil(t, body.Ssl.CertificateAuthorities)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "certificate_authorities")
	require.Contains(t, string(encoded), `"ssl":{"certificate":"cert","key":"key"}`)
}

func TestModelToAPICreateModelOmitsNullSSL(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
	}

	body, bodyDiags := state.toAPICreateModel(ctx)
	require.False(t, bodyDiags.HasError())
	require.Nil(t, body.Ssl)
}

func TestModelToAPIUpdateModelSendsSSL(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		}),
	}

	body, bodyDiags := plan.toAPIUpdateModel(ctx, model{})
	require.False(t, bodyDiags.HasError())
	require.NotNil(t, body.Ssl)
	require.Equal(t, "cert", *body.Ssl.Certificate)
	require.Equal(t, []string{"ca"}, *body.Ssl.CertificateAuthorities)
	require.Equal(t, "key", *body.Ssl.Key)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"ssl":{"certificate":"cert","certificate_authorities":["ca"],"key":"key"}`)
}

func TestModelToAPIUpdateModelOmitsDroppedCertificateAuthorities(t *testing.T) {
	ctx := context.Background()
	sslObject := func(cas types.List) types.Object {
		return types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": cas,
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		})
	}
	plan := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl:     sslObject(types.ListNull(types.StringType)),
	}
	prior := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl:     sslObject(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")})),
	}

	body, bodyDiags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, bodyDiags.HasError())
	require.NotNil(t, body.Ssl)
	require.Nil(t, body.Ssl.CertificateAuthorities)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "certificate_authorities")
	require.Contains(t, string(encoded), `"ssl":{"certificate":"cert","key":"key"}`)
}

func TestModelToAPIUpdateModelRemovesSSLWithEmptyObject(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
	}
	prior := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		}),
	}

	body, bodyDiags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, bodyDiags.HasError())
	require.NotNil(t, body.Ssl)
	require.Nil(t, body.Ssl.Certificate)
	require.Nil(t, body.Ssl.CertificateAuthorities)
	require.Nil(t, body.Ssl.Key)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"ssl":{}`)
}

func TestModelToAPIUpdateModelOmitsSSLOnUnchangedAbsence(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
	}

	body, bodyDiags := plan.toAPIUpdateModel(ctx, model{})
	require.False(t, bodyDiags.HasError())
	require.Nil(t, body.Ssl)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "ssl")
}

func TestModelGetVersionRequirementsWithSSL(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("download-source"),
		Host: types.StringValue("https://artifacts.example.com/elastic-agent"),
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringNull(),
			"key_wo":                  types.StringNull(),
		}),
	}

	reqs, diags := m.GetVersionRequirements(ctx)
	require.False(t, diags.HasError())
	require.Len(t, reqs, 2)

	resourceRequirement := reqs[0]
	require.Nil(t, resourceRequirement.AttributePath)
	require.Equal(t, "8.13.0", resourceRequirement.MinVersion.String())

	sslRequirement := reqs[1]
	require.NotNil(t, sslRequirement.AttributePath)
	require.Equal(t, "ssl", sslRequirement.AttributePath.String())
	require.Equal(t, "9.1.0", sslRequirement.MinVersion.String())
}

func TestModelGetVersionRequirementsWithoutSSL(t *testing.T) {
	ctx := context.Background()
	m := model{
		Name: types.StringValue("download-source"),
		Host: types.StringValue("https://artifacts.example.com/elastic-agent"),
	}

	reqs, diags := m.GetVersionRequirements(ctx)
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Nil(t, reqs[0].AttributePath)
}

func TestModelToAPIUpdateModelSendsEmptyAuthWhenRemoved(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(false),
	}
	prior := model{
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
	}

	body, diags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.Nil(t, body.Auth.Username)
	require.Nil(t, body.Auth.Password)
	require.Nil(t, body.Auth.ApiKey)
	require.Nil(t, body.Auth.Headers)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"auth":{}`)
}

func TestModelToAPIUpdateModelOmitsAuthWhenNeverConfigured(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("renamed"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(false),
	}

	body, diags := plan.toAPIUpdateModel(ctx, model{})
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.Nil(t, body.Auth)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), `"auth"`)
}

func TestModelToAPIUpdateModelReplacesAuthWithAPIKeyOnly(t *testing.T) {
	ctx := context.Background()
	plan := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(false),
		Auth: &authModel{
			APIKey: types.StringValue("rotated-key"),
		},
	}
	prior := model{
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
	}

	body, diags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.Nil(t, body.Auth.Username)
	require.Nil(t, body.Auth.Password)
	require.NotNil(t, body.Auth.ApiKey)
	require.Equal(t, "rotated-key", *body.Auth.ApiKey)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"api_key":"rotated-key"`)
	require.NotContains(t, string(encoded), `"username"`)
	require.NotContains(t, string(encoded), `"password"`)
}

func mustAuthHeaders(t *testing.T, headers ...authHeaderModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), authHeadersElementType(context.Background()), headers)
	require.Empty(t, diags)
	return list
}

func TestModelToAPICreateModelSendsAuthAndSSLTogether(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("download-source"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default: types.BoolValue(true),
		Auth: &authModel{
			Username: types.StringValue("svc-account"),
			Password: types.StringValue("secret"),
		},
		Ssl: types.ObjectValueMust(getSslAttrTypes(ctx), map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca")}),
			"certificate":             types.StringValue("cert"),
			"key":                     types.StringValue("key"),
			"key_wo":                  types.StringNull(),
		}),
	}

	body, diags := state.toAPICreateModel(ctx)
	require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)
	require.NotNil(t, body.Auth)
	require.Equal(t, "svc-account", *body.Auth.Username)
	require.Equal(t, "secret", *body.Auth.Password)
	require.NotNil(t, body.Ssl)
	require.Equal(t, "cert", *body.Ssl.Certificate)
	require.Equal(t, []string{"ca"}, *body.Ssl.CertificateAuthorities)
	require.Equal(t, "key", *body.Ssl.Key)

	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"ssl":{"certificate":"cert","certificate_authorities":["ca"],"key":"key"}`)
	require.Contains(t, string(encoded), `"username":"svc-account"`)
}
