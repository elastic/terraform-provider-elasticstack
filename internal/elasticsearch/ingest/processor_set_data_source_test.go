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

func TestAccDataSourceIngestProcessorSet(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "field", "count"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "value", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "override", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_empty_value", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "media_type", "application/json"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set.test", "json", expectedJSONSet),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("copy_from"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "field", "archived_count"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "copy_from", "count"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "value"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "override", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_empty_value", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "media_type", "application/json"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set.test", "json", expectedJSONSetCopyFrom),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_default_flags"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "value", "plain-text"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "override", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_empty_value", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "media_type", "text/plain"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set.test", "json", expectedJSONSetNonDefaultFlags),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_set.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "field", "event.kind"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "value", "alert"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "description", "Set the event kind when a severity is present"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "if", "ctx.severity != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set.test", "on_failure.0", `{"set":{"field":"error.message","value":"set processor failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set.test", "tag", "set-event-kind"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_set.test", "json", expectedJSONSetAllAttributes),
				),
			},
		},
	})
}

const expectedJSONSet = `{
	"set": {
		"field": "count",
		"ignore_empty_value": false,
		"ignore_failure": false,
		"media_type": "application/json",
		"override": true,
		"value": "1"
	}
}`

const expectedJSONSetCopyFrom = `{
	"set": {
		"copy_from": "count",
		"field": "archived_count",
		"ignore_empty_value": false,
		"ignore_failure": false,
		"media_type": "application/json",
		"override": true
	}
}`

const expectedJSONSetNonDefaultFlags = `{
	"set": {
		"field": "message",
		"ignore_empty_value": true,
		"ignore_failure": false,
		"media_type": "text/plain",
		"override": false,
		"value": "plain-text"
	}
}`

const expectedJSONSetAllAttributes = `{
	"set": {
		"description": "Set the event kind when a severity is present",
		"field": "event.kind",
		"if": "ctx.severity != null",
		"ignore_empty_value": false,
		"ignore_failure": true,
		"media_type": "application/json",
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "set processor failed"
				}
			}
		],
		"override": true,
		"tag": "set-event-kind",
		"value": "alert"
	}
}`
