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

func TestAccDataSourceIngestProcessorConvert(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorConvert,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_convert.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "field", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "type", "integer"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "description", "converts the content of the id field to an integer"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_convert.test", "json", expectedJSONConvert),
				),
			},
			{
				Config: testAccDataSourceIngestProcessorConvertAllAttributes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_convert.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "field", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "target_field", "converted_id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "type", "integer"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "description", "converts the content of the id field to an integer"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "if", "ctx.id != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "tag", "convert-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_convert.test", "json", expectedJSONConvertAllAttributes),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorConvertOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorConvertOnFailure,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "on_failure.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "ignore_failure", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "ignore_missing", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_convert.test", "on_failure.0", `{"set":{"field":"error.message","value":"convert failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_convert.test", "json", expectedJSONConvertOnFailure),
				),
			},
		},
	})
}

const expectedJSONConvert = `{
	"convert": {
		"description": "converts the content of the id field to an integer",
		"field": "id",
		"type": "integer",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONConvertAllAttributes = `{
	"convert": {
		"description": "converts the content of the id field to an integer",
		"field": "id",
		"target_field": "converted_id",
		"type": "integer",
		"if": "ctx.id != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"tag": "convert-tag"
	}
}`

const expectedJSONConvertOnFailure = `{
	"convert": {
		"field": "id",
		"type": "integer",
		"ignore_failure": false,
		"ignore_missing": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "convert failed"
				}
			}
		]
	}
}`

const testAccDataSourceIngestProcessorConvert = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  description = "converts the content of the id field to an integer"
  field       = "id"
  type        = "integer"
}
`

const testAccDataSourceIngestProcessorConvertAllAttributes = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  description    = "converts the content of the id field to an integer"
  field          = "id"
  target_field   = "converted_id"
  type           = "integer"
  if             = "ctx.id != null"
  ignore_missing = true
  ignore_failure = true
  tag            = "convert-tag"
}
`

const testAccDataSourceIngestProcessorConvertOnFailure = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  field = "id"
  type  = "integer"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "convert failed"
      }
    })
  ]
}
`
