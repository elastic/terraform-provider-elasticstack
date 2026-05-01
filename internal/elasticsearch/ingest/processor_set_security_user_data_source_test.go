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

func TestAccDataSourceIngestProcessorSetSecurityUser(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "field", "user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "json", expectedJSONSetSecurityUser),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "field", "actor"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "description", "set security user metadata"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "if", "ctx.user != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "tag", "set-security-user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "properties.#", "3"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "properties.*", "username"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "properties.*", "roles"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "properties.*", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "on_failure.0", `{"set":{"field":"error.message","value":"fallback"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "json", expectedJSONSetSecurityUserAllAttributes),
				),
			},
		},
	})
}

const expectedJSONSetSecurityUser = `{
  "set_security_user": {
    "ignore_failure": false,
    "field": "user"
  }
}`

const expectedJSONSetSecurityUserAllAttributes = `{
	"set_security_user": {
		"description": "set security user metadata",
		"field": "actor",
		"if": "ctx.user != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "fallback"
				}
			}
		],
		"properties": [
      "email",
      "roles",
      "username"
    ],
		"tag": "set-security-user"
	}
}`
