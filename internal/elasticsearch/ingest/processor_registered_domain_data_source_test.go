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

func TestAccDataSourceIngestProcessorRegisteredDomain(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "field", "fqdn"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "target_field", "url"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "json", expectedJSONRegisteredDomain),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "field", "fqdn"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "target_field", "url_parts"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "description", "Extract registered domain parts"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "if", "ctx.fqdn != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.0", `{"set":{"field":"error.message","value":"registered domain failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "tag", "registered-domain"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "json", expectedJSONRegisteredDomainAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "field", "dns.question.name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "target_field", "domain_parts"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "description", "Extract domain details from DNS question"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "if", "ctx.dns?.question?.name != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.0", `{"set":{"field":"error.message","value":"registered domain lookup failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "tag", "registered-domain-update"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "json", expectedJSONRegisteredDomainUpdated),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("target_field_omitted"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "field", "host.name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "json", expectedJSONRegisteredDomainTargetFieldOmitted),
				),
			},
		},
	})
}

const expectedJSONRegisteredDomain = `{
	"registered_domain": {
		"field": "fqdn",
		"ignore_failure": false,
		"ignore_missing": false,
		"target_field": "url"
	}
}`

const expectedJSONRegisteredDomainAllAttributes = `{
	"registered_domain": {
		"description": "Extract registered domain parts",
		"field": "fqdn",
		"if": "ctx.fqdn != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "registered domain failed"
				}
			}
		],
		"tag": "registered-domain",
		"target_field": "url_parts"
	}
}`

const expectedJSONRegisteredDomainUpdated = `{
	"registered_domain": {
		"description": "Extract domain details from DNS question",
		"field": "dns.question.name",
		"if": "ctx.dns?.question?.name != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "registered domain lookup failed"
				}
			}
		],
		"tag": "registered-domain-update",
		"target_field": "domain_parts"
	}
}`

const expectedJSONRegisteredDomainTargetFieldOmitted = `{
	"registered_domain": {
		"field": "host.name",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`
