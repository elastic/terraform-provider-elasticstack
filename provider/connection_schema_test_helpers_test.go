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
	"fmt"
	"reflect"
	"sort"
	"testing"

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

// runFrameworkConnectionResourceSubtests runs per-entity subtests asserting that each
// resource exposes blockKey as a block matching expected with no deprecation message.
func runFrameworkConnectionResourceSubtests(ctx context.Context, t *testing.T, entities []frameworkResourceEntity, blockKey string, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/resource/%s", entity.name), func(t *testing.T) {
			resp := fwresource.SchemaResponse{}
			entity.resource.Schema(ctx, fwresource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[blockKey]
			if !ok {
				t.Fatalf("resource %q is missing %q block", entity.name, blockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("resource %q %q block does not exactly match helper definition", entity.name, blockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("resource %q %q block has unexpected deprecation message: %q", entity.name, blockKey, msg)
			}
		})
	}
}

// runFrameworkConnectionDataSourceSubtests runs per-entity subtests asserting that each
// data source exposes blockKey as a block matching expected with no deprecation message.
func runFrameworkConnectionDataSourceSubtests(ctx context.Context, t *testing.T, entities []frameworkDataSourceEntity, blockKey string, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/data_source/%s", entity.name), func(t *testing.T) {
			resp := datasource.SchemaResponse{}
			entity.dataSource.Schema(ctx, datasource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[blockKey]
			if !ok {
				t.Fatalf("data source %q is missing %q block", entity.name, blockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("data source %q %q block does not exactly match helper definition", entity.name, blockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("data source %q %q block has unexpected deprecation message: %q", entity.name, blockKey, msg)
			}
		})
	}
}

// runSDKConnectionEntitySubtests runs per-entity subtests asserting that each SDK
// resource/data source for which include(entityKind, name) is true exposes blockKey
// as a schema entry matching expected with no deprecation warning.
func runSDKConnectionEntitySubtests(t *testing.T, entityKind string, entities map[string]*sdkschema.Resource, blockKey string, expected *sdkschema.Schema, include func(kind, name string) bool) {
	t.Helper()

	names := make([]string, 0, len(entities))
	for name := range entities {
		if include(entityKind, name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		entityName := name
		entity := entities[entityName]

		t.Run(fmt.Sprintf("sdk/%s/%s", entityKind, entityName), func(t *testing.T) {
			if entity == nil {
				t.Fatalf("entity %q is nil", entityName)
			}

			actual, ok := entity.Schema[blockKey]
			if !ok {
				t.Fatalf("entity %q is missing %q schema", entityName, blockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("entity %q %q schema does not exactly match helper definition", entityName, blockKey)
			}

			if actual.Deprecated != "" {
				t.Fatalf("entity %q %q schema has unexpected deprecation warning: %q", entityName, blockKey, actual.Deprecated)
			}
		})
	}
}
