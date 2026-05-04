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

func TestAccDataSourceIngestProcessorKV(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_kv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field_split", " "),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "value_split", "="),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.*", "tags"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "prefix", "setting_"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "strip_brackets", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_kv.test", "json", expectedJSONKV),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_kv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field", "log.original"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field_split", "&"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "value_split", ":"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "target_field", "labels"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.*", "env"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.*", "region"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.*", "debug"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "prefix", "kv_"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_key", "_"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_value", "|"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "strip_brackets", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "description", "Parse selected labels"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "if", "ctx.log?.original != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_kv.test", "on_failure.0", `{"set":{"field":"error.message","value":"kv failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "tag", "kv-all-attributes"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_kv.test", "json", expectedJSONKVAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_kv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field", "event.original"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field_split", "&"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "value_split", "="),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.#"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_missing", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "prefix"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_key"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_value"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "strip_brackets", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "if"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_kv.test", "json", expectedJSONKVDefaults),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_kv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field", "labels"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "field_split", ";"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "value_split", "=>"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "target_field", "parsed_labels"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.*", "service"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "include_keys.*", "zone"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.*", "debug"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "exclude_keys.*", "temp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "prefix", "meta_"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_key", "-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "trim_value", "~"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "strip_brackets", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_kv.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_kv.test", "json", expectedJSONKVUpdatedValues),
				),
			},
		},
	})
}

const expectedJSONKV = `{
  "kv": {
		"exclude_keys": [
			"tags"
		],
		"field": "message",
		"field_split": " ",
		"ignore_failure": false,
		"ignore_missing": false,
		"prefix": "setting_",
		"strip_brackets": false,
		"value_split": "="
	}
}
`

const expectedJSONKVAllAttributes = `{
	"kv": {
		"description": "Parse selected labels",
		"if": "ctx.log?.original != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "kv failed"
				}
			}
		],
		"tag": "kv-all-attributes",
		"field": "log.original",
		"target_field": "labels",
		"ignore_missing": true,
		"field_split": "&",
		"value_split": ":",
		"include_keys": [
			"env",
			"region"
		],
		"exclude_keys": [
			"debug"
		],
		"prefix": "kv_",
		"trim_key": "_",
		"trim_value": "|",
		"strip_brackets": true
	}
}`

const expectedJSONKVDefaults = `{
	"kv": {
		"field": "event.original",
		"ignore_failure": false,
		"ignore_missing": false,
		"field_split": "&",
		"value_split": "=",
		"strip_brackets": false
	}
}`

const expectedJSONKVUpdatedValues = `{
	"kv": {
		"field": "labels",
		"target_field": "parsed_labels",
		"ignore_failure": false,
		"ignore_missing": false,
		"field_split": ";",
		"value_split": "=>",
		"include_keys": [
			"service",
			"zone"
		],
		"exclude_keys": [
      "debug",
      "temp"
    ],
		"prefix": "meta_",
		"trim_key": "-",
		"trim_value": "~",
		"strip_brackets": false
	}
}`
