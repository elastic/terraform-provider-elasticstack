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

func TestAccDataSourceIngestProcessorCSV(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "field", "my_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.0", "field1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.1", "field2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "separator", ","),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "quote", `"`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "trim", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSV),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ignore_missing"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "field", "csv_payload"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.#", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.0", "first_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.1", "role"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "target_fields.2", "city"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "ignore_missing", "true"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSVIgnoreMissing),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("parsing_options"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "separator", ";"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "quote", "'"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "trim", "true"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSVParsingOptions),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "empty_value", "N/A"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSVEmptyValue),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("metadata"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "description", "Parse CSV when payload is present"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "if", "ctx.csv_payload != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "tag", "csv-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSVMetadata),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_csv.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_csv.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "on_failure.0", `{"set":{"field":"error.message","value":"csv failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_csv.test", "json", expectedJSONCSVOnFailure),
				),
			},
		},
	})
}

const expectedJSONCSV = `{
	"csv": {
		"field": "my_field",
		"target_fields": ["field1", "field2"],
		"separator": ",",
		"trim": false,
		"quote": "\"",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONCSVIgnoreMissing = `{
	"csv": {
		"field": "csv_payload",
		"target_fields": ["first_name", "role", "city"],
		"separator": ",",
		"trim": false,
		"quote": "\"",
		"ignore_failure": false,
		"ignore_missing": true
	}
}`

const expectedJSONCSVParsingOptions = `{
	"csv": {
		"field": "csv_payload",
		"target_fields": ["first_name", "role", "city"],
		"separator": ";",
		"trim": true,
		"quote": "'",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONCSVEmptyValue = `{
	"csv": {
		"field": "csv_payload",
		"target_fields": ["first_name", "role"],
		"separator": ",",
		"trim": false,
		"quote": "\"",
		"ignore_failure": false,
		"ignore_missing": false,
		"empty_value": "N/A"
	}
}`

const expectedJSONCSVMetadata = `{
	"csv": {
		"description": "Parse CSV when payload is present",
		"if": "ctx.csv_payload != null",
		"tag": "csv-tag",
		"field": "csv_payload",
		"target_fields": ["first_name", "role"],
		"separator": ",",
		"trim": false,
		"quote": "\"",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONCSVOnFailure = `{
	"csv": {
		"field": "csv_payload",
		"target_fields": ["first_name", "role"],
		"separator": ",",
		"trim": false,
		"quote": "\"",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "csv failed"
				}
			}
		],
		"ignore_missing": false
	}
}`
