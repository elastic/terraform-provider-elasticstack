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

func TestAccDataSourceIngestProcessorUppercase(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "field", "foo"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "json", expectedJSONUppercase),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "field", "source_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "target_field", "uppercased_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "description", "Normalize message to uppercase"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "if", "ctx.source_field != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "on_failure.0", `{"set":{"field":"error.message","value":"uppercase failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "tag", "uppercase-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "json", expectedJSONUppercaseAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "field", "updated_source_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "target_field", "updated_uppercased_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_uppercase.test", "json", expectedJSONUppercaseUpdatedValues),
				),
			},
		},
	})
}

const expectedJSONUppercase = `{
	"uppercase": {
		"field": "foo",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONUppercaseAllAttributes = `{
	"uppercase": {
		"description": "Normalize message to uppercase",
		"field": "source_field",
		"if": "ctx.source_field != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "uppercase failed"
				}
			}
		],
		"tag": "uppercase-tag",
		"target_field": "uppercased_field"
	}
}`

const expectedJSONUppercaseUpdatedValues = `{
	"uppercase": {
		"field": "updated_source_field",
		"ignore_failure": false,
		"ignore_missing": false,
		"target_field": "updated_uppercased_field"
	}
}`
