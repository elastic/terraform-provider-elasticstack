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

func TestAccDataSourceIngestProcessorDate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "field", "initial_date"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "target_field", "timestamp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.0", "dd/MM/yyyy HH:mm:ss"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "timezone", "Europe/Amsterdam"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "locale", "ENGLISH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "output_format", "yyyy-MM-dd'T'HH:mm:ss.SSSXXX"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date.test", "json", expectedJSONDate),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("all_attributes"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "field", "event_date"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "target_field", "parsed_timestamp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.#", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.0", "ISO8601"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.1", "UNIX"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.2", "dd/MM/yyyy"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "locale", "FRENCH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "output_format", "yyyy-MM-dd"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "description", "Parse date from event_date field"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "if", "ctx.event_date != null"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "ignore_failure", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "tag", "date-tag"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date.test", "json", expectedJSONDateAllAttributes),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("on_failure"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "on_failure.#", "1"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date.test", "on_failure.0", `{"set":{"field":"error.message","value":"date parse failed"}}`),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date.test", "json", expectedJSONDateOnFailure),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_ingest_processor_date.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "field", "timestamp_raw"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "formats.0", "ISO8601"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "target_field", "@timestamp"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "timezone", "UTC"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "locale", "ENGLISH"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "output_format", "yyyy-MM-dd'T'HH:mm:ss.SSSXXX"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_date.test", "ignore_failure", "false"),
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_date.test", "json", expectedJSONDateDefaults),
				),
			},
		},
	})
}

const expectedJSONDate = `{
  "date": {
    "field": "initial_date",
    "formats": [
      "dd/MM/yyyy HH:mm:ss"
    ],
    "ignore_failure": false,
    "locale": "ENGLISH",
    "output_format": "yyyy-MM-dd'T'HH:mm:ss.SSSXXX",
    "target_field": "timestamp",
    "timezone": "Europe/Amsterdam"
  }
}
`

const expectedJSONDateAllAttributes = `{
  "date": {
    "description": "Parse date from event_date field",
    "field": "event_date",
    "formats": ["ISO8601", "UNIX", "dd/MM/yyyy"],
    "if": "ctx.event_date != null",
    "ignore_failure": true,
    "locale": "FRENCH",
    "output_format": "yyyy-MM-dd",
    "tag": "date-tag",
    "target_field": "parsed_timestamp",
    "timezone": "America/New_York"
  }
}
`

const expectedJSONDateOnFailure = `{
  "date": {
    "field": "initial_date",
    "formats": ["dd/MM/yyyy HH:mm:ss"],
    "ignore_failure": false,
    "locale": "ENGLISH",
    "on_failure": [
      {
        "set": {
          "field": "error.message",
          "value": "date parse failed"
        }
      }
    ],
    "output_format": "yyyy-MM-dd'T'HH:mm:ss.SSSXXX",
    "timezone": "UTC"
  }
}
`

const expectedJSONDateDefaults = `{
  "date": {
    "field": "timestamp_raw",
    "formats": ["ISO8601"],
    "ignore_failure": false,
    "locale": "ENGLISH",
    "output_format": "yyyy-MM-dd'T'HH:mm:ss.SSSXXX",
    "timezone": "UTC"
  }
}
`
