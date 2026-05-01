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

func TestAccDataSourceIngestProcessorUserAgent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "field", "agent"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "ignore_missing", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "json", expectedJSONUserAgent),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "field", "http.request.headers.user-agent"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "target_field", "user_agent_details"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "regex_file", "custom-regexes.yml"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "extract_device_type", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "description", "parse user agent"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "if", "ctx.agent != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "tag", "ua-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "on_failure.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "properties.#", "3"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "properties.*", "name"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "properties.*", "os"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "properties.*", "device"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "json", expectedJSONUserAgentAllAttributes),
				),
			},
		},
	})
}

const expectedJSONUserAgent = `{
  "user_agent": {
    "ignore_failure": false,
    "field": "agent",
    "ignore_missing": false
  }
}`

const expectedJSONUserAgentAllAttributes = `{
  "user_agent": {
    "description": "parse user agent",
    "if": "ctx.agent != null",
    "ignore_failure": true,
    "on_failure": [
      {
        "set": {
          "field": "error.message",
          "value": "ua failed"
        }
      }
    ],
    "tag": "ua-tag",
    "field": "http.request.headers.user-agent",
    "target_field": "user_agent_details",
    "ignore_missing": true,
    "regex_file": "custom-regexes.yml",
    "properties": [
      "device",
      "name",
      "os"
    ],
    "extract_device_type": true
  }
}`
