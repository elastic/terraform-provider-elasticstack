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

func TestAccDataSourceIngestProcessorAppend(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorAppend,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_append.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "field", "tags"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_append.test", "json", expectedJSONAppend),
				),
			},
			{
				Config: testAccDataSourceIngestProcessorAppendAllAttributes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_append.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "media_type", "application/json"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "if", "ctx.error != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "tag", "append-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_append.test", "on_failure.0", `{"set":{"field":"error.message","value":"append failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_append.test", "json", expectedJSONAppendAllAttributes),
				),
			},
		},
	})
}

const expectedJSONAppend = `{
	"append": {
		"field": "tags", 
		"value": ["production", "{{{app}}}", "{{{owner}}}"], 
		"allow_duplicates": true,
		"description": "Append tags to the doc", 
		"ignore_failure": false
	}
}`

const expectedJSONAppendAllAttributes = `{
	"append": {
		"field": "tags",
		"value": ["404"],
		"allow_duplicates": false,
		"media_type": "application/json",
		"description": "Append a numeric-like error code to tags",
		"if": "ctx.error != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "append failed"
				}
			}
		],
		"tag": "append-tag"
	}
}`

const testAccDataSourceIngestProcessorAppend = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  description      = "Append tags to the doc"
  field            = "tags"
  value            = ["production", "{{{app}}}", "{{{owner}}}"]
  allow_duplicates = true
}
`

const testAccDataSourceIngestProcessorAppendAllAttributes = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  description      = "Append a numeric-like error code to tags"
  field            = "tags"
  value            = ["404"]
  allow_duplicates = false
  media_type       = "application/json"
  if               = "ctx.error != null"
  ignore_failure   = true
  tag              = "append-tag"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "append failed"
      }
    })
  ]
}
`
