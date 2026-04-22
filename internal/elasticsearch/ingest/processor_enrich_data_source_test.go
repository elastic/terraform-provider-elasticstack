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

func TestAccDataSourceIngestProcessorEnrich(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "target_field", "user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "policy_name", "users-policy"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "json", expectedJSONEnrich),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "override", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "max_matches", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "shape_relation", "INTERSECTS"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "description", "Enrich user details from a policy"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "if", "ctx.email != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "tag", "enrich-users"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "json", expectedJSONEnrichAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "on_failure.#", "2"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "on_failure.0", `{"set":{"field":"error.message","value":"enrich failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "on_failure.1", `{"set":{"field":"error.type","value":"enrich"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "json", expectedJSONEnrichOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "id"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "shape_relation"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "tag"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "on_failure.#"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "override", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "max_matches", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_enrich.test", "json", expectedJSONEnrichDefaults),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorEnrichInvalidOnFailure(t *testing.T) {
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

const expectedJSONEnrich = `{
	"enrich": {
		"ignore_failure": false,
		"field": "email",
		"target_field": "user",
		"ignore_missing": false,
		"policy_name": "users-policy",
		"override": true,
		"max_matches": 1
	}
}`

const expectedJSONEnrichAllAttributes = `{
	"enrich": {
		"description": "Enrich user details from a policy",
		"if": "ctx.email != null",
		"ignore_failure": true,
		"tag": "enrich-users",
		"field": "email",
		"target_field": "user.profile",
		"ignore_missing": true,
		"policy_name": "users-policy",
		"override": false,
		"max_matches": 3,
		"shape_relation": "INTERSECTS"
	}
}`

const expectedJSONEnrichOnFailure = `{
	"enrich": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "enrich failed"
				}
			},
			{
				"set": {
					"field": "error.type",
					"value": "enrich"
				}
			}
		],
		"field": "email",
		"target_field": "user",
		"ignore_missing": false,
		"policy_name": "users-policy",
		"override": true,
		"max_matches": 1
	}
}`

const expectedJSONEnrichDefaults = `{
	"enrich": {
		"ignore_failure": false,
		"field": "email",
		"target_field": "user",
		"ignore_missing": false,
		"policy_name": "users-policy",
		"override": true,
		"max_matches": 1
	}
}`
