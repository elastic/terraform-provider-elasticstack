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

func TestAccDataSourceIngestProcessorDateIndexName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "field", "date1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_rounding", "M"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "description", "monthly date-time index naming"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "index_name_prefix", "my-index-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "timezone", "UTC"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "locale", "ENGLISH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "index_name_format", "yyyy-MM-dd"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "json", expectedJSONDateIndexName),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "field", "event_date"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_rounding", "d"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_formats.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_formats.0", "ISO8601"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_formats.1", "UNIX"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "index_name_prefix", "logs-"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "locale", "FRENCH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "index_name_format", "yyyy-MM"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "description", "route documents by event date"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "if", "ctx.event_date != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "tag", "date-index-name-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "json", expectedJSONDateIndexNameAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "on_failure.0", `{"set":{"field":"error.message","value":"date index routing failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "json", expectedJSONDateIndexNameOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "field", "timestamp_raw"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "date_rounding", "d"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "timezone", "UTC"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "locale", "ENGLISH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "index_name_format", "yyyy-MM-dd"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date_index_name.test", "json", expectedJSONDateIndexNameDefaults),
				),
			},
		},
	})
}

const expectedJSONDateIndexName = `{
  "date_index_name": {
    "date_rounding": "M",
    "description": "monthly date-time index naming",
    "field": "date1",
    "ignore_failure": false,
    "index_name_format": "yyyy-MM-dd",
    "index_name_prefix": "my-index-",
    "locale": "ENGLISH",
    "timezone": "UTC"
  }
}
`

const expectedJSONDateIndexNameAllAttributes = `{
  "date_index_name": {
    "date_formats": ["ISO8601", "UNIX"],
    "date_rounding": "d",
    "description": "route documents by event date",
    "field": "event_date",
    "if": "ctx.event_date != null",
    "ignore_failure": true,
    "index_name_format": "yyyy-MM",
    "index_name_prefix": "logs-",
    "locale": "FRENCH",
    "tag": "date-index-name-tag",
    "timezone": "America/New_York"
  }
}
`

const expectedJSONDateIndexNameOnFailure = `{
  "date_index_name": {
    "date_rounding": "M",
    "field": "date1",
    "ignore_failure": false,
    "index_name_format": "yyyy-MM-dd",
    "locale": "ENGLISH",
    "on_failure": [
      {
        "set": {
          "field": "error.message",
          "value": "date index routing failed"
        }
      }
    ],
    "timezone": "UTC"
  }
}
`

const expectedJSONDateIndexNameDefaults = `{
  "date_index_name": {
    "date_rounding": "d",
    "field": "timestamp_raw",
    "ignore_failure": false,
    "index_name_format": "yyyy-MM-dd",
    "locale": "ENGLISH",
    "timezone": "UTC"
  }
}
`
