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
	"strings"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	kbEntityPrefix       = "elasticstack_kibana_"
	fleetEntityPrefix    = "elasticstack_fleet_"
	kbConnectionBlockKey = "kibana_connection"
)

func TestSDKKibanaEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetKibanaEntityConnectionSchema()

	runSDKKibanaEntitySubtests(t, "resource", p.ResourcesMap, expected)
	runSDKKibanaEntitySubtests(t, "data_source", p.DataSourcesMap, expected)
}

func TestSDKFleetEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetKibanaEntityConnectionSchema()

	runSDKFleetEntitySubtests(t, "resource", p.ResourcesMap, expected)
	runSDKFleetEntitySubtests(t, "data_source", p.DataSourcesMap, expected)
}

func runSDKKibanaEntitySubtests(t *testing.T, entityKind string, entities map[string]*sdkschema.Resource, expected *sdkschema.Schema) {
	t.Helper()
	runSDKKibanaFleetEntitySubtests(t, entityKind, entities, expected, kbEntityPrefix)
}

func runSDKFleetEntitySubtests(t *testing.T, entityKind string, entities map[string]*sdkschema.Resource, expected *sdkschema.Schema) {
	t.Helper()
	runSDKKibanaFleetEntitySubtests(t, entityKind, entities, expected, fleetEntityPrefix)
}

func runSDKKibanaFleetEntitySubtests(t *testing.T, entityKind string, entities map[string]*sdkschema.Resource, expected *sdkschema.Schema, prefix string) {
	t.Helper()

	names := make([]string, 0, len(entities))
	for name := range entities {
		if isCoveredKibanaFleetEntity(name, prefix) {
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

			actual, ok := entity.Schema[kbConnectionBlockKey]
			if !ok {
				t.Fatalf("entity %q is missing %q schema", entityName, kbConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("entity %q %q schema does not exactly match helper definition", entityName, kbConnectionBlockKey)
			}

			if actual.Deprecated != "" {
				t.Fatalf("entity %q %q schema has unexpected deprecation warning: %q", entityName, kbConnectionBlockKey, actual.Deprecated)
			}
		})
	}
}

func TestFrameworkKibanaEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetKbFWConnectionBlock()

	resourceEntities := frameworkKibanaResourceEntities(ctx, baseProvider)
	dataSourceEntities := frameworkKibanaDataSourceEntities(ctx, baseProvider)

	runFrameworkKibanaResourceSubtests(ctx, t, resourceEntities, expected)
	runFrameworkKibanaDataSourceSubtests(ctx, t, dataSourceEntities, expected)
}

func TestFrameworkFleetEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetKbFWConnectionBlock()

	resourceEntities := frameworkFleetResourceEntities(ctx, baseProvider)
	dataSourceEntities := frameworkFleetDataSourceEntities(ctx, baseProvider)

	runFrameworkKibanaResourceSubtests(ctx, t, resourceEntities, expected)
	runFrameworkKibanaDataSourceSubtests(ctx, t, dataSourceEntities, expected)
}

func frameworkKibanaResourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkResourceEntity {
	return collectFrameworkResourceEntities(ctx, p, kbEntityPrefix)
}

func frameworkFleetResourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkResourceEntity {
	return collectFrameworkResourceEntities(ctx, p, fleetEntityPrefix)
}

func collectFrameworkResourceEntities(ctx context.Context, p fwprovider.Provider, prefix string) []frameworkResourceEntity {
	entities := make([]frameworkResourceEntity, 0)
	for _, factory := range p.Resources(ctx) {
		r := factory()
		name := frameworkResourceTypeName(ctx, r)
		if strings.HasPrefix(name, prefix) {
			entities = append(entities, frameworkResourceEntity{name: name, resource: r})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
}

func frameworkKibanaDataSourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkDataSourceEntity {
	return collectFrameworkDataSourceEntities(ctx, p, kbEntityPrefix)
}

func frameworkFleetDataSourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkDataSourceEntity {
	return collectFrameworkDataSourceEntities(ctx, p, fleetEntityPrefix)
}

func collectFrameworkDataSourceEntities(ctx context.Context, p fwprovider.Provider, prefix string) []frameworkDataSourceEntity {
	entities := make([]frameworkDataSourceEntity, 0)
	for _, factory := range p.DataSources(ctx) {
		d := factory()
		name := frameworkDataSourceTypeName(ctx, d)
		if strings.HasPrefix(name, prefix) {
			entities = append(entities, frameworkDataSourceEntity{name: name, dataSource: d})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
}

func runFrameworkKibanaResourceSubtests(ctx context.Context, t *testing.T, entities []frameworkResourceEntity, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/resource/%s", entity.name), func(t *testing.T) {
			resp := fwresource.SchemaResponse{}
			entity.resource.Schema(ctx, fwresource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[kbConnectionBlockKey]
			if !ok {
				t.Fatalf("resource %q is missing %q block", entity.name, kbConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("resource %q %q block does not exactly match helper definition", entity.name, kbConnectionBlockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("resource %q %q block has unexpected deprecation message: %q", entity.name, kbConnectionBlockKey, msg)
			}
		})
	}
}

func runFrameworkKibanaDataSourceSubtests(ctx context.Context, t *testing.T, entities []frameworkDataSourceEntity, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/data_source/%s", entity.name), func(t *testing.T) {
			resp := datasource.SchemaResponse{}
			entity.dataSource.Schema(ctx, datasource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[kbConnectionBlockKey]
			if !ok {
				t.Fatalf("data source %q is missing %q block", entity.name, kbConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("data source %q %q block does not exactly match helper definition", entity.name, kbConnectionBlockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("data source %q %q block has unexpected deprecation message: %q", entity.name, kbConnectionBlockKey, msg)
			}
		})
	}
}

func isCoveredKibanaFleetEntity(entityName, prefix string) bool {
	return strings.HasPrefix(entityName, prefix)
}
