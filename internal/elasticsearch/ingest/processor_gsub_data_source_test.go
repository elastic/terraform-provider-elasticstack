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

func TestAccDataSourceIngestProcessorGsub(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "field", "field1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "pattern", "\\."),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "replacement", "-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "json", expectedJSONGsub),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "field", "field1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "pattern", "\\."),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "replacement", "-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "target_field", "normalized_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "description", "Normalize a dotted field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "if", "ctx.message != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.0", `{"append":{"field":"errors","value":["gsub failed"]}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "tag", "gsub-normalize"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "json", expectedJSONGsubAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "field", "field2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "pattern", ":"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "replacement", "_"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "target_field", "normalized_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "description", "Normalize colon-delimited field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "if", "ctx.message != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.0", `{"append":{"field":"errors","value":["gsub failed"]}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "tag", "gsub-normalize"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "json", expectedJSONGsubUpdated),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "field", "field1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "pattern", "\\."),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "replacement", "-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_gsub.test", "json", expectedJSONGsub),
				),
			},
		},
	})
}

const expectedJSONGsub = `{
	"gsub": {
		"field": "field1",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern": "\\.",
		"replacement": "-"
	}
}`

const expectedJSONGsubAllAttributes = `{
	"gsub": {
		"description": "Normalize a dotted field",
		"field": "field1",
		"if": "ctx.message != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"append": {
					"field": "errors",
					"value": [
						"gsub failed"
					]
				}
			}
		],
		"pattern": "\\.",
		"replacement": "-",
		"tag": "gsub-normalize",
		"target_field": "normalized_field"
	}
}`

const expectedJSONGsubUpdated = `{
	"gsub": {
		"description": "Normalize colon-delimited field",
		"field": "field2",
		"if": "ctx.message != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"append": {
					"field": "errors",
					"value": [
						"gsub failed"
					]
				}
			}
		],
		"pattern": ":",
		"replacement": "_",
		"tag": "gsub-normalize",
		"target_field": "normalized_field"
	}
}`
