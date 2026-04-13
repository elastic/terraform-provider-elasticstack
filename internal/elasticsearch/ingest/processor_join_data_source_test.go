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

func TestAccDataSourceIngestProcessorJoin(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_join.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "field", "joined_array_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "separator", "-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_join.test", "json", expectedJSONJoin),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_join.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "field", "joined_array_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "separator", "::"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "target_field", "joined_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "description", "Join array values into a single field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "if", "ctx.tags != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_join.test", "on_failure.0", `{"set":{"field":"error.message","value":"join failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "tag", "join-tags"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_join.test", "json", expectedJSONJoinAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_join.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "field", "updated_array_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "separator", "|"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_join.test", "json", expectedJSONJoinUpdated),
				),
			},
		},
	})
}

const expectedJSONJoin = `{
	"join": {
		"field": "joined_array_field",
		"ignore_failure": false,
		"separator": "-"
	}
}`

const expectedJSONJoinAllAttributes = `{
	"join": {
		"description": "Join array values into a single field",
		"field": "joined_array_field",
		"if": "ctx.tags != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "join failed"
				}
			}
		],
		"separator": "::",
		"tag": "join-tags",
		"target_field": "joined_field"
	}
}`

const expectedJSONJoinUpdated = `{
	"join": {
		"field": "updated_array_field",
		"ignore_failure": false,
		"separator": "|"
	}
}`
