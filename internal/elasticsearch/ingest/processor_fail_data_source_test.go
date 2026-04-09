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

func TestAccDataSourceIngestProcessorFail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fail.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "message", "The production tag is not present, found tags: {{{tags}}}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "if", "ctx.tags.contains('production') != true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJSONFail),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fail.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "message", "Document is missing a required deployment identifier"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "description", "Fail when deployment metadata is missing"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "if", "ctx.deployment_id == null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "tag", "fail-missing-deployment-id"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJSONFailAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fail.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "on_failure.0", `{"set":{"field":"error.message","value":"fail processor triggered"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJSONFailOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_fail.test", "id"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "tag"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "on_failure.#"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "message", "Reject documents without an event category"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_fail.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_fail.test", "json", expectedJSONFailDefaults),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorFailInvalidOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_on_failure"),
				ExpectError:              regexp.MustCompile(`"on_failure\.0" contains an invalid JSON`),
			},
		},
	})
}

const expectedJSONFail = `{
  "fail": {
		"message": "The production tag is not present, found tags: {{{tags}}}",
		"ignore_failure": false,
		"if" : "ctx.tags.contains('production') != true"
	}
}
`

const expectedJSONFailAllAttributes = `{
  "fail": {
		"description": "Fail when deployment metadata is missing",
		"if": "ctx.deployment_id == null",
		"ignore_failure": true,
		"tag": "fail-missing-deployment-id",
		"message": "Document is missing a required deployment identifier"
  }
}
`

const expectedJSONFailOnFailure = `{
  "fail": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "fail processor triggered"
				}
			}
		],
		"message": "Reject documents without a service name"
  }
}
`

const expectedJSONFailDefaults = `{
  "fail": {
		"message": "Reject documents without an event category",
		"ignore_failure": false
  }
}
`
