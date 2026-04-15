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
	"strings"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/provider"
)

const (
	esEntityPrefix       = "elasticstack_elasticsearch_"
	esIngestDSPrefix     = "elasticstack_elasticsearch_ingest_processor"
	esConnectionBlockKey = "elasticsearch_connection"
)

func TestSDKElasticsearchEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	p := provider.New("dev")
	expected := providerschema.GetEsConnectionSchema(esConnectionBlockKey, false)

	runSDKConnectionEntitySubtests(t, "resource", p.ResourcesMap, esConnectionBlockKey, expected, isCoveredElasticsearchEntity)
	runSDKConnectionEntitySubtests(t, "data_source", p.DataSourcesMap, esConnectionBlockKey, expected, isCoveredElasticsearchEntity)
}

func TestFrameworkElasticsearchEntities_ConnectionSchemaMatchesHelper(t *testing.T) {
	ctx := context.Background()
	baseProvider := provider.NewFrameworkProvider("dev")
	expected := providerschema.GetEsFWConnectionBlock()

	resourceEntities := collectFrameworkResourceEntities(ctx, baseProvider, func(name string) bool {
		return strings.HasPrefix(name, esEntityPrefix)
	})
	dataSourceEntities := collectFrameworkDataSourceEntities(ctx, baseProvider, func(name string) bool {
		return isCoveredElasticsearchEntity("data_source", name)
	})

	runFrameworkConnectionResourceSubtests(ctx, t, resourceEntities, esConnectionBlockKey, expected)
	runFrameworkConnectionDataSourceSubtests(ctx, t, dataSourceEntities, esConnectionBlockKey, expected)
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
