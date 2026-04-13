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

func TestAccDataSourceIngestProcessorInference(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_inference.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "model_id", "my_endpoint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.input_field", "foo"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.output_field", "bar"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "field_map.%"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "target_field"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_inference.test", "json", expectedJSONInference),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_inference.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "model_id", "my_endpoint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.input_field", "foo"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.output_field", "bar"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "field_map.%", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "field_map.content", "text_field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "target_field", "ml.inference"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "description", "Run inference on foo"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "if", "ctx.lang == 'en'"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_inference.test", "on_failure.0", `{"set":{"field":"error.message","value":"inference failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "tag", "inference-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_inference.test", "json", expectedJSONInferenceAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("changed_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_inference.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "model_id", "updated_endpoint"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.input_field", "body.content"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "input_output.0.output_field", ""),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "field_map.%"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "target_field", "ml.updated"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "description", "Run inference on body.content"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "if", "ctx.body?.content != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "on_failure.#"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_inference.test", "tag", "updated-inference-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_inference.test", "json", expectedJSONInferenceChangedValues),
				),
			},
		},
	})
}

const expectedJSONInference = `{
  "inference": {
    "model_id": "my_endpoint",
    "input_output": {
      "input_field": "foo",
      "output_field": "bar"
    },
    "ignore_failure": false
  }
}`

const expectedJSONInferenceAllAttributes = `{
  "inference": {
    "model_id": "my_endpoint",
    "input_output": {
      "input_field": "foo",
      "output_field": "bar"
    },
    "field_map": {
      "content": "text_field"
    },
    "target_field": "ml.inference",
    "description": "Run inference on foo",
    "if": "ctx.lang == 'en'",
    "ignore_failure": true,
    "on_failure": [
      {
        "set": {
          "field": "error.message",
          "value": "inference failed"
        }
      }
    ],
    "tag": "inference-tag"
  }
}`

const expectedJSONInferenceChangedValues = `{
  "inference": {
    "model_id": "updated_endpoint",
    "input_output": {
      "input_field": "body.content"
    },
    "target_field": "ml.updated",
    "description": "Run inference on body.content",
    "if": "ctx.body?.content != null",
    "ignore_failure": false,
    "tag": "updated-inference-tag"
  }
}`
