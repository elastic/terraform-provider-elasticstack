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

	body, bodyDiags := state.toAPICreateModel(ctx)
	require.False(t, bodyDiags.HasError())
	require.Equal(t, "download-source", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-123", *body.ProxyId)
	require.NotNil(t, body.Id)
	require.Equal(t, "source-123", *body.Id)
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

	body, bodyDiags := state.toAPICreateModel(ctx)
	require.False(t, bodyDiags.HasError())
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

	body, bodyDiags := state.toAPIUpdateModel(ctx, model{})
	require.False(t, bodyDiags.HasError())
	require.Equal(t, "updated-name", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent-updated", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-456", *body.ProxyId)
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

	body, bodyDiags := plan.toAPIUpdateModel(ctx, prior)
	require.False(t, bodyDiags.HasError())
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

	body, bodyDiags := plan.toAPIUpdateModel(ctx, model{ProxyID: types.StringNull()})
	require.False(t, bodyDiags.HasError())
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
