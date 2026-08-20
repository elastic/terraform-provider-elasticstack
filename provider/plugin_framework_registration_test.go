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

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

const (
	managedIntegrationResourceType = "elasticstack_fleet_managed_integration"
	removedAgentlessPolicyType     = "elasticstack_fleet_agentless_policy"
	kibanaTagResourceType          = "elasticstack_kibana_tag"
)

func registeredResourceTypeNames(ctx context.Context, p *Provider) map[string]struct{} {
	names := make(map[string]struct{})
	for _, newRes := range p.Resources(ctx) {
		r := newRes()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: ProviderTypeName}, &resp)
		names[resp.TypeName] = struct{}{}
	}
	return names
}

func stableResourceTypeNames(ctx context.Context, p *Provider) map[string]struct{} {
	names := make(map[string]struct{})
	for _, newRes := range p.resources(ctx) {
		r := newRes()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: ProviderTypeName}, &resp)
		names[resp.TypeName] = struct{}{}
	}
	return names
}

func resourceTypeNames(ctx context.Context, factories []func() resource.Resource) []string {
	names := make([]string, 0, len(factories))
	for _, newRes := range factories {
		r := newRes()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: ProviderTypeName}, &resp)
		names = append(names, resp.TypeName)
	}
	return names
}

func TestProvider_stableResourcesIncludeManagedIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := &Provider{version: "0.16.2"}

	stable := stableResourceTypeNames(ctx, p)
	require.Contains(t, stable, managedIntegrationResourceType)
	require.NotContains(t, stable, removedAgentlessPolicyType)

	experimental := resourceTypeNames(ctx, p.experimentalResources(ctx))
	require.NotContains(t, experimental, managedIntegrationResourceType)
	require.NotContains(t, experimental, removedAgentlessPolicyType)
}

func TestProvider_kibanaTagNotRegisteredInDefaultProvider(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, "")

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: "0.16.2"})

	require.NotContains(t, names, kibanaTagResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_kibanaTagRegisteredWhenExperimentalEnvEnabled(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, envVarEnabled)

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: "0.16.2"})

	require.Contains(t, names, kibanaTagResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_kibanaTagRegisteredWithAccTestVersion(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, "")

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: AccTestVersion})

	require.Contains(t, names, kibanaTagResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_managedIntegrationRegisteredWithoutOptIn(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, "")

	ctx := context.Background()

	released := registeredResourceTypeNames(ctx, &Provider{version: "0.16.2"})
	require.Contains(t, released, managedIntegrationResourceType)
	require.NotContains(t, released, removedAgentlessPolicyType)

	accTest := registeredResourceTypeNames(ctx, &Provider{version: AccTestVersion})
	require.Contains(t, accTest, managedIntegrationResourceType)
	require.NotContains(t, accTest, removedAgentlessPolicyType)
}

func TestProvider_managedIntegrationRegisteredExactlyOnceWhenExperimentalEnabled(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, envVarEnabled)

	ctx := context.Background()
	p := &Provider{version: "0.16.2"}

	count := 0
	for _, typeName := range resourceTypeNames(ctx, p.Resources(ctx)) {
		if typeName == managedIntegrationResourceType {
			count++
		}
	}
	require.Equal(t, 1, count, "%s must be registered exactly once", managedIntegrationResourceType)
}
