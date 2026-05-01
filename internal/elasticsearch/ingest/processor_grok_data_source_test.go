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

func TestAccDataSourceIngestProcessorGrok(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_grok.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.0", "%{FAVORITE_DOG:pet}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.1", "%{FAVORITE_CAT:pet}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.FAVORITE_DOG", "beagle"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.FAVORITE_CAT", "burmese"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "trace_match", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJSONGrok),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_grok.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "log.original"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.0", "%{CUSTOMLEVEL:log.level} %{GREEDYDATA:message}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.1", "%{CUSTOMLEVEL:log.level}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.CUSTOMLEVEL", "INFO|WARN|ERROR"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ecs_compatibility", "v1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "trace_match", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "description", "Parse ECS-compatible log lines"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "if", "ctx.log != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "tag", "grok-all-attributes"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJSONGrokAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_grok.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.0", "%{WORD:log.level}: %{GREEDYDATA:message}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "on_failure.0", `{"set":{"field":"error.message","value":"failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJSONGrokOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_grok.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "event.original"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.0", "%{WORD:event.action}"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.%"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ecs_compatibility"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "trace_match", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJSONGrokDefaults),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_pattern_definitions"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_grok.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "patterns.0", "%{WORD:status}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "pattern_definitions.%", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "trace_match", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_missing", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJSONGrokEmptyPatternDefinitions),
				),
			},
		},
	})
}

const expectedJSONGrok = `{
  "grok": {
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern_definitions": {
			"FAVORITE_CAT": "burmese",
			"FAVORITE_DOG": "beagle"
		},
		"patterns": [
			"%{FAVORITE_DOG:pet}",
			"%{FAVORITE_CAT:pet}"
		],
		"trace_match": false
	}
}
`

const expectedJSONGrokAllAttributes = `{
	"grok": {
		"description": "Parse ECS-compatible log lines",
		"ecs_compatibility": "v1",
		"field": "log.original",
		"if": "ctx.log != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"pattern_definitions": {
			"CUSTOMLEVEL": "INFO|WARN|ERROR"
		},
		"patterns": [
			"%{CUSTOMLEVEL:log.level} %{GREEDYDATA:message}",
			"%{CUSTOMLEVEL:log.level}"
		],
		"tag": "grok-all-attributes",
		"trace_match": true
	}
}
`

const expectedJSONGrokOnFailure = `{
	"grok": {
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "failed"
				}
			}
		],
		"patterns": [
			"%{WORD:log.level}: %{GREEDYDATA:message}"
		],
		"trace_match": false
	}
}
`

const expectedJSONGrokDefaults = `{
	"grok": {
		"field": "event.original",
		"ignore_failure": false,
		"ignore_missing": false,
		"patterns": [
			"%{WORD:event.action}"
		],
		"trace_match": false
	}
}
`

const expectedJSONGrokEmptyPatternDefinitions = `{
	"grok": {
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"patterns": [
			"%{WORD:status}"
		],
		"trace_match": false
	}
}
`
