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

func TestAccDataSourceIngestProcessorDissect(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "field", "message"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "json", expectedJSONDissect),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "append_separator", "|"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "description", "Dissect log line"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "if", "ctx.message != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "tag", "dissect-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "json", expectedJSONDissectAllAttributes),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorDissectOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure", "field", "message"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure", "on_failure.0", `{"set":{"field":"error.message","value":"{{ _ingest.on_failure_message }}"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure", "json", expectedJSONDissectOnFailure),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorDissectOnFailureMulti(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure_multi", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure_multi", "on_failure.#", "2"),
					CheckResourceJSON(
						"data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure_multi",
						"on_failure.0",
						`{"set":{"field":"error.message","value":"{{ _ingest.on_failure_message }}"}}`,
					),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure_multi", "on_failure.1", `{"set":{"field":"event.kind","value":"pipeline_error"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_dissect.test_on_failure_multi", "json", expectedJSONDissectOnFailureMulti),
				),
			},
		},
	})
}

const expectedJSONDissect = `{
  "dissect": {
		"append_separator": "",
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern": "%{clientip} %{ident} %{auth} [%{@timestamp}] \"%{verb} %{request} HTTP/%{httpversion}\" %{status} %{size}"
	}
}
`

const expectedJSONDissectAllAttributes = `{
	"dissect": {
		"append_separator": "|",
		"description": "Dissect log line",
		"field": "message",
		"if": "ctx.message != null",
		"ignore_failure": true,
		"ignore_missing": true,
		"pattern": "%{clientip} %{ident} %{auth} [%{@timestamp}] \"%{verb} %{request} HTTP/%{httpversion}\" %{status} %{size}",
		"tag": "dissect-tag"
	}
}`

const expectedJSONDissectOnFailure = `{
	"dissect": {
		"append_separator": "",
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "{{ _ingest.on_failure_message }}"
				}
			}
		],
		"pattern": "%{clientip} %{ident} %{auth}"
	}
}`

const expectedJSONDissectOnFailureMulti = `{
	"dissect": {
		"append_separator": "",
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "{{ _ingest.on_failure_message }}"
				}
			},
			{
				"set": {
					"field": "event.kind",
					"value": "pipeline_error"
				}
			}
		],
		"pattern": "%{clientip} %{ident} %{auth}"
	}
}`
