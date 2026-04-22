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

func TestAccDataSourceIngestProcessorCircle(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_circle.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "field", "circle"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "error_distance", "28.1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "shape_type", "geo_shape"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_circle.test", "json", expectedJSONCircle),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_circle.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "field", "location"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "target_field", "location_shape"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "error_distance", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "shape_type", "geo_shape"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "description", "Convert circle to polygon"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "if", "ctx.location != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_circle.test", "on_failure.0", `{"set":{"field":"error.message","value":"circle failed"}}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "tag", "circle-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_circle.test", "json", expectedJSONCircleAllAttributes),
				),
			},
		},
	})
}

func TestAccDataSourceIngestProcessorCircleShapeType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_circle.test_shape", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test_shape", "field", "circle"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test_shape", "error_distance", "10"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test_shape", "shape_type", "shape"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_circle.test_shape", "json", expectedJSONCircleShapeType),
				),
			},
		},
	})
}

const expectedJSONCircle = `{
	"circle": {
		"field": "circle",
		"error_distance": 28.1,
		"shape_type": "geo_shape",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const expectedJSONCircleAllAttributes = `{
	"circle": {
		"field": "location",
		"target_field": "location_shape",
		"ignore_missing": true,
		"error_distance": 5,
		"shape_type": "geo_shape",
		"description": "Convert circle to polygon",
		"if": "ctx.location != null",
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "circle failed"
				}
			}
		],
		"tag": "circle-tag"
	}
}`

const expectedJSONCircleShapeType = `{
	"circle": {
		"field": "circle",
		"error_distance": 10,
		"shape_type": "shape",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`
