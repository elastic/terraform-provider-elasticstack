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

func TestAccDataSourceIngestProcessorReroute(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "destination", "logs-generic-default"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "ignore_failure", "false"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "dataset"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "namespace"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "description"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "if"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "on_failure.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "json", expectedJSONReroute),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_optional_fields"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "destination", "logs-app-default"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "dataset", "application"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "namespace", "production"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "description", "Route application logs"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "if", "ctx.service?.name != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "ignore_failure", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "tag", "reroute-app-logs"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "json", expectedJSONRerouteAllOptionalFields),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "destination", "logs-fallback-default"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "on_failure.0", `{"set":{"field":"error.message","value":"reroute failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_reroute.test", "json", expectedJSONRerouteOnFailure),
				),
			},
		},
	})
}

const expectedJSONReroute = `{
	"reroute": {
		"ignore_failure": false,
		"destination": "logs-generic-default"
	}
}`

const expectedJSONRerouteAllOptionalFields = `{
	"reroute": {
		"description": "Route application logs",
		"if": "ctx.service?.name != null",
		"ignore_failure": false,
		"tag": "reroute-app-logs",
		"destination": "logs-app-default",
		"dataset": "application",
		"namespace": "production"
	}
}`

const expectedJSONRerouteOnFailure = `{
	"reroute": {
		"ignore_failure": true,
		"on_failure": [
			{
				"set": {
					"field": "error.message",
					"value": "reroute failed"
				}
			}
		],
		"destination": "logs-fallback-default"
	}
}`
