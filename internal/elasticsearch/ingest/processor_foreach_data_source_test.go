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

func TestAccDataSourceIngestProcessorForeach(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "field", "values"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "json", expectedJSONForeach),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "field", "values"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "description", "foreach test"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "if", "ctx.values != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "tag", "foreach-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "on_failure.0", `{"set":{"field":"error.message","value":"foreach failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_foreach.test", "json", expectedJSONForeachAllAttributes),
				),
			},
		},
	})
}

const expectedJSONForeach = `{
  "foreach": {
		"field": "values",
		"ignore_failure": false,
		"ignore_missing": false,
		"processor": {
			"convert": {
				"field": "_ingest._value",
				"ignore_failure": false,
				"ignore_missing": false,
				"type": "integer"
			}
		}
	}
}
`

const expectedJSONForeachAllAttributes = `{
	"foreach": {
		"description": "foreach test",
		"field": "values",
		"if": "ctx.values != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "foreach failed"
				}
			}
		],
		"processor": {
			"convert": {
				"field": "_ingest._value",
				"ignore_failure": false,
				"ignore_missing": false,
				"type": "integer"
			}
		},
		"tag": "foreach-tag"
	}
}`
