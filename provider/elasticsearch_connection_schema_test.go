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
	esEntityPrefix       = "elasticstack_elasticsearch_"
	esIngestDSPrefix     = "elasticstack_elasticsearch_ingest_processor"
	esConnectionBlockKey = "elasticsearch_connection"
)

func TestSDKElasticsearchEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetEsConnectionSchema(esConnectionBlockKey, false)

	runSDKEntitySubtests(t, "resource", p.ResourcesMap, expected)
	runSDKEntitySubtests(t, "data_source", p.DataSourcesMap, expected)
}

func runSDKEntitySubtests(t *testing.T, entityKind string, entities map[string]*sdkschema.Resource, expected *sdkschema.Schema) {
	t.Helper()

	names := make([]string, 0, len(entities))
	for name := range entities {
		if isCoveredElasticsearchEntity(entityKind, name) {
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

			actual, ok := entity.Schema[esConnectionBlockKey]
			if !ok {
				t.Fatalf("entity %q is missing %q schema", entityName, esConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("entity %q %q schema does not exactly match helper definition", entityName, esConnectionBlockKey)
			}

			if actual.Deprecated != "" {
				t.Fatalf("entity %q %q schema has unexpected deprecation warning: %q", entityName, esConnectionBlockKey, actual.Deprecated)
			}
		})
	}
}

func TestFrameworkElasticsearchEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetEsFWConnectionBlock()

	resourceEntities := frameworkResourceEntities(ctx, baseProvider)
	dataSourceEntities := frameworkDataSourceEntities(ctx, baseProvider)

	runFrameworkResourceSubtests(t, ctx, resourceEntities, expected)
	runFrameworkDataSourceSubtests(t, ctx, dataSourceEntities, expected)
}

type frameworkResourceEntity struct {
	name     string
	resource fwresource.Resource
}

type frameworkDataSourceEntity struct {
	name       string
	dataSource datasource.DataSource
}

func frameworkResourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkResourceEntity {
	entities := make([]frameworkResourceEntity, 0)
	for _, factory := range p.Resources(ctx) {
		r := factory()
		name := frameworkResourceTypeName(ctx, r)
		if strings.HasPrefix(name, esEntityPrefix) {
			entities = append(entities, frameworkResourceEntity{name: name, resource: r})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
}

func frameworkDataSourceEntities(ctx context.Context, p fwprovider.Provider) []frameworkDataSourceEntity {
	entities := make([]frameworkDataSourceEntity, 0)
	for _, factory := range p.DataSources(ctx) {
		d := factory()
		name := frameworkDataSourceTypeName(ctx, d)
		if isCoveredElasticsearchEntity("data_source", name) {
			entities = append(entities, frameworkDataSourceEntity{name: name, dataSource: d})
		}
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].name < entities[j].name })
	return entities
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

func runFrameworkResourceSubtests(t *testing.T, ctx context.Context, entities []frameworkResourceEntity, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/resource/%s", entity.name), func(t *testing.T) {
			resp := fwresource.SchemaResponse{}
			entity.resource.Schema(ctx, fwresource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[esConnectionBlockKey]
			if !ok {
				t.Fatalf("resource %q is missing %q block", entity.name, esConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("resource %q %q block does not exactly match helper definition", entity.name, esConnectionBlockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("resource %q %q block has unexpected deprecation message: %q", entity.name, esConnectionBlockKey, msg)
			}
		})
	}
}

func runFrameworkDataSourceSubtests(t *testing.T, ctx context.Context, entities []frameworkDataSourceEntity, expected any) {
	t.Helper()

	for _, e := range entities {
		entity := e
		t.Run(fmt.Sprintf("framework/data_source/%s", entity.name), func(t *testing.T) {
			resp := datasource.SchemaResponse{}
			entity.dataSource.Schema(ctx, datasource.SchemaRequest{}, &resp)

			actual, ok := resp.Schema.Blocks[esConnectionBlockKey]
			if !ok {
				t.Fatalf("data source %q is missing %q block", entity.name, esConnectionBlockKey)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("data source %q %q block does not exactly match helper definition", entity.name, esConnectionBlockKey)
			}

			if msg := actual.GetDeprecationMessage(); msg != "" {
				t.Fatalf("data source %q %q block has unexpected deprecation message: %q", entity.name, esConnectionBlockKey, msg)
			}
		})
	}
}

func isCoveredElasticsearchEntity(entityKind, entityName string) bool {
	if !strings.HasPrefix(entityName, esEntityPrefix) {
		return false
	}
	// Ingest processor data sources build processor payloads and do not use Elasticsearch clients.
	if entityKind == "data_source" && strings.HasPrefix(entityName, esIngestDSPrefix) {
		return false
	}
	return true
}
