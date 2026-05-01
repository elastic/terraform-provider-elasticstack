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

func TestAccDataSourceIngestProcessorGeoip(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "field", "ip"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "json", expectedJSONGeoip),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "field", "ip"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "target_field", "geoip"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "ignore_missing", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "description", "geoip lookup"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "if", "ctx.ip != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "tag", "geoip-tag"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "json", expectedJSONGeoipAllAttributes),
				),
			},
		},
	})
}

const expectedJSONGeoip = `{
  "geoip": {
    "ignore_failure": false,
    "field": "ip",
    "target_field": "geoip",
    "ignore_missing": false,
    "first_only": true
  }
}
`

const expectedJSONGeoipAllAttributes = `{
  "geoip": {
    "description": "geoip lookup",
    "if": "ctx.ip != null",
    "ignore_failure": true,
    "on_failure": [
      {
        "set": {
          "field": "error.message",
          "value": "geoip failed"
        }
      }
    ],
    "tag": "geoip-tag",
    "field": "ip",
    "target_field": "geoip",
    "ignore_missing": true,
    "first_only": true
  }
}
`
