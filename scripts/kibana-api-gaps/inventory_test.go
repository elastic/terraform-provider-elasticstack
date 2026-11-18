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

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverResourcesListsRegisteredResourcesWithOwningPackage(t *testing.T) {
	resources := discoverResources()
	byName := make(map[string]Resource, len(resources))
	for _, r := range resources {
		byName[r.Name] = r
		assert.NotEmpty(t, r.PkgPath, r.Name)
		assert.False(t, strings.HasSuffix(r.PkgPath, "/internal/entitycore"), "%s resolved to the generic envelope package", r.Name)
	}

	for _, name := range []string{"elasticstack_kibana_slo", "elasticstack_fleet_agent_policy", "elasticstack_apm_agent_configuration"} {
		_, ok := byName[name]
		assert.True(t, ok, name)
	}
	assert.True(t, strings.HasSuffix(byName["elasticstack_kibana_slo"].PkgPath, "/internal/kibana/slo"))
}

func TestDiscoverResourcesExcludesDataSources(t *testing.T) {
	registered := make(map[string]bool)
	for _, r := range discoverResources() {
		registered[r.Name] = true
	}

	ctx := context.Background()
	checked := 0
	for _, factory := range provider.NewFrameworkProvider("test").DataSources(ctx) {
		var meta datasource.MetadataResponse
		factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
		if !strings.HasPrefix(meta.TypeName, "elasticstack_kibana_") || isRegisteredResource(t, meta.TypeName) {
			continue
		}
		checked++
		assert.False(t, registered[meta.TypeName], meta.TypeName)
	}
	assert.Positive(t, checked)
}

func isRegisteredResource(t *testing.T, name string) bool {
	t.Helper()
	ctx := context.Background()
	for _, factory := range provider.NewFrameworkProvider("test").Resources(ctx) {
		var meta resource.MetadataResponse
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
		if meta.TypeName == name {
			return true
		}
	}
	return false
}

func TestDiscoverResourcesIsSortedByName(t *testing.T) {
	resources := discoverResources()
	for i := 1; i < len(resources); i++ {
		require.Less(t, resources[i-1].Name, resources[i].Name)
	}
}

func TestFactoryPackageStripsFunctionName(t *testing.T) {
	assert.Equal(t, "github.com/x/y/internal/kibana/slo", factoryPackage("github.com/x/y/internal/kibana/slo.NewResource"))
	assert.Equal(t, "github.com/x/y/pkg", factoryPackage("github.com/x/y/pkg.(*T).Method-fm"))
	assert.Equal(t, "main", factoryPackage("main.run.func1"))
}
