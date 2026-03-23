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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_convert.test", "field", "id"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_convert.test", "json", expectedJSONConvert),
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
