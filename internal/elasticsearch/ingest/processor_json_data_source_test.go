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

func TestAccDataSourceIngestProcessorJSON(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_json.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "string_source"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "target_field", "json_target"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJSONJSON),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_to_root"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_json.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "json_payload"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "add_to_root", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "add_to_root_conflict_strategy", "merge"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "target_field"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJSONJSONAddToRoot),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_json.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "document.json"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "allow_duplicate_keys", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "description", "Parse document JSON"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "if", "ctx.document?.json != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "on_failure.0", `{"set":{"field":"error.message","value":"json processor failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "tag", "json-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJSONJSONAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_json.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "field", "updated_string_source"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_json.test", "target_field", "updated_json_target"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_json.test", "json", expectedJSONJSONUpdatedValues),
				),
			},
		},
	})
}

const expectedJSONJSON = `{
	"json": {
		"field": "string_source",
		"ignore_failure": false,
		"target_field": "json_target"
	}
}`

const expectedJSONJSONAddToRoot = `{
	"json": {
		"add_to_root": true,
		"add_to_root_conflict_strategy": "merge",
		"field": "json_payload",
		"ignore_failure": false
	}
}`

const expectedJSONJSONAllAttributes = `{
	"json": {
		"allow_duplicate_keys": true,
		"description": "Parse document JSON",
		"field": "document.json",
		"if": "ctx.document?.json != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "json processor failed"
				}
			}
		],
		"tag": "json-tag"
	}
}`

const expectedJSONJSONUpdatedValues = `{
	"json": {
		"field": "updated_string_source",
		"ignore_failure": false,
		"target_field": "updated_json_target"
	}
}`
