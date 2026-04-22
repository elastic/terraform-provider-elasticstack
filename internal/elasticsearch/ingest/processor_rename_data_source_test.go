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

func TestAccDataSourceIngestProcessorRename(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_rename.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "field", "provider"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "target_field", "cloud.provider"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_rename.test", "json", expectedJSONRename),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_rename.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "field", "provider"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "target_field", "cloud.provider"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "description", "Rename provider field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "if", "ctx.provider != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_rename.test", "on_failure.0", `{"set":{"field":"error.message","value":"rename failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "tag", "rename-provider"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_rename.test", "json", expectedJSONRenameAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_rename.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "field", "service.name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "target_field", "service.type"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_rename.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_rename.test", "json", expectedJSONRenameUpdated),
				),
			},
		},
	})
}

const expectedJSONRename = `{
	"rename": {
		"field": "provider",
		"target_field": "cloud.provider",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONRenameAllAttributes = `{
	"rename": {
		"description": "Rename provider field",
		"field": "provider",
		"if": "ctx.provider != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "rename failed"
				}
			}
		],
		"tag": "rename-provider",
		"target_field": "cloud.provider"
	}
}`

const expectedJSONRenameUpdated = `{
	"rename": {
		"field": "service.name",
		"target_field": "service.type",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`
