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

func TestAccDataSourceIngestProcessorURIParts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "field", "input_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "target_field", "url"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "keep_original", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "remove_if_successful", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "json", expectedJSONURIParts),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "field", "request.uri"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "target_field", "parsed_url"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "keep_original", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "remove_if_successful", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "description", "Parse URI parts from request"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "if", "ctx.request?.uri != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "tag", "uri-parts-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "on_failure.0", `{"set":{"field":"error.message","value":"uri parts failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uri_parts.test", "json", expectedJSONURIPartsAllAttributes),
				),
			},
		},
	})
}

const expectedJSONURIParts = `{
	"uri_parts": {
		"field": "input_field",
		"ignore_failure": false,
		"keep_original": true,
		"remove_if_successful": false,
		"target_field": "url"
	}
}`

const expectedJSONURIPartsAllAttributes = `{
	"uri_parts": {
		"description": "Parse URI parts from request",
		"field": "request.uri",
		"if": "ctx.request?.uri != null",
		"ignore_failure": true,
		"keep_original": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "uri parts failed"
				}
			}
		],
		"remove_if_successful": true,
		"tag": "uri-parts-tag",
		"target_field": "parsed_url"
	}
}`
