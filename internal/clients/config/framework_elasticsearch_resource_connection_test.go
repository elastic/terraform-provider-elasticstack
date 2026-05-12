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

package config

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestNewFromFrameworkElasticsearchResourceConnection_preservesCredentialsWhenEnvSet(t *testing.T) {
	t.Setenv("ELASTICSEARCH_USERNAME", "env-user")
	t.Setenv("ELASTICSEARCH_PASSWORD", "env-pass")

	esConns := []ElasticsearchConnection{
		{
			Username: types.StringValue("scoped-user"),
			Password: types.StringValue("scoped-pass"),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://127.0.0.1:9200"),
			}),
			Insecure: types.BoolValue(true),
		},
	}

	client, diags := NewFromFrameworkElasticsearchResourceConnection(context.Background(), esConns, "unit-testing")
	require.False(t, diags.HasError())
	require.NotNil(t, client.Elasticsearch)
	require.Equal(t, "scoped-user", client.Elasticsearch.Username)
	require.Equal(t, "scoped-pass", client.Elasticsearch.Password)
}

func TestNewFromFrameworkElasticsearchResourceConnection_fillsMissingFromEnv(t *testing.T) {
	t.Setenv("ELASTICSEARCH_USERNAME", "from-env-user")
	t.Setenv("ELASTICSEARCH_PASSWORD", "from-env-pass")

	esConns := []ElasticsearchConnection{
		{
			Username: types.StringNull(),
			Password: types.StringNull(),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://127.0.0.1:9200"),
			}),
			Insecure: types.BoolValue(true),
		},
	}

	client, diags := NewFromFrameworkElasticsearchResourceConnection(context.Background(), esConns, "unit-testing")
	require.False(t, diags.HasError())
	require.NotNil(t, client.Elasticsearch)
	require.Equal(t, "from-env-user", client.Elasticsearch.Username)
	require.Equal(t, "from-env-pass", client.Elasticsearch.Password)
}

func TestNewFromFrameworkElasticsearchResourceConnection_basicAuthDoesNotPickAPIKeyOrBearerFromEnv(t *testing.T) {
	t.Setenv("ELASTICSEARCH_USERNAME", "env-user")
	t.Setenv("ELASTICSEARCH_PASSWORD", "env-pass")
	t.Setenv("ELASTICSEARCH_API_KEY", "env-api-key")
	t.Setenv("ELASTICSEARCH_BEARER_TOKEN", "env-bearer")

	esConns := []ElasticsearchConnection{
		{
			Username: types.StringValue("scoped-user"),
			Password: types.StringValue("scoped-pass"),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://127.0.0.1:9200"),
			}),
			Insecure: types.BoolValue(true),
		},
	}

	client, diags := NewFromFrameworkElasticsearchResourceConnection(context.Background(), esConns, "unit-testing")
	require.False(t, diags.HasError())
	require.NotNil(t, client.Elasticsearch)
	require.Equal(t, "scoped-user", client.Elasticsearch.Username)
	require.Equal(t, "scoped-pass", client.Elasticsearch.Password)
	require.Empty(t, client.Elasticsearch.APIKey)
	authz := client.Elasticsearch.Header.Get("Authorization")
	require.NotContains(t, authz, "env-bearer")
}

func TestNewFromFrameworkElasticsearchResourceConnection_apiKeyDoesNotPickUsernameFromEnv(t *testing.T) {
	t.Setenv("ELASTICSEARCH_USERNAME", "env-user")
	t.Setenv("ELASTICSEARCH_PASSWORD", "env-pass")

	esConns := []ElasticsearchConnection{
		{
			APIKey: types.StringValue("scoped-api-key"),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://127.0.0.1:9200"),
			}),
			Insecure: types.BoolValue(true),
		},
	}

	client, diags := NewFromFrameworkElasticsearchResourceConnection(context.Background(), esConns, "unit-testing")
	require.False(t, diags.HasError())
	require.NotNil(t, client.Elasticsearch)
	require.Equal(t, "scoped-api-key", client.Elasticsearch.APIKey)
	require.Empty(t, client.Elasticsearch.Username)
	require.Empty(t, client.Elasticsearch.Password)
}

func TestNewFromFrameworkElasticsearchResourceConnection_preservesBearerWhenBasicEnvSet(t *testing.T) {
	t.Setenv("ELASTICSEARCH_USERNAME", "env-user")
	t.Setenv("ELASTICSEARCH_PASSWORD", "env-pass")

	esConns := []ElasticsearchConnection{
		{
			BearerToken: types.StringValue("scoped-bearer"),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://127.0.0.1:9200"),
			}),
			Insecure: types.BoolValue(true),
		},
	}

	client, diags := NewFromFrameworkElasticsearchResourceConnection(context.Background(), esConns, "unit-testing")
	require.False(t, diags.HasError())
	require.NotNil(t, client.Elasticsearch)
	require.Equal(t, "Bearer scoped-bearer", client.Elasticsearch.Header.Get("Authorization"))
}
