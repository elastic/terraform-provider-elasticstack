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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

// nonNilTestFactory returns a distinct, zero-valued *ProviderClientFactory. This is
// enough for [ResourceBase.Configure] / [ResourceBase.Client] semantics tests; resolution methods
// on the factory are not invoked.
func nonNilTestFactory() *clients.ProviderClientFactory {
	return new(clients.ProviderClientFactory)
}

func TestResourceBase_Configure(t *testing.T) {
	ctx := context.Background()

	// ProviderData is an untyped nil interface: conversion succeeds and assigns nil.
	t.Run("nil_provider_data_stores_nil_client", func(t *testing.T) {
		t.Parallel()
		c := NewResourceBase(ComponentElasticsearch, "x")
		var resp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Nil(t, c.Client())
	})

	t.Run("valid_factory_stores_that_pointer", func(t *testing.T) {
		t.Parallel()
		c := NewResourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var resp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Same(t, f, c.Client())
	})

	// After a non-nil factory is stored, ProviderData that converts with no error
	// replaces it: untyped nil succeeds and clears the stored factory (delta spec).
	t.Run("success_then_untyped_nil_provider_data_replaces_with_nil_client", func(t *testing.T) {
		t.Parallel()
		c := NewResourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var first resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &first)
		require.False(t, first.Diagnostics.HasError())
		require.Same(t, f, c.Client())

		var second resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: nil}, &second)
		require.False(t, second.Diagnostics.HasError())
		require.Nil(t, c.Client())
	})

	t.Run("invalid_provider_data_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		c := NewResourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
		var okResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: f}, &okResp)
		require.False(t, okResp.Diagnostics.HasError())
		require.Same(t, f, c.Client())

		var badResp resource.ConfigureResponse
		c.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong-type"}, &badResp)
		require.True(t, badResp.Diagnostics.HasError())
		require.Same(t, f, c.Client(), "client must stay the last successful assignment")
	})

	// Typed *ProviderClientFactory nil is not the same as an untyped nil interface:
	// ConvertProviderDataToFactory treats the former as "set but invalid" and errors,
	// distinct from the untyped nil success path in nil_provider_data_stores_nil_client
	// and success_then_untyped_nil_provider_data_replaces_with_nil_client.
	t.Run("typed_nil_factory_pointer_leaves_prior_client", func(t *testing.T) {
		t.Parallel()
		c := NewResourceBase(ComponentElasticsearch, "x")
		f := nonNilTestFactory()
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
