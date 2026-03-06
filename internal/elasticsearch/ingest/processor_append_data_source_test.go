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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "field", "tags"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_append.test", "json", expectedJSONAppend),
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
