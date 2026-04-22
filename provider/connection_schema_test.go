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
	elasticsearchConnectionTestKey = "elasticsearch_connection"
	kibanaConnectionTestKey        = "kibana_connection"
)

type registeredProviderEntity struct {
	id         string
	name       string
	entityKind string

	sdkEntity    *sdkschema.Resource
	frameworkRes fwresource.Resource
	frameworkDS  datasource.DataSource
}

func TestProviderEntities_ConnectionSchemas(t *testing.T) {
	ctx := context.Background()
	sdkProvider := provider.New("dev")
	frameworkProvider := provider.NewFrameworkProvider("dev")

	registered := collectRegisteredProviderEntities(ctx, sdkProvider, frameworkProvider)
	validated := make(map[string]struct{}, len(registered))

	expectedSDKElasticsearch := providerschema.GetEsConnectionSchema(elasticsearchConnectionTestKey, false)
	expectedSDKKibana := providerschema.GetKibanaEntityConnectionSchema()
	expectedFrameworkElasticsearch := providerschema.GetEsFWConnectionBlock()
	expectedFrameworkKibana := providerschema.GetKbFWConnectionBlock()

	for _, entity := range registered {
		t.Run(entity.id, func(t *testing.T) {
			validated[entity.id] = struct{}{}

			switch expectedConnectionForEntity(entity) {
			case elasticsearchConnectionTestKey:
				if entity.sdkEntity != nil {
					assertSDKConnectionSchemaMatches(t, entity, elasticsearchConnectionTestKey, expectedSDKElasticsearch)
				} else {
					assertFrameworkConnectionBlockMatches(ctx, t, entity, elasticsearchConnectionTestKey, expectedFrameworkElasticsearch)
				}
			case kibanaConnectionTestKey:
				if entity.sdkEntity != nil {
					assertSDKConnectionSchemaMatches(t, entity, kibanaConnectionTestKey, expectedSDKKibana)
				} else {
					assertFrameworkConnectionBlockMatches(ctx, t, entity, kibanaConnectionTestKey, expectedFrameworkKibana)
				}
			default:
				assertConnectionSchemaAbsent(ctx, t, entity, elasticsearchConnectionTestKey)
				assertConnectionSchemaAbsent(ctx, t, entity, kibanaConnectionTestKey)
			}
		})
	}

	t.Run("all_registered_entities_validated", func(t *testing.T) {
		missing := make([]string, 0)
		for _, entity := range registered {
			if _, ok := validated[entity.id]; !ok {
				missing = append(missing, entity.id)
			}
		}
		if len(missing) > 0 {
			t.Fatalf("registered entities missing validation: %s", strings.Join(missing, ", "))
		}
	})
}

func collectRegisteredProviderEntities(ctx context.Context, sdkProvider *sdkschema.Provider, frameworkProvider fwprovider.Provider) []registeredProviderEntity {
	entities := make([]registeredProviderEntity, 0, len(sdkProvider.ResourcesMap)+len(sdkProvider.DataSourcesMap))

	for _, name := range sortedSDKEntityNames(sdkProvider.ResourcesMap) {
		entities = append(entities, registeredProviderEntity{
			id:         fmt.Sprintf("sdk/resource/%s", name),
			name:       name,
			entityKind: "resource",
			sdkEntity:  sdkProvider.ResourcesMap[name],
		})
	}

	for _, name := range sortedSDKEntityNames(sdkProvider.DataSourcesMap) {
		entities = append(entities, registeredProviderEntity{
			id:         fmt.Sprintf("sdk/data_source/%s", name),
			name:       name,
			entityKind: "data_source",
			sdkEntity:  sdkProvider.DataSourcesMap[name],
		})
	}

	for _, entity := range collectFrameworkResourceEntities(ctx, frameworkProvider, func(string) bool { return true }) {
		entities = append(entities, registeredProviderEntity{
			id:           fmt.Sprintf("framework/resource/%s", entity.name),
			name:         entity.name,
			entityKind:   "resource",
			frameworkRes: entity.resource,
		})
	}

	for _, entity := range collectFrameworkDataSourceEntities(ctx, frameworkProvider, func(string) bool { return true }) {
		entities = append(entities, registeredProviderEntity{
			id:          fmt.Sprintf("framework/data_source/%s", entity.name),
			name:        entity.name,
			entityKind:  "data_source",
			frameworkDS: entity.dataSource,
		})
	}

	sort.Slice(entities, func(i, j int) bool { return entities[i].id < entities[j].id })
	return entities
}

func expectedConnectionForEntity(entity registeredProviderEntity) string {
	if entity.sdkEntity != nil && entity.entityKind == "data_source" && strings.HasPrefix(entity.name, "elasticstack_elasticsearch_ingest_processor_") {
		// These SDK data sources construct ingest processor payloads only and do not use provider clients.
		return ""
	}
	if strings.HasPrefix(entity.name, "elasticstack_elasticsearch_") {
		return elasticsearchConnectionTestKey
	}
	return kibanaConnectionTestKey
}

func assertSDKConnectionSchemaMatches(t *testing.T, entity registeredProviderEntity, blockKey string, expected *sdkschema.Schema) {
	t.Helper()
	if entity.sdkEntity == nil {
		t.Fatalf("entity %q is not an SDK entity", entity.id)
	}

	actual, ok := entity.sdkEntity.Schema[blockKey]
	if !ok {
		t.Fatalf("entity %q is missing %q schema", entity.id, blockKey)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("entity %q %q schema does not exactly match helper definition", entity.id, blockKey)
	}
	if actual.Deprecated != "" {
		t.Fatalf("entity %q %q schema has unexpected deprecation warning: %q", entity.id, blockKey, actual.Deprecated)
	}
}

func assertConnectionSchemaAbsent(ctx context.Context, t *testing.T, entity registeredProviderEntity, blockKey string) {
	t.Helper()

	switch {
	case entity.sdkEntity != nil:
		if actual, ok := entity.sdkEntity.Schema[blockKey]; ok {
			t.Fatalf("entity %q unexpectedly defines %q schema: %#v", entity.id, blockKey, actual)
		}
	case entity.frameworkRes != nil:
		resp := fwresource.SchemaResponse{}
		entity.frameworkRes.Schema(ctx, fwresource.SchemaRequest{}, &resp)
		if actual, ok := resp.Schema.Blocks[blockKey]; ok {
			t.Fatalf("entity %q unexpectedly defines %q block: %#v", entity.id, blockKey, actual)
		}
	case entity.frameworkDS != nil:
		resp := datasource.SchemaResponse{}
		entity.frameworkDS.Schema(ctx, datasource.SchemaRequest{}, &resp)
		if actual, ok := resp.Schema.Blocks[blockKey]; ok {
			t.Fatalf("entity %q unexpectedly defines %q block: %#v", entity.id, blockKey, actual)
		}
	default:
		t.Fatalf("entity %q has no supported implementation", entity.id)
	}
}

func assertFrameworkConnectionBlockMatches(ctx context.Context, t *testing.T, entity registeredProviderEntity, blockKey string, expected any) {
	t.Helper()

	switch {
	case entity.frameworkRes != nil:
		resp := fwresource.SchemaResponse{}
		entity.frameworkRes.Schema(ctx, fwresource.SchemaRequest{}, &resp)

		actual, ok := resp.Schema.Blocks[blockKey]
		if !ok {
			t.Fatalf("entity %q is missing %q block", entity.id, blockKey)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("entity %q %q block does not exactly match helper definition", entity.id, blockKey)
		}
		if msg := actual.GetDeprecationMessage(); msg != "" {
			t.Fatalf("entity %q %q block has unexpected deprecation message: %q", entity.id, blockKey, msg)
		}
	case entity.frameworkDS != nil:
		resp := datasource.SchemaResponse{}
		entity.frameworkDS.Schema(ctx, datasource.SchemaRequest{}, &resp)

		actual, ok := resp.Schema.Blocks[blockKey]
		if !ok {
			t.Fatalf("entity %q is missing %q block", entity.id, blockKey)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("entity %q %q block does not exactly match helper definition", entity.id, blockKey)
		}
		if msg := actual.GetDeprecationMessage(); msg != "" {
			t.Fatalf("entity %q %q block has unexpected deprecation message: %q", entity.id, blockKey, msg)
		}
	default:
		t.Fatalf("entity %q is not a framework resource or data source", entity.id)
	}
}
