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

func TestAccDataSourceIngestProcessorBytes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "field", "file.size"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "json", expectedJSONBytes),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "field", "document.size"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "target_field", "document.size_bytes"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "description", "Convert document size to bytes"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "if", "ctx.document?.size != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "tag", "bytes-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_bytes.test", "json", expectedJSONBytesAllAttributes),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorBytesOnFailure(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_bytes.test_on_failure", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test_on_failure", "field", "file.size"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_bytes.test_on_failure", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_bytes.test_on_failure", "on_failure.0", `{"set":{"field":"error.message","value":"{{ _ingest.on_failure_message }}"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_bytes.test_on_failure", "json", expectedJSONBytesOnFailure),
				),
			},
		},
	})
}

const expectedJSONBytes = `{
	"bytes": {
		"field": "file.size", 
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONBytesAllAttributes = `{
	"bytes": {
		"description": "Convert document size to bytes",
		"if": "ctx.document?.size != null",
		"ignore_failure": true,
		"tag": "bytes-tag",
		"field": "document.size",
		"target_field": "document.size_bytes",
		"ignore_missing": true
	}
}`

const expectedJSONBytesOnFailure = `{
	"bytes": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "{{ _ingest.on_failure_message }}"
				}
			}
		],
		"field": "file.size",
		"ignore_missing": false
	}
}`
