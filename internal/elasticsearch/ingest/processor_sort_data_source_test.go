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

package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorSort(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_sort.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "field", "array_field_to_sort"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "order", "asc"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_sort.test", "json", expectedJSONSortDefaults),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_sort.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "field", "items"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "order", "desc"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "target_field", "sorted_items"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "description", "sort array"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "if", "ctx.items != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_sort.test", "on_failure.0", `{"append":{"field":"errors","value":"sort_failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "tag", "sort-items"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_sort.test", "json", expectedJSONSortAllAttributes),
				),
			},
		},
	})
}

const expectedJSONSortDefaults = `{
	"sort": {
		"field": "array_field_to_sort",
		"ignore_failure": false,
		"order": "asc"
	}
}`

const expectedJSONSortAllAttributes = `{
	"sort": {
		"description": "sort array",
		"if": "ctx.items != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"append": {
					"field": "errors",
					"value": "sort_failed"
				}
			}
		],
		"tag": "sort-items",
		"field": "items",
		"order": "desc",
		"target_field": "sorted_items"
	}
}`
