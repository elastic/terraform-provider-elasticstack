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

func TestAccDataSourceIngestProcessorPipeline(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "name", "pipeline_a"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "json", expectedJSONPipeline),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "name", "pipeline_with_metadata"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "description", "Route documents through the metadata pipeline"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "if", "ctx.service?.name != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "tag", "pipeline-metadata-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "json", expectedJSONPipelineAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "name", "pipeline_with_failure_handler"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "on_failure.0", `{"set":{"field":"error.message","value":"pipeline processor failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "json", expectedJSONPipelineOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated_values"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "name", "pipeline_b"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "json", expectedJSONPipelineUpdatedValues),
				),
			},
		},
	})
}

const expectedJSONPipeline = `{
	"pipeline": {
		"name": "pipeline_a",
		"ignore_failure": false
	}
}`

const expectedJSONPipelineAllAttributes = `{
	"pipeline": {
		"description": "Route documents through the metadata pipeline",
		"if": "ctx.service?.name != null",
		"ignore_failure": true,
		"tag": "pipeline-metadata-tag",
		"name": "pipeline_with_metadata"
	}
}`

const expectedJSONPipelineOnFailure = `{
	"pipeline": {
		"ignore_failure": false,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "pipeline processor failed"
				}
			}
		],
		"name": "pipeline_with_failure_handler"
	}
}`

const expectedJSONPipelineUpdatedValues = `{
	"pipeline": {
		"name": "pipeline_b",
		"ignore_failure": false
	}
}`
