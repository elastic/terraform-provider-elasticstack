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

package provider_test

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type frameworkResourceEntity struct {
	name     string
	resource fwresource.Resource
}

type frameworkDataSourceEntity struct {
	name       string
	dataSource datasource.DataSource
}

func frameworkResourceTypeName(ctx context.Context, r fwresource.Resource) string {
	resp := fwresource.MetadataResponse{}
	r.Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	return resp.TypeName
}

func frameworkDataSourceTypeName(ctx context.Context, d datasource.DataSource) string {
	resp := datasource.MetadataResponse{}
	d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &resp)
	return resp.TypeName
}

// collectFrameworkResourceEntities enumerates provider resources, returning those
// for which include(typeName) is true, sorted by name.
func collectFrameworkResourceEntities(ctx context.Context, p fwprovider.Provider, include func(name string) bool) []frameworkResourceEntity {
	entities := make([]frameworkResourceEntity, 0)
	for _, factory := range p.Resources(ctx) {
		r := factory()
		name := frameworkResourceTypeName(ctx, r)
		if include(name) {
			entities = append(entities, frameworkResourceEntity{name: name, resource: r})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
}

// collectFrameworkDataSourceEntities enumerates provider data sources, returning those
// for which include(typeName) is true, sorted by name.
func collectFrameworkDataSourceEntities(ctx context.Context, p fwprovider.Provider, include func(name string) bool) []frameworkDataSourceEntity {
	entities := make([]frameworkDataSourceEntity, 0)
	for _, factory := range p.DataSources(ctx) {
		d := factory()
		name := frameworkDataSourceTypeName(ctx, d)
		if include(name) {
			entities = append(entities, frameworkDataSourceEntity{name: name, dataSource: d})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
}

func sortedSDKEntityNames(entities map[string]*sdkschema.Resource) []string {
	keys := make([]string, 0, len(entities))
	for k := range entities {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
