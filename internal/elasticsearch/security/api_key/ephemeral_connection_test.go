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

package apikey

import (
	"context"
	"testing"

	clientconfig "github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchConnectionRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	connection, diags := types.ListValueFrom(ctx, providerschema.ElasticsearchConnectionObjectType(), []clientconfig.ElasticsearchConnection{
		{
			Username:    types.StringValue("elastic"),
			Password:    types.StringValue("secret"),
			APIKey:      types.StringValue("api-key-value"),
			BearerToken: types.StringValue("bearer-token"),
			Endpoints: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("https://es.example:9200"),
				types.StringValue("https://es-backup.example:9200"),
			}),
			Headers: types.MapValueMust(types.StringType, map[string]attr.Value{
				"X-Custom": types.StringValue("header-value"),
			}),
			Insecure: types.BoolValue(true),
			CAFile:   types.StringValue("/path/to/ca.pem"),
		},
	})
	require.False(t, diags.HasError())

	encoded, encodeDiags := encodeElasticsearchConnection(ctx, connection)
	require.False(t, encodeDiags.HasError())
	require.NotEmpty(t, encoded)

	decoded, decodeDiags := decodeElasticsearchConnection(ctx, encoded)
	require.False(t, decodeDiags.HasError())

	var decodedConnections []clientconfig.ElasticsearchConnection
	require.False(t, decoded.ElementsAs(ctx, &decodedConnections, false).HasError())
	require.Len(t, decodedConnections, 1)

	decodedConn := decodedConnections[0]
	require.Equal(t, "elastic", decodedConn.Username.ValueString())
	require.Equal(t, "secret", decodedConn.Password.ValueString())
	require.Equal(t, "api-key-value", decodedConn.APIKey.ValueString())
	require.Equal(t, "bearer-token", decodedConn.BearerToken.ValueString())
	require.True(t, decodedConn.Insecure.ValueBool())
	require.Equal(t, "/path/to/ca.pem", decodedConn.CAFile.ValueString())

	var endpoints []string
	require.False(t, decodedConn.Endpoints.ElementsAs(ctx, &endpoints, false).HasError())
	require.Equal(t, []string{"https://es.example:9200", "https://es-backup.example:9200"}, endpoints)

	var headers map[string]string
	require.False(t, decodedConn.Headers.ElementsAs(ctx, &headers, false).HasError())
	require.Equal(t, map[string]string{"X-Custom": "header-value"}, headers)
}
