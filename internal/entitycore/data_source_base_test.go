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

package entitycore

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

func TestDataSourceBase_Configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_provider_data_stores_nil_client", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentElasticsearch, "x")
		var resp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Nil(t, b.Client())
	})

	t.Run("valid_factory_stores_that_pointer", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var resp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Same(t, f, b.Client())
	})

	t.Run("invalid_provider_data_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var okResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, b.Client())

		var badResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: "wrong-type"}, &badResp)
		require.True(t, badResp.Diagnostics.HasError())
		require.Same(t, f, b.Client(), "client must stay the last successful assignment")
	})

	t.Run("typed_nil_factory_pointer_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var okResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, b.Client())

		var nilFactory *clients.ProviderClientFactory
		var badResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: nilFactory}, &badResp)
		require.True(t, badResp.Diagnostics.HasError())
		require.Same(t, f, b.Client(), "client must stay the last successful assignment")
	})

	t.Run("untyped_nil_clears_prior_factory", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var okResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, b.Client())

		var nilResp datasource.ConfigureResponse
		b.Configure(ctx, datasource.ConfigureRequest{ProviderData: nil}, &nilResp)
		require.False(t, nilResp.Diagnostics.HasError(), "untyped nil should succeed and clear client")
		require.Nil(t, b.Client(), "client should be nil after untyped nil reconfiguration")
	})
}

func TestDataSourceBase_Metadata_typeNamesPerComponent(t *testing.T) {
	cases := []struct {
		name           string
		component      Component
		dataSourceName string
		want           string
	}{
		{
			name:           "elasticsearch",
			component:      ComponentElasticsearch,
			dataSourceName: "enrich_policy",
			want:           "elasticstack_elasticsearch_enrich_policy",
		},
		{
			name:           "kibana",
			component:      ComponentKibana,
			dataSourceName: "spaces",
			want:           "elasticstack_kibana_spaces",
		},
		{
			name:           "fleet",
			component:      ComponentFleet,
			dataSourceName: "enrollment_tokens",
			want:           "elasticstack_fleet_enrollment_tokens",
		},
		{
			name:           "apm",
			component:      ComponentAPM,
			dataSourceName: "agent_configuration",
			want:           "elasticstack_apm_agent_configuration",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := NewDataSourceBase(tc.component, tc.dataSourceName)
			var resp datasource.MetadataResponse
			b.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)
			require.Equal(t, tc.want, resp.TypeName)
		})
	}
}

func TestDataSourceBase_Client_nilSafe(t *testing.T) {
	t.Run("nil_receiver", func(t *testing.T) {
		t.Parallel()
		var b *DataSourceBase
		require.Nil(t, b.Client())
	})

	t.Run("non_nil_before_configure", func(t *testing.T) {
		t.Parallel()
		b := NewDataSourceBase(ComponentFleet, "enrollment_tokens")
		require.Nil(t, b.Client())
	})
}
