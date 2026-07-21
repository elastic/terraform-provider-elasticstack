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
)

func registeredResourceTypeNames(ctx context.Context, p *Provider) map[string]struct{} {
	names := make(map[string]struct{})
	for _, newRes := range p.Resources(ctx) {
		r := newRes()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
		names[resp.TypeName] = struct{}{}
	}
	return names
}

func TestProvider_managedIntegrationRegisteredWithExperimentalAccTestVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: AccTestVersion})

	require.Contains(t, names, managedIntegrationResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_managedIntegrationNotRegisteredInDefaultProvider(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, "")

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: "0.16.2"})

	require.NotContains(t, names, managedIntegrationResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_managedIntegrationRegisteredWhenExperimentalEnvEnabled(t *testing.T) {
	t.Setenv(IncludeExperimentalEnvVar, envVarEnabled)

	ctx := context.Background()
	names := registeredResourceTypeNames(ctx, &Provider{version: "0.16.2"})

	require.Contains(t, names, managedIntegrationResourceType)
	require.NotContains(t, names, removedAgentlessPolicyType)
}

func TestProvider_experimentalResourcesIncludesManagedIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var found bool
	for _, newRes := range (&Provider{}).experimentalResources(ctx) {
		r := newRes()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
		if resp.TypeName == managedIntegrationResourceType {
			found = true
			break
		}
	}
	require.True(t, found, "experimentalResources() must register %s", managedIntegrationResourceType)
}
