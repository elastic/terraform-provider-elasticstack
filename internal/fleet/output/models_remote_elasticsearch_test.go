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

package output

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputModelToAPICreateRemoteElasticsearchModel(t *testing.T) {
	t.Parallel()

	model := outputModel{
		OutputID:                    types.StringValue("remote-output-id"),
		Name:                        types.StringValue("remote-output"),
		Type:                        types.StringValue("remote_elasticsearch"),
		Hosts:                       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://remote-es:9200")}),
		ServiceToken:                types.StringValue("service-token-value"),
		DefaultIntegrations:         types.BoolValue(true),
		DefaultMonitoring:           types.BoolValue(false),
		SyncIntegrations:            types.BoolValue(true),
		SyncUninstalledIntegrations: types.BoolValue(false),
		WriteToLogsStreams:          types.BoolValue(false),
	}

	union, diags := model.toAPICreateRemoteElasticsearchModel(context.Background())
	require.False(t, diags.HasError())

	body, err := union.AsNewOutputRemoteElasticsearch()
	require.NoError(t, err)

	assert.Equal(t, kbapi.KibanaHTTPAPIsNewOutputRemoteElasticsearchTypeRemoteElasticsearch, body.Type)
	assert.Equal(t, "remote-output", body.Name)
	assert.Equal(t, []string{"https://remote-es:9200"}, body.Hosts)
	require.NotNil(t, body.ServiceToken)
	assert.Equal(t, "service-token-value", *body.ServiceToken)
	require.NotNil(t, body.SyncIntegrations)
	assert.True(t, *body.SyncIntegrations)
	require.NotNil(t, body.SyncUninstalledIntegrations)
	assert.False(t, *body.SyncUninstalledIntegrations)
	require.NotNil(t, body.WriteToLogsStreams)
	assert.False(t, *body.WriteToLogsStreams)
}

func TestOutputModelRemoteElasticsearchModelMapsSSLCertificateAuthoritiesAndClientKeypair(t *testing.T) {
	t.Parallel()

	sslObj := types.ObjectValueMust(
		getSslAttrTypes(),
		map[string]attr.Value{
			"certificate_authorities": types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("ca-1"),
				types.StringValue("ca-2"),
			}),
			"certificate": types.StringValue("client-cert"),
			"key":         types.StringValue("client-key"),
		},
	)

	t.Run("create", func(t *testing.T) {
		t.Parallel()

		model := outputModel{
			OutputID:     types.StringValue("remote-output-id"),
			Name:         types.StringValue("remote-output"),
			Type:         types.StringValue("remote_elasticsearch"),
			Hosts:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://remote-es:9200")}),
			ServiceToken: types.StringValue("service-token-value"),
			Ssl:          sslObj,
		}

		union, diags := model.toAPICreateRemoteElasticsearchModel(context.Background())
		require.False(t, diags.HasError())

		body, err := union.AsNewOutputRemoteElasticsearch()
		require.NoError(t, err)

		require.NotNil(t, body.Ssl)
		require.NotNil(t, body.Ssl.CertificateAuthorities)
		assert.Equal(t, []string{"ca-1", "ca-2"}, *body.Ssl.CertificateAuthorities)
		require.NotNil(t, body.Ssl.Certificate)
		assert.Equal(t, "client-cert", *body.Ssl.Certificate)
		require.NotNil(t, body.Ssl.Key)
		assert.Equal(t, "client-key", *body.Ssl.Key)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		model := outputModel{
			Name:         types.StringValue("updated-remote-output"),
			Type:         types.StringValue("remote_elasticsearch"),
			Hosts:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://remote-es-2:9200")}),
			ServiceToken: types.StringValue("updated-service-token"),
			Ssl:          sslObj,
		}

		union, diags := model.toAPIUpdateRemoteElasticsearchModel(context.Background())
		require.False(t, diags.HasError())

		body, err := union.AsUpdateOutputRemoteElasticsearch()
		require.NoError(t, err)

		require.NotNil(t, body.Ssl)
		require.NotNil(t, body.Ssl.CertificateAuthorities)
		assert.Equal(t, []string{"ca-1", "ca-2"}, *body.Ssl.CertificateAuthorities)
		require.NotNil(t, body.Ssl.Certificate)
		assert.Equal(t, "client-cert", *body.Ssl.Certificate)
		require.NotNil(t, body.Ssl.Key)
		assert.Equal(t, "client-key", *body.Ssl.Key)
	})
}

func TestOutputModelToAPIUpdateRemoteElasticsearchModel(t *testing.T) {
	t.Parallel()

	model := outputModel{
		Name:         types.StringValue("updated-remote-output"),
		Type:         types.StringValue("remote_elasticsearch"),
		Hosts:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://remote-es-2:9200")}),
		ServiceToken: types.StringValue("updated-service-token"),
	}

	union, diags := model.toAPIUpdateRemoteElasticsearchModel(context.Background())
	require.False(t, diags.HasError())

	body, err := union.AsUpdateOutputRemoteElasticsearch()
	require.NoError(t, err)

	require.NotNil(t, body.Type)
	assert.Equal(t, kbapi.RemoteElasticsearch, *body.Type)
	require.NotNil(t, body.Name)
	assert.Equal(t, "updated-remote-output", *body.Name)
	require.NotNil(t, body.Hosts)
	assert.Equal(t, []string{"https://remote-es-2:9200"}, *body.Hosts)
	require.NotNil(t, body.ServiceToken)
	assert.Equal(t, "updated-service-token", *body.ServiceToken)
}

func TestOutputModelFromAPIRemoteElasticsearchModelPreservesServiceToken(t *testing.T) {
	t.Parallel()

	model := outputModel{
		ServiceToken: types.StringValue("existing-token"),
		SpaceIDs:     types.SetNull(types.StringType),
	}

	diags := model.fromAPIRemoteElasticsearchModel(context.Background(), &kbapi.OutputRemoteElasticsearch{
		Id:                 new("output-id"),
		Name:               "remote-output",
		Type:               kbapi.KibanaHTTPAPIsOutputRemoteElasticsearchTypeRemoteElasticsearch,
		Hosts:              []string{"https://remote-es:9200"},
		SyncIntegrations:   new(true),
		WriteToLogsStreams: new(false),
		// ServiceToken intentionally omitted to simulate redaction.
	})
	require.False(t, diags.HasError())

	assert.Equal(t, "existing-token", model.ServiceToken.ValueString())
	assert.Equal(t, "output-id", model.OutputID.ValueString())
	assert.Equal(t, "remote_elasticsearch", model.Type.ValueString())
	assert.True(t, model.SyncIntegrations.ValueBool())
	assert.False(t, model.WriteToLogsStreams.ValueBool())
}
