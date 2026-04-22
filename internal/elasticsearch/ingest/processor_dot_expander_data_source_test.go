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

func TestAccDataSourceIngestProcessorDotExpander(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "field", "foo.bar"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "override", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJSONDotExpander),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_options"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "field", "foo.bar"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "path", "nested"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "override", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "description", "Expand dot fields"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "if", "ctx.foo != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "tag", "dot-expander-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJSONDotExpanderAllOptions),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorDotExpanderOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure_single"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "on_failure.0", `{"set":{"field":"error.message","value":"{{ _ingest.on_failure_message }}"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJSONDotExpanderOnFailureSingle),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure_multi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "on_failure.#", "2"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "on_failure.0", `{"set":{"field":"error.message","value":"{{ _ingest.on_failure_message }}"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "on_failure.1", `{"set":{"field":"error.type","value":"dot_expander"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJSONDotExpanderOnFailureMulti),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorDotExpanderWildcard(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "field", "*"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dot_expander.test", "json", expectedJSONDotExpanderWildcard),
				),
			},
		},
	})
}

const expectedJSONDotExpander = `{
  "dot_expander": {
		"field": "foo.bar",
		"ignore_failure": false,
		"override": false
	}
}
`

const expectedJSONDotExpanderAllOptions = `{
	"dot_expander": {
		"description": "Expand dot fields",
		"if": "ctx.foo != null",
		"ignore_failure": true,
		"tag": "dot-expander-tag",
		"field": "foo.bar",
		"path": "nested",
		"override": true
	}
}`

const expectedJSONDotExpanderOnFailureSingle = `{
	"dot_expander": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "{{ _ingest.on_failure_message }}"
				}
			}
		],
		"field": "foo.bar",
		"override": false
	}
}`

const expectedJSONDotExpanderOnFailureMulti = `{
	"dot_expander": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "{{ _ingest.on_failure_message }}"
				}
			},
			{
				"set": {
					"field": "error.type",
					"value": "dot_expander"
				}
			}
		],
		"field": "foo.bar",
		"override": false
	}
}`

const expectedJSONDotExpanderWildcard = `{
	"dot_expander": {
		"field": "*",
		"ignore_failure": false,
		"override": false
	}
}`
