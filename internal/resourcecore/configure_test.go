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

package resourcecore

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestCore_Configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_provider_data_stores_nil_client", func(t *testing.T) {
		t.Parallel()
		c := New(ComponentElasticsearch, "x")
		var resp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Nil(t, c.Client())
	})

	t.Run("valid_factory_stores_that_pointer", func(t *testing.T) {
		t.Parallel()
		c := New(ComponentElasticsearch, "x")
		f := clients.NewTestProviderClientFactoryForResourceUnitTests(t)
		var resp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Same(t, f, c.Client())
	})

	t.Run("invalid_provider_data_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		c := New(ComponentElasticsearch, "x")
		f := clients.NewTestProviderClientFactoryForResourceUnitTests(t)
		var okResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, c.Client())

		var badResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, &badResp)
		require.True(t, badResp.Diagnostics.HasError())
		require.Same(t, f, c.Client(), "client must stay the last successful assignment")
	})

	t.Run("nil_typed_factory_value_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		c := New(ComponentElasticsearch, "x")
		f := clients.NewTestProviderClientFactoryForResourceUnitTests(t)
		var okResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, c.Client())

		var nilFactory *clients.ProviderClientFactory
		var badResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: nilFactory}, &badResp)
		require.True(t, badResp.Diagnostics.HasError())
		require.Same(t, f, c.Client(), "client must stay the last successful assignment")
	})
}
