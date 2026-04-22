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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorDrop(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "if", "ctx.network_name == 'Guest'"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_drop.test", "json", expectedJSONDrop),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "tag"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "on_failure.#"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_drop.test", "json", expectedJSONDropMinimal),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_common_fields"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "description", "Drop guest traffic"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "if", "ctx.network_name == 'Guest'"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "tag", "drop-guest-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_drop.test", "json", expectedJSONDropAllCommonFields),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_drop.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_drop.test", "on_failure.0", `{"set":{"field":"error.message","value":"drop failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_drop.test", "json", expectedJSONDropOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("id_determinism"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.first", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_drop.second", "id"),
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_elasticsearch_ingest_processor_drop.first",
						"id",
						"data.elasticstack_elasticsearch_ingest_processor_drop.second",
						"id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorDropInvalidOnFailure(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceIngestProcessorDropInvalidOnFailureConfig,
				ExpectError: regexp.MustCompile(`"on_failure\.0" contains an invalid JSON`),
			},
		},
	})
}

const expectedJSONDrop = `{
	"drop": {
		"if": "ctx.network_name == 'Guest'",
		"ignore_failure": false
	}
}`

const expectedJSONDropMinimal = `{
	"drop": {
		"ignore_failure": false
	}
}`

const expectedJSONDropAllCommonFields = `{
	"drop": {
		"description": "Drop guest traffic",
		"if": "ctx.network_name == 'Guest'",
		"ignore_failure": true,
		"tag": "drop-guest-tag"
	}
}`

const expectedJSONDropOnFailure = `{
	"drop": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "drop failed"
				}
			}
		]
	}
}`

const testAccDataSourceIngestProcessorDropInvalidOnFailureConfig = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "test" {
  on_failure = ["{\"set\":"]
}
`
